package probe

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"net"
	"net/netip"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func RunPing(ctx context.Context, req PingRequest, onSample func(PingSample)) (PingSummary, error) {
	if !req.TargetAddr.IsValid() {
		return PingSummary{}, errors.New("target_addr is required")
	}
	if req.Count <= 0 {
		req.Count = 4
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = 1500
	}
	if req.IntervalMs <= 0 {
		req.IntervalMs = 350
	}
	proto := strings.ToLower(strings.TrimSpace(req.Protocol))
	if proto == "" {
		proto = "icmp"
	}

	switch proto {
	case "tcp":
		port := req.Port
		if port <= 0 || port > 65535 {
			port = 443
		}
		return runTCPPing(ctx, req, port, false, "go-tcp", onSample)
	case "icmp":
		sum, err := runNativeICMPPing(ctx, req, onSample)
		if err != nil && isPermError(err) {
			sum, err = runSystemPingFallback(ctx, req, onSample)
			if err != nil && isPermError(err) {
				port := req.Port
				if port <= 0 || port > 65535 {
					port = 443
				}
				return runTCPPing(ctx, req, port, true, "go-tcp-unpriv-fallback", onSample)
			}
		}
		return sum, err
	default:
		return PingSummary{}, fmt.Errorf("unsupported ping protocol %q", proto)
	}
}

func runTCPPing(ctx context.Context, req PingRequest, port int, acceptRST bool, engine string, onSample func(PingSample)) (PingSummary, error) {
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	interval := time.Duration(req.IntervalMs) * time.Millisecond
	targetHostPort := net.JoinHostPort(req.TargetAddr.String(), strconv.Itoa(port))
	protoLabel := fmt.Sprintf("tcp/%d", port)

	samples := make([]PingSample, 0, req.Count)
	for seq := 1; seq <= req.Count; seq++ {
		if err := ctx.Err(); err != nil {
			break
		}
		dialer := net.Dialer{Timeout: timeout}
		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", targetHostPort)
		rtt := float64(time.Since(start).Microseconds()) / 1000.0

		sample := PingSample{
			Seq:    seq,
			Target: req.Target,
			Addr:   req.TargetAddr.String(),
			Proto:  protoLabel,
		}
		if err == nil {
			_ = conn.Close()
			sample.RTTMs = roundMs(rtt)
		} else if acceptRST && errors.Is(err, syscall.ECONNREFUSED) {
			// Remote host actively replied with TCP RST -> proves reachability & measures 1-RTT
			sample.RTTMs = roundMs(rtt)
		} else if errors.Is(err, context.Canceled) {
			break
		} else {
			sample.Timeout = true
			sample.Error = conciseNetErr(err)
		}
		samples = append(samples, sample)
		if onSample != nil {
			onSample(sample)
		}
		if seq < req.Count {
			if !sleepCtx(ctx, interval) {
				break
			}
		}
	}

	return computePingSummary(req.Target, req.TargetAddr.String(), protoLabel, engine, samples), nil
}

func runNativeICMPPing(ctx context.Context, req PingRequest, onSample func(PingSample)) (PingSummary, error) {
	ip := req.TargetAddr.Unmap()
	var (
		rawNet, dgramNet string
		listenAddr       string
		protoNum         int
		echoType         icmp.Type
		replyType        icmp.Type
	)
	if ip.Is4() {
		rawNet, dgramNet = "ip4:icmp", "udp4"
		listenAddr = "0.0.0.0"
		protoNum = 1
		echoType = ipv4.ICMPTypeEcho
		replyType = ipv4.ICMPTypeEchoReply
	} else {
		rawNet, dgramNet = "ip6:ipv6-icmp", "udp6"
		listenAddr = "::"
		protoNum = 58
		echoType = ipv6.ICMPTypeEchoRequest
		replyType = ipv6.ICMPTypeEchoReply
	}

	isDatagram := false
	conn, err := icmp.ListenPacket(rawNet, listenAddr)
	if err != nil {
		conn, err = icmp.ListenPacket(dgramNet, listenAddr)
		if err != nil {
			return PingSummary{}, err
		}
		isDatagram = true
	}
	defer conn.Close()

	if ip.Is4() && conn.IPv4PacketConn() != nil {
		_ = conn.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
	} else if ip.Is6() && conn.IPv6PacketConn() != nil {
		_ = conn.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit, true)
	}

	var dst net.Addr
	if isDatagram {
		dst = &net.UDPAddr{IP: net.IP(ip.AsSlice())}
	} else {
		dst = &net.IPAddr{IP: net.IP(ip.AsSlice())}
	}

	echoID := rand.IntN(0xfffe) + 1
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	interval := time.Duration(req.IntervalMs) * time.Millisecond
	payload := []byte("vmlg- looking-glass-probe")

	samples := make([]PingSample, 0, req.Count)
	buf := make([]byte, 1500)

	for seq := 1; seq <= req.Count; seq++ {
		if err := ctx.Err(); err != nil {
			break
		}

		msg := icmp.Message{
			Type: echoType,
			Code: 0,
			Body: &icmp.Echo{
				ID:   echoID,
				Seq:  seq,
				Data: payload,
			},
		}
		wire, err := msg.Marshal(nil)
		if err != nil {
			return PingSummary{}, err
		}

		start := time.Now()
		deadline := start.Add(timeout)
		_ = conn.SetDeadline(deadline)

		sample := PingSample{
			Seq:    seq,
			Target: req.Target,
			Addr:   ip.String(),
			Proto:  "icmp",
		}

		if _, err := conn.WriteTo(wire, dst); err != nil {
			sample.Timeout = true
			sample.Error = conciseNetErr(err)
			samples = append(samples, sample)
			if onSample != nil {
				onSample(sample)
			}
			continue
		}

		matched := false
		for {
			if err := ctx.Err(); err != nil {
				return computePingSummary(req.Target, ip.String(), "icmp", "go-icmp", samples), nil
			}
			var (
				n    int
				peer net.Addr
				ttl  int
				rerr error
			)
			if ip.Is4() && conn.IPv4PacketConn() != nil {
				var cm *ipv4.ControlMessage
				n, cm, peer, rerr = conn.IPv4PacketConn().ReadFrom(buf)
				if cm != nil {
					ttl = cm.TTL
				}
			} else if ip.Is6() && conn.IPv6PacketConn() != nil {
				var cm *ipv6.ControlMessage
				n, cm, peer, rerr = conn.IPv6PacketConn().ReadFrom(buf)
				if cm != nil {
					ttl = cm.HopLimit
				}
			} else {
				n, peer, rerr = conn.ReadFrom(buf)
			}

			if rerr != nil {
				sample.Timeout = true
				sample.Error = "timeout"
				break
			}

			rtt := float64(time.Since(start).Microseconds()) / 1000.0
			peerAddr := addrFromNetAddr(peer)
			if peerAddr.IsValid() && peerAddr != ip {
				continue
			}

			parsed, perr := icmp.ParseMessage(protoNum, buf[:n])
			if perr != nil || parsed.Type != replyType {
				continue
			}
			echoBody, ok := parsed.Body.(*icmp.Echo)
			if !ok {
				continue
			}
			if !isDatagram && echoBody.ID != echoID {
				continue
			}
			if echoBody.Seq != seq {
				continue
			}

			sample.RTTMs = roundMs(rtt)
			sample.TTL = ttl
			matched = true
			break
		}
		if !matched && !sample.Timeout {
			sample.Timeout = true
		}

		samples = append(samples, sample)
		if onSample != nil {
			onSample(sample)
		}
		if seq < req.Count {
			if !sleepCtx(ctx, interval) {
				break
			}
		}
	}

	engine := "go-icmp"
	if isDatagram {
		engine = "go-icmp-dgram"
	}
	return computePingSummary(req.Target, ip.String(), "icmp", engine, samples), nil
}

var (
	rePingReply = regexp.MustCompile(`icmp_seq=(\d+)\s+(?:ttl=(\d+)\s+)?time=([0-9.]+)\s*ms`)
)

func runSystemPingFallback(ctx context.Context, req PingRequest, onSample func(PingSample)) (PingSummary, error) {
	bin, err := exec.LookPath("ping")
	if err != nil {
		return PingSummary{}, errors.New("ICMP socket not permitted and system ping binary not found")
	}

	ip := req.TargetAddr.Unmap()
	intervalSec := fmt.Sprintf("%.2f", max(0.2, float64(req.IntervalMs)/1000.0))
	timeoutSec := strconv.Itoa(max(1, (req.TimeoutMs+999)/1000))

	args := []string{"-n", "-c", strconv.Itoa(req.Count), "-i", intervalSec, "-W", timeoutSec}
	if ip.Is6() {
		args = append([]string{"-6"}, args...)
	} else {
		args = append([]string{"-4"}, args...)
	}
	args = append(args, ip.String())

	totalTimeout := time.Duration(req.Count*(req.IntervalMs+req.TimeoutMs)+3000) * time.Millisecond
	cmdCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return PingSummary{}, err
	}
	if err := cmd.Start(); err != nil {
		return PingSummary{}, err
	}

	receivedSeq := make(map[int]PingSample, req.Count)
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		m := rePingReply.FindStringSubmatch(line)
		if len(m) == 4 {
			seq, _ := strconv.Atoi(m[1])
			ttl, _ := strconv.Atoi(m[2])
			rtt, _ := strconv.ParseFloat(m[3], 64)
			s := PingSample{
				Seq:    seq,
				Target: req.Target,
				Addr:   ip.String(),
				Proto:  "icmp",
				TTL:    ttl,
				RTTMs:  roundMs(rtt),
			}
			receivedSeq[seq] = s
			if onSample != nil {
				onSample(s)
			}
		}
	}
	_ = cmd.Wait()

	samples := make([]PingSample, 0, req.Count)
	for seq := 1; seq <= req.Count; seq++ {
		if s, ok := receivedSeq[seq]; ok {
			samples = append(samples, s)
		} else {
			to := PingSample{
				Seq:     seq,
				Target:  req.Target,
				Addr:    ip.String(),
				Proto:   "icmp",
				Timeout: true,
				Error:   "timeout",
			}
			samples = append(samples, to)
			if onSample != nil {
				onSample(to)
			}
		}
	}

	return computePingSummary(req.Target, ip.String(), "icmp", "sys-ping", samples), nil
}

func computePingSummary(target, addr, proto, engine string, samples []PingSample) PingSummary {
	sum := PingSummary{
		Target:   target,
		Addr:     addr,
		Protocol: proto,
		Engine:   engine,
		Sent:     len(samples),
		Samples:  samples,
	}
	if len(samples) == 0 {
		return sum
	}

	var rtts []float64
	var total float64
	for _, s := range samples {
		if !s.Timeout && s.Error == "" {
			rtts = append(rtts, s.RTTMs)
			total += s.RTTMs
		}
	}
	sum.Recv = len(rtts)
	sum.LossPct = roundMs(float64(sum.Sent-sum.Recv) * 100.0 / float64(sum.Sent))

	if len(rtts) == 0 {
		return sum
	}

	minV, maxV := rtts[0], rtts[0]
	var jitterSum float64
	for i, v := range rtts {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
		if i > 0 {
			jitterSum += math.Abs(v - rtts[i-1])
		}
	}
	avg := total / float64(len(rtts))

	var sqDiff float64
	for _, v := range rtts {
		d := v - avg
		sqDiff += d * d
	}

	sum.MinMs = roundMs(minV)
	sum.AvgMs = roundMs(avg)
	sum.MaxMs = roundMs(maxV)
	if len(rtts) > 1 {
		sum.JitterMs = roundMs(jitterSum / float64(len(rtts)-1))
	}
	sum.StDevMs = roundMs(math.Sqrt(sqDiff / float64(len(rtts))))
	return sum
}

func addrFromNetAddr(a net.Addr) netip.Addr {
	switch v := a.(type) {
	case *net.IPAddr:
		if ip, ok := netip.AddrFromSlice(v.IP); ok {
			return ip.Unmap()
		}
	case *net.UDPAddr:
		if ip, ok := netip.AddrFromSlice(v.IP); ok {
			return ip.Unmap()
		}
	}
	return netip.Addr{}
}

func isPermError(err error) bool {
	return errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) || strings.Contains(strings.ToLower(err.Error()), "operation not permitted") || strings.Contains(strings.ToLower(err.Error()), "permission denied")
}

func conciseNetErr(err error) string {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	if errors.Is(err, syscall.ECONNREFUSED) {
		return "connection refused"
	}
	if errors.Is(err, syscall.ENETUNREACH) || errors.Is(err, syscall.EHOSTUNREACH) {
		return "unreachable"
	}
	return "error"
}

func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func roundMs(v float64) float64 {
	return math.Round(v*100) / 100
}
