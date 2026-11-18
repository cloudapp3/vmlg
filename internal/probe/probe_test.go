package probe

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"strconv"
	"testing"

	"github.com/cloudapp3/vmlg/internal/enrich"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func TestTCPPingLoopback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	var streamed int
	sum, err := RunPing(context.Background(), PingRequest{
		Target:     "127.0.0.1",
		TargetAddr: netip.MustParseAddr("127.0.0.1"),
		Protocol:   "tcp",
		Port:       port,
		Count:      3,
		TimeoutMs:  500,
		IntervalMs: 20,
	}, func(_ PingSample) {
		streamed++
	})
	if err != nil {
		t.Fatalf("TCP ping failed: %v", err)
	}
	if sum.Sent != 3 || sum.Recv != 3 || streamed != 3 || sum.LossPct != 0 {
		t.Fatalf("unexpected TCP ping summary: %+v (streamed=%d)", sum, streamed)
	}
}

func TestICMPPingAndMTRLoopback(t *testing.T) {
	sum, err := RunPing(context.Background(), PingRequest{
		Target:     "127.0.0.1",
		TargetAddr: netip.MustParseAddr("127.0.0.1"),
		Protocol:   "icmp",
		Count:      2,
		TimeoutMs:  800,
		IntervalMs: 50,
	}, nil)
	if err != nil {
		t.Fatalf("ICMP ping loopback failed: %v", err)
	}
	if sum.Recv == 0 {
		t.Fatalf("expected loopback ICMP reply, got summary: %+v", sum)
	}

	var updates int
	snap, err := RunMTR(context.Background(), MTRRequest{
		Target:     "127.0.0.1",
		TargetAddr: netip.MustParseAddr("127.0.0.1"),
		Cycles:     2,
		MaxHops:    4,
		TimeoutMs:  500,
	}, enrich.New(false, false), func(_ MTRSnapshot) {
		updates++
	})
	if err != nil {
		t.Fatalf("MTR loopback failed: %v", err)
	}
	if len(snap.Hops) == 0 || updates == 0 || !snap.Completed {
		t.Fatalf("unexpected MTR snapshot: %+v (updates=%d)", snap, updates)
	}
}

func TestMatchMTRPacketTimeExceededIPv4(t *testing.T) {
	// Construct fake inner IPv4 header (20 bytes) + 8 bytes ICMP Echo Request
	inner := make([]byte, 28)
	inner[0] = 0x45 // Version 4, IHL 5 (20 bytes)
	inner[9] = 1    // Protocol ICMPv4
	binary.BigEndian.PutUint16(inner[24:26], 0x1234)
	binary.BigEndian.PutUint16(inner[26:28], 0x0205)

	msg := icmp.Message{
		Type: ipv4.ICMPTypeTimeExceeded,
		Code: 0,
		Body: &icmp.TimeExceeded{Data: inner},
	}
	wire, err := msg.Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}

	id, seq, isDest, ok := matchMTRPacket(true, 1, wire)
	if !ok || isDest || id != 0x1234 || seq != 0x0205 {
		t.Fatalf("unexpected match result: id=%#x seq=%#x isDest=%v ok=%v", id, seq, isDest, ok)
	}
}
