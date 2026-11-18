//go:build linux

package probe

import (
	"context"
	"encoding/binary"
	"net/netip"
	"time"

	"github.com/cloudapp3/vmlg/internal/enrich"
	"golang.org/x/sys/unix"
)

// runPlatformFallbackMTR executes unprivileged hop-by-hop UDP MTR using Linux
// IP_RECVERR / IPV6_RECVERR + MSG_ERRQUEUE (the same kernel mechanism as tracepath),
// requiring neither root nor CAP_NET_RAW nor setuid binaries.
func runPlatformFallbackMTR(ctx context.Context, req MTRRequest, enricher *enrich.Enricher, onUpdate func(MTRSnapshot)) (MTRSnapshot, bool, error) {
	ip := req.TargetAddr.Unmap()
	isV4 := ip.Is4()

	domain := unix.AF_INET
	if !isV4 {
		domain = unix.AF_INET6
	}

	// Quick probe: verify socket creation + RECVERR option works
	testFd, err := unix.Socket(domain, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, unix.IPPROTO_UDP)
	if err != nil {
		return MTRSnapshot{}, false, err
	}
	_ = unix.Close(testFd)

	stats := make([]hopStat, req.MaxHops+1)
	effectiveMaxHop := req.MaxHops
	var lastSnap MTRSnapshot

	for cycle := 1; cycle <= req.Cycles; cycle++ {
		if ctx.Err() != nil {
			break
		}

		for ttl := 1; ttl <= effectiveMaxHop; ttl++ {
			if ctx.Err() != nil {
				break
			}
			stats[ttl].sent++
			peer, rttMs, reachedDest, ok := probeSingleUDPTTL(ctx, ip, ttl, req.TimeoutMs)
			if ok {
				stats[ttl].recordHit(peer, rttMs)
				if (reachedDest || peer == ip) && ttl < effectiveMaxHop {
					effectiveMaxHop = ttl
					break
				}
			}
		}

		completed := cycle == req.Cycles
		lastSnap = buildMTRSnapshot(ctx, req, "go-udp-errqueue", cycle, completed, stats[1:effectiveMaxHop+1], ip, enricher)
		if onUpdate != nil {
			onUpdate(lastSnap)
		}
	}

	lastSnap.Completed = true
	return lastSnap, true, nil
}

func probeSingleUDPTTL(ctx context.Context, ip netip.Addr, ttl int, timeoutMs int) (peer netip.Addr, rttMs float64, isDest bool, ok bool) {
	isV4 := ip.Is4()
	domain := unix.AF_INET
	if !isV4 {
		domain = unix.AF_INET6
	}

	fd, err := unix.Socket(domain, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK, unix.IPPROTO_UDP)
	if err != nil {
		return netip.Addr{}, 0, false, false
	}
	defer unix.Close(fd)

	if isV4 {
		if err := unix.SetsockoptInt(fd, unix.SOL_IP, unix.IP_RECVERR, 1); err != nil {
			return netip.Addr{}, 0, false, false
		}
		if err := unix.SetsockoptInt(fd, unix.SOL_IP, unix.IP_TTL, ttl); err != nil {
			return netip.Addr{}, 0, false, false
		}
	} else {
		if err := unix.SetsockoptInt(fd, unix.SOL_IPV6, unix.IPV6_RECVERR, 1); err != nil {
			return netip.Addr{}, 0, false, false
		}
		if err := unix.SetsockoptInt(fd, unix.SOL_IPV6, unix.IPV6_UNICAST_HOPS, ttl); err != nil {
			return netip.Addr{}, 0, false, false
		}
	}

	dstPort := 33434 + (ttl % 64)
	var sa unix.Sockaddr
	if isV4 {
		b := ip.As4()
		sa = &unix.SockaddrInet4{Port: dstPort, Addr: b}
	} else {
		b := ip.As16()
		sa = &unix.SockaddrInet6{Port: dstPort, Addr: b}
	}

	payload := []byte("vmlg-probe")
	start := time.Now()
	if err := unix.Sendto(fd, payload, 0, sa); err != nil {
		return netip.Addr{}, 0, false, false
	}

	deadline := start.Add(time.Duration(min(timeoutMs, 800)) * time.Millisecond)
	buf := make([]byte, 512)
	oob := make([]byte, 512)

	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return netip.Addr{}, 0, false, false
		}
		remainMs := int(time.Until(deadline).Milliseconds())
		if remainMs <= 0 {
			break
		}
		pfds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLERR}}
		n, perr := unix.Poll(pfds, min(remainMs, 100))
		if perr != nil || n == 0 {
			continue
		}

		_, oobn, _, _, rerr := unix.Recvmsg(fd, buf, oob, unix.MSG_ERRQUEUE)
		if rerr != nil || oobn == 0 {
			continue
		}
		rtt := float64(time.Since(start).Microseconds()) / 1000.0

		cmsgs, cerr := unix.ParseSocketControlMessage(oob[:oobn])
		if cerr != nil {
			continue
		}
		for _, cm := range cmsgs {
			isIPv4Err := cm.Header.Level == unix.SOL_IP && cm.Header.Type == unix.IP_RECVERR
			isIPv6Err := cm.Header.Level == unix.SOL_IPV6 && cm.Header.Type == unix.IPV6_RECVERR
			if (!isIPv4Err && !isIPv6Err) || len(cm.Data) < 20 {
				continue
			}

			eeOrigin := cm.Data[4]
			eeType := cm.Data[5]
			family := binary.NativeEndian.Uint16(cm.Data[16:18])

			var offender netip.Addr
			if family == unix.AF_INET && len(cm.Data) >= 24 {
				var a4 [4]byte
				copy(a4[:], cm.Data[20:24])
				offender = netip.AddrFrom4(a4)
			} else if family == unix.AF_INET6 && len(cm.Data) >= 44 {
				var a16 [16]byte
				copy(a16[:], cm.Data[24:40])
				offender = netip.AddrFrom16(a16).Unmap()
			}
			if !offender.IsValid() {
				continue
			}

			dest := offender == ip ||
				(eeOrigin == unix.SO_EE_ORIGIN_ICMP && eeType == 3) ||
				(eeOrigin == unix.SO_EE_ORIGIN_ICMP6 && eeType == 1)

			return offender, roundMs(rtt), dest, true
		}
	}

	return netip.Addr{}, 0, false, false
}
