package probe

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"math"
	"math/rand/v2"
	"net"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudapp3/vmlg/internal/enrich"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

type hopStat struct {
	addr netip.Addr
	sent int
	rtts []float64
}

func (h *hopStat) recordHit(peer netip.Addr, rttMs float64) {
	if !h.addr.IsValid() && peer.IsValid() {
		h.addr = peer
	}
	h.rtts = append(h.rtts, rttMs)
}

func RunMTR(ctx context.Context, req MTRRequest, enricher *enrich.Enricher, onUpdate func(MTRSnapshot)) (MTRSnapshot, error) {
	if !req.TargetAddr.IsValid() {
		return MTRSnapshot{}, errors.New("target_addr is required")
	}
	if req.Cycles <= 0 {
		req.Cycles = 5
	}
	if req.MaxHops <= 0 {
		req.MaxHops = 30
	}
	if req.TimeoutMs <= 0 {
		req.TimeoutMs = 1200
	}
	if enricher == nil {
		enricher = enrich.New(false, false)
	}

	snap, err := runNativeICMPMTR(ctx, req, enricher, onUpdate)
	if err != nil && isPermError(err) {
		if fbSnap, handled, fbErr := runPlatformFallbackMTR(ctx, req, enricher, onUpdate); handled {
			return fbSnap, fbErr
		}
		return runSystemMTRFallback(ctx, req, enricher, onUpdate)
	}
	return snap, err
}

func runNativeICMPMTR(ctx context.Context, req MTRRequest, enricher *enrich.Enricher, onUpdate func(MTRSnapshot)) (MTRSnapshot, error) {
	ip := req.TargetAddr.Unmap()
	var (
		network    string
		listenAddr string
		protoNum   int
		echoType   icmp.Type
	)
	if ip.Is4() {
		network = "ip4:icmp"
		listenAddr = "0.0.0.0"
		protoNum = 1
		echoType = ipv4.ICMPTypeEcho
	} else {
		network = "ip6:ipv6-icmp"
		listenAddr = "::"
		protoNum = 58
		echoType = ipv6.ICMPTypeEchoRequest
	}

	conn, err := icmp.ListenPacket(network, listenAddr)
	if err != nil {
		return MTRSnapshot{}, err
	}
	defer conn.Close()

	dst := &net.IPAddr{IP: net.IP(ip.AsSlice())}
	sessionID := uint16(rand.IntN(0xfffe) + 1)
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	payload := []byte("vmlg-mtr-probe")

	stats := make([]hopStat, req.MaxHops+1)
	effectiveMaxHop := req.MaxHops

	var lastSnap MTRSnapshot
	buf := make([]byte, 1500)

	for cycle := 1; cycle <= req.Cycles; cycle++ {
		if ctx.Err() != nil {
			break
		}

		sentTimes := make(map[uint16]time.Time, effectiveMaxHop)
		responded := make(map[int]bool, effectiveMaxHop)

		// Send TTL 1..effectiveMaxHop probes with slight pacing (8ms) to avoid burst drops
		for ttl := 1; ttl <= effectiveMaxHop; ttl++ {
			if ctx.Err() != nil {
				break
			}
			if ip.Is4() {
				_ = conn.IPv4PacketConn().SetTTL(ttl)
			} else {
				_ = conn.IPv6PacketConn().SetHopLimit(ttl)
			}

			seq := uint16((cycle << 8) | (ttl & 0xff))
			msg := icmp.Message{
				Type: echoType,
				Code: 0,
				Body: &icmp.Echo{
					ID:   int(sessionID),
					Seq:  int(seq),
					Data: payload,
				},
			}
			wire, merr := msg.Marshal(nil)
			if merr != nil {
				continue
			}

			stats[ttl].sent++
			sentTimes[seq] = time.Now()
			_, _ = conn.WriteTo(wire, dst)

			if ttl < effectiveMaxHop {
				time.Sleep(8 * time.Millisecond)
			}
		}

		// Collect responses until all sent TTLs respond or timeout elapses
		waitDeadline := time.Now().Add(timeout)
		for len(responded) < len(sentTimes) && time.Now().Before(waitDeadline) {
			if ctx.Err() != nil {
				break
			}
			_ = conn.SetReadDeadline(minTime(waitDeadline, time.Now().Add(150*time.Millisecond)))
			n, peer, rerr := conn.ReadFrom(buf)
			if rerr != nil {
				var netErr net.Error
				if errors.As(rerr, &netErr) && netErr.Timeout() {
					continue
				}
				break
			}

			recvAt := time.Now()
			peerAddr := addrFromNetAddr(peer)
			if !peerAddr.IsValid() {
				continue
			}

			origID, origSeq, isDest, ok := matchMTRPacket(ip.Is4(), protoNum, buf[:n])
			if !ok || origID != sessionID {
				continue
			}
			sentAt, exists := sentTimes[origSeq]
			if !exists {
				continue
			}
			ttl := int(origSeq & 0xff)
			if ttl < 1 || ttl > req.MaxHops || responded[ttl] {
				continue
			}

			responded[ttl] = true
			rttMs := float64(recvAt.Sub(sentAt).Microseconds()) / 1000.0
			stats[ttl].recordHit(peerAddr, rttMs)

			if (isDest || peerAddr == ip) && ttl < effectiveMaxHop {
				effectiveMaxHop = ttl
			}
		}

		completed := cycle == req.Cycles
		lastSnap = buildMTRSnapshot(ctx, req, "go-icmp-mtr", cycle, completed, stats[1:effectiveMaxHop+1], ip, enricher)
		if onUpdate != nil {
			onUpdate(lastSnap)
		}
	}

	lastSnap.Completed = true
	return lastSnap, nil
}

// matchMTRPacket extracts the original ICMP Echo ID and Sequence from either an
// EchoReply (destination reached) or TimeExceeded / DestinationUnreachable inner header.
func matchMTRPacket(isV4 bool, protoNum int, raw []byte) (id uint16, seq uint16, isDest bool, ok bool) {
	parsed, err := icmp.ParseMessage(protoNum, raw)
	if err != nil {
		return 0, 0, false, false
	}

	if echo, isEcho := parsed.Body.(*icmp.Echo); isEcho {
		if (isV4 && parsed.Type == ipv4.ICMPTypeEchoReply) || (!isV4 && parsed.Type == ipv6.ICMPTypeEchoReply) {
			return uint16(echo.ID), uint16(echo.Seq), true, true
		}
		return 0, 0, false, false
	}

	var inner []byte
	switch b := parsed.Body.(type) {
	case *icmp.TimeExceeded:
		inner = b.Data
	case *icmp.DstUnreach:
		inner = b.Data
		isDest = true
	default:
		return 0, 0, false, false
	}

	if isV4 {
		if len(inner) < 28 {
			return 0, 0, false, false
		}
		ihl := int(inner[0]&0x0f) * 4
		if ihl < 20 || len(inner) < ihl+8 || inner[9] != 1 {
			return 0, 0, false, false
		}
		id = binary.BigEndian.Uint16(inner[ihl+4 : ihl+6])
		seq = binary.BigEndian.Uint16(inner[ihl+6 : ihl+8])
		return id, seq, isDest, true
	}

	// IPv6 inner packet: 40-byte fixed header + 8-byte ICMPv6 Echo header
	if len(inner) < 48 || inner[6] != 58 {
		return 0, 0, false, false
	}
	id = binary.BigEndian.Uint16(inner[44:46])
	seq = binary.BigEndian.Uint16(inner[46:48])
	return id, seq, isDest, true
}

func runSystemMTRFallback(ctx context.Context, req MTRRequest, enricher *enrich.Enricher, onUpdate func(MTRSnapshot)) (MTRSnapshot, error) {
	ip := req.TargetAddr.Unmap()
	if mtrBin, err := exec.LookPath("mtr"); err == nil {
		return execMTRRaw(ctx, mtrBin, req, ip, enricher, onUpdate)
	}
	if trBin, err := exec.LookPath("traceroute"); err == nil {
		return execTraceroute(ctx, trBin, req, ip, enricher, onUpdate)
	}
	return MTRSnapshot{}, errors.New("raw ICMP socket not permitted and neither mtr nor traceroute is installed")
}

func execMTRRaw(ctx context.Context, bin string, req MTRRequest, ip netip.Addr, enricher *enrich.Enricher, onUpdate func(MTRSnapshot)) (MTRSnapshot, error) {
	args := []string{"--raw", "-n", "-c", strconv.Itoa(req.Cycles), "-m", strconv.Itoa(req.MaxHops)}
	if ip.Is6() {
		args = append(args, "-6")
	} else {
		args = append(args, "-4")
	}
	args = append(args, ip.String())

	totalTimeout := time.Duration(req.Cycles*2500+8000) * time.Millisecond
	cmdCtx, cancel := context.WithTimeout(ctx, totalTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return MTRSnapshot{}, err
	}
	if err := cmd.Start(); err != nil {
		return MTRSnapshot{}, err
	}

	stats := make([]hopStat, req.MaxHops+1)
	maxSeenHop := 1
	targetHop := req.MaxHops

	scanner := bufio.NewScanner(stdout)
	lastEmit := time.Now()
	emitCount := 0

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		hopIdx, herr := strconv.Atoi(fields[1])
		if herr != nil || hopIdx < 0 || hopIdx >= req.MaxHops {
			continue
		}
		ttl := hopIdx + 1
		if ttl > maxSeenHop {
			maxSeenHop = ttl
		}

		switch fields[0] {
		case "h":
			if parsedIP, perr := netip.ParseAddr(fields[2]); perr == nil {
				parsedIP = parsedIP.Unmap()
				stats[ttl].addr = parsedIP
				if parsedIP == ip && ttl < targetHop {
					targetHop = ttl
				}
			}
		case "x":
			stats[ttl].sent++
		case "p":
			usec, perr := strconv.ParseFloat(fields[2], 64)
			if perr == nil {
				if stats[ttl].sent <= len(stats[ttl].rtts) {
					stats[ttl].sent = len(stats[ttl].rtts) + 1
				}
				stats[ttl].rtts = append(stats[ttl].rtts, usec/1000.0)
			}
			if time.Since(lastEmit) > 400*time.Millisecond {
				lastEmit = time.Now()
				emitCount = min(req.Cycles, emitCount+1)
				limit := min(maxSeenHop, targetHop)
				snap := buildMTRSnapshot(ctx, req, "sys-mtr", max(1, emitCount), false, stats[1:limit+1], ip, enricher)
				if onUpdate != nil {
					onUpdate(snap)
				}
			}
		}
	}
	_ = cmd.Wait()

	limit := min(maxSeenHop, targetHop)
	finalSnap := buildMTRSnapshot(ctx, req, "sys-mtr", req.Cycles, true, stats[1:limit+1], ip, enricher)
	if onUpdate != nil {
		onUpdate(finalSnap)
	}
	return finalSnap, nil
}

func execTraceroute(ctx context.Context, bin string, req MTRRequest, ip netip.Addr, enricher *enrich.Enricher, onUpdate func(MTRSnapshot)) (MTRSnapshot, error) {
	args := []string{"-n", "-q", "1", "-w", "1", "-m", strconv.Itoa(req.MaxHops), ip.String()}
	cmdCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, bin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return MTRSnapshot{}, err
	}
	if err := cmd.Start(); err != nil {
		return MTRSnapshot{}, err
	}

	stats := make([]hopStat, req.MaxHops+1)
	maxSeenHop := 0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		ttl, herr := strconv.Atoi(fields[0])
		if herr != nil || ttl < 1 || ttl > req.MaxHops {
			continue
		}
		maxSeenHop = max(maxSeenHop, ttl)
		stats[ttl].sent = 1
		if fields[1] != "*" {
			if parsedIP, perr := netip.ParseAddr(fields[1]); perr == nil {
				stats[ttl].addr = parsedIP.Unmap()
			}
			for i := 2; i < len(fields)-1; i++ {
				if fields[i+1] == "ms" {
					if v, ferr := strconv.ParseFloat(fields[i], 64); ferr == nil {
						stats[ttl].rtts = append(stats[ttl].rtts, v)
						break
					}
				}
			}
		}
		snap := buildMTRSnapshot(ctx, req, "sys-traceroute", 1, false, stats[1:maxSeenHop+1], ip, enricher)
		if onUpdate != nil {
			onUpdate(snap)
		}
	}
	_ = cmd.Wait()

	if maxSeenHop == 0 {
		maxSeenHop = 1
	}
	finalSnap := buildMTRSnapshot(ctx, req, "sys-traceroute", 1, true, stats[1:maxSeenHop+1], ip, enricher)
	if onUpdate != nil {
		onUpdate(finalSnap)
	}
	return finalSnap, nil
}

func buildMTRSnapshot(ctx context.Context, req MTRRequest, engine string, cycle int, completed bool, rawHops []hopStat, targetIP netip.Addr, enricher *enrich.Enricher) MTRSnapshot {
	// Trim trailing all-loss hops beyond targetIP
	trimLen := len(rawHops)
	for i, h := range rawHops {
		if h.addr == targetIP {
			trimLen = i + 1
			break
		}
	}
	if completed {
		for trimLen > 1 && !rawHops[trimLen-1].addr.IsValid() {
			trimLen--
		}
	}
	rawHops = rawHops[:trimLen]

	metas := make([]enrich.HopMeta, len(rawHops))
	var wg sync.WaitGroup
	for i, h := range rawHops {
		if h.addr.IsValid() {
			wg.Add(1)
			go func(idx int, addr netip.Addr) {
				defer wg.Done()
				metas[idx] = enricher.Lookup(ctx, addr)
			}(i, h.addr)
		}
	}
	wg.Wait()

	hops := make([]MTRHop, 0, len(rawHops))
	validMetas := make([]enrich.HopMeta, 0, len(rawHops))

	for i, h := range rawHops {
		sent := max(h.sent, len(h.rtts))
		if sent == 0 {
			sent = cycle
		}
		recv := len(h.rtts)
		lossPct := 0.0
		if sent > 0 {
			lossPct = roundMs(float64(sent-recv) * 100.0 / float64(sent))
		}

		mh := MTRHop{
			Hop:     i + 1,
			Sent:    sent,
			Recv:    recv,
			LossPct: lossPct,
		}
		if h.addr.IsValid() {
			mh.Addr = h.addr.String()
			m := metas[i]
			mh.PTR = m.PTR
			mh.ASN = m.ASN
			mh.Org = m.Org
			mh.LineTag = m.LineTag
			mh.Tier = m.Tier
			mh.IXP = m.IXP
			validMetas = append(validMetas, m)
		}
		if recv > 0 {
			best, wrst, sum := h.rtts[0], h.rtts[0], 0.0
			for _, v := range h.rtts {
				if v < best {
					best = v
				}
				if v > wrst {
					wrst = v
				}
				sum += v
			}
			avg := sum / float64(recv)
			var sq float64
			for _, v := range h.rtts {
				d := v - avg
				sq += d * d
			}
			mh.LastMs = roundMs(h.rtts[recv-1])
			mh.BestMs = roundMs(best)
			mh.AvgMs = roundMs(avg)
			mh.WrstMs = roundMs(wrst)
			mh.StDevMs = roundMs(math.Sqrt(sq / float64(recv)))
		}
		hops = append(hops, mh)
	}

	return MTRSnapshot{
		Target:    req.Target,
		Addr:      targetIP.String(),
		Engine:    engine,
		Cycle:     cycle,
		MaxCycles: req.Cycles,
		Completed: completed,
		Route:     enrich.SummarizeRoute(validMetas),
		Hops:      hops,
	}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
