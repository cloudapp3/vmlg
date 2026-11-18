package bgp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
)

type birdDriver struct {
	cfg config.BGPConfig
}

func newBirdDriver(cfg config.BGPConfig) *birdDriver {
	return &birdDriver{cfg: cfg}
}

func (d *birdDriver) Name() string { return "bird" }

func (d *birdDriver) Available() (bool, string) {
	st, err := os.Stat(d.cfg.BirdSocket)
	if err != nil {
		return false, fmt.Sprintf("bird socket %s unavailable", d.cfg.BirdSocket)
	}
	if st.Mode()&os.ModeSocket == 0 {
		return false, fmt.Sprintf("%s is not a unix socket", d.cfg.BirdSocket)
	}
	return true, ""
}

func (d *birdDriver) querySocket(ctx context.Context, cmd string) (string, error) {
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(ctx, "unix", d.cfg.BirdSocket)
	if err != nil {
		return "", fmt.Errorf("connect bird socket: %w", err)
	}
	defer conn.Close()

	deadline := time.Now().Add(8 * time.Second)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}
	_ = conn.SetDeadline(deadline)

	reader := bufio.NewReader(conn)
	// Read greeting banner (e.g. "0001 BIRD 2.15.1 ready.")
	if _, err := reader.ReadString('\n'); err != nil {
		return "", fmt.Errorf("read bird greeting: %w", err)
	}

	// Enter BIRD read-only restricted mode
	if _, err := fmt.Fprint(conn, "restrict\n"); err != nil {
		return "", err
	}
	if _, err := readBirdReply(reader); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(conn, "%s\n", cmd); err != nil {
		return "", err
	}
	return readBirdReply(reader)
}

func readBirdReply(r *bufio.Reader) (string, error) {
	var cleanLines []string
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return strings.Join(cleanLines, "\n"), err
		}
		line = strings.TrimRight(line, "\r\n")
		if len(line) == 0 {
			continue
		}

		// Continuation line starts with space
		if line[0] == ' ' {
			cleanLines = append(cleanLines, line[1:])
			continue
		}

		if len(line) >= 4 && isDigits(line[:4]) {
			code, _ := strconv.Atoi(line[:4])
			sep := byte(' ')
			if len(line) > 4 {
				sep = line[4]
			}
			rest := ""
			if len(line) > 5 {
				rest = line[5:]
			}

			switch {
			case code == 0 || code == 16:
				// 0000 OK or 0016 Access restricted -> terminal success
				return strings.Join(cleanLines, "\n"), nil
			case code >= 8000:
				return strings.Join(cleanLines, "\n"), errors.New(rest)
			default:
				if rest != "" && code != 1007 && code != 1002 {
					cleanLines = append(cleanLines, rest)
				} else if (code == 1007 || code == 1002) && rest != "" {
					cleanLines = append(cleanLines, rest)
				}
				if sep == ' ' && code < 1000 {
					return strings.Join(cleanLines, "\n"), nil
				}
			}
		} else {
			cleanLines = append(cleanLines, line)
		}
	}
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

var (
	reBirdRouteHeader = regexp.MustCompile(`^(\S+)?\s*(?:unicast|blackhole|unreachable|prohibited)?\s*\[([^\s\]]+)\s+([^\]]+)\]\s*(\*)?`)
	reBirdParens      = regexp.MustCompile(`\(([^)]+)\)`)
)

func (d *birdDriver) LookupRoute(ctx context.Context, target string) (RouteLookupResult, error) {
	norm, _, err := ParseQueryTarget(target)
	if err != nil {
		return RouteLookupResult{}, err
	}

	raw, err := d.querySocket(ctx, "show route for "+norm+" all")
	if err != nil {
		return RouteLookupResult{}, err
	}

	res := RouteLookupResult{
		Target:    norm,
		Driver:    "bird",
		RouterASN: normalizeRouterASN(d.cfg.RouterASN),
		RawOutput: raw,
	}

	var cur *RoutePath
	flushCur := func() {
		if cur != nil {
			res.Paths = append(res.Paths, *cur)
			cur = nil
		}
	}

	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Table ") {
			continue
		}

		if strings.Contains(line, "[") && (strings.Contains(line, "unicast") || !strings.HasPrefix(line, "\t")) {
			if m := reBirdRouteHeader.FindStringSubmatch(line); len(m) >= 5 {
				flushCur()
				if m[1] != "" && (strings.Contains(m[1], ".") || strings.Contains(m[1], ":")) {
					res.Prefix = m[1]
				}
				metaParts := strings.Fields(m[3])
				since := ""
				nextHop := ""
				if len(metaParts) > 0 {
					since = metaParts[0]
				}
				for idx, tok := range metaParts {
					if tok == "from" && idx+1 < len(metaParts) {
						nextHop = metaParts[idx+1]
						break
					}
				}
				cur = &RoutePath{
					Best:     m[4] == "*",
					Protocol: m[2],
					Since:    since,
					NextHop:  nextHop,
				}
				continue
			}
		}

		if cur == nil {
			continue
		}

		switch {
		case strings.HasPrefix(trimmed, "via "):
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				cur.NextHop = fields[1]
			}
			for i, f := range fields {
				if f == "on" && i+1 < len(fields) {
					cur.Interface = fields[i+1]
				}
			}
		case strings.HasPrefix(trimmed, "BGP.next_hop:"):
			fields := strings.Fields(strings.TrimPrefix(trimmed, "BGP.next_hop:"))
			if len(fields) > 0 && cur.NextHop == "" {
				cur.NextHop = fields[0]
			}
		case strings.HasPrefix(trimmed, "BGP.as_path:"):
			tokens := strings.Fields(strings.TrimPrefix(trimmed, "BGP.as_path:"))
			cur.ASPath = buildASHops(tokens)
		case strings.HasPrefix(trimmed, "BGP.local_pref:"):
			v, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(trimmed, "BGP.local_pref:")))
			cur.LocalPref = v
		case strings.HasPrefix(trimmed, "BGP.med:"):
			v, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(trimmed, "BGP.med:")))
			cur.MED = v
		case strings.HasPrefix(trimmed, "BGP.origin:"):
			cur.Origin = strings.TrimSpace(strings.TrimPrefix(trimmed, "BGP.origin:"))
		case strings.HasPrefix(trimmed, "BGP.community:") || strings.HasPrefix(trimmed, "BGP.large_community:"):
			matches := reBirdParens.FindAllStringSubmatch(trimmed, -1)
			for _, cm := range matches {
				if len(cm) >= 2 {
					cur.Communities = append(cur.Communities, decorateCommunity(cm[1], d.cfg.Communities))
				}
			}
		}
	}
	flushCur()

	return res, nil
}

func (d *birdDriver) Summary(ctx context.Context) (SummaryResult, error) {
	raw, err := d.querySocket(ctx, "show protocols")
	if err != nil {
		return SummaryResult{}, err
	}

	res := SummaryResult{
		Driver:    "bird",
		RouterASN: normalizeRouterASN(d.cfg.RouterASN),
		RawOutput: raw,
	}

	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[0] == "Name" || fields[1] != "BGP" {
			continue
		}
		info := ""
		if len(fields) >= 6 {
			info = strings.Join(fields[5:], " ")
		}
		res.Peers = append(res.Peers, PeerSession{
			Name:  fields[0],
			State: fields[3],
			Since: fields[4],
			Info:  info,
			Org:   enrich.LookupASNName(fields[0]),
		})
	}
	return res, nil
}
