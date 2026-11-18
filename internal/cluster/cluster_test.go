package cluster

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/cloudapp3/vmlg/internal/bgp"
	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
	"github.com/cloudapp3/vmlg/internal/guard"
	"github.com/cloudapp3/vmlg/internal/probe"
)

func TestHubAgentTunnelEndToEnd(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	reg := NewRegistry("shared-cluster-token", "hmac-sign-key", 5*time.Minute, nil, silentLog)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/cluster/ws", reg.ServeAgentWS)
	hubSrv := httptest.NewServer(mux)
	defer hubSrv.Close()

	// Spin up a local TCP target so remote Agent can probe it deterministically
	tcpLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer tcpLn.Close()
	go func() {
		for {
			c, err := tcpLn.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()
	_, portStr, _ := net.SplitHostPort(tcpLn.Addr().String())
	port, _ := strconv.Atoi(portStr)

	// Start two simulated regional Agents: HKG and TYO
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, node := range []struct{ id, name, region string }{
		{"hkg-01", "Hong Kong 01", "APAC"},
		{"tyo-01", "Tokyo 01", "APAC"},
	} {
		acfg := config.Default(config.ModeAgent)
		acfg.Node.ID = node.id
		acfg.Node.Name = node.name
		acfg.Node.Region = node.region
		acfg.Agent.HubURL = hubSrv.URL
		acfg.Agent.Token = "shared-cluster-token"
		acfg.Guard.AllowPrivateTargets = true
		acfg.BGP.Enabled = true
		acfg.BGP.Driver = "mock"
		acfg.BGP.RouterASN = "AS64512"

		g, err := guard.New(acfg.Guard, nil)
		if err != nil {
			t.Fatal(err)
		}
		ln := NewLocalNode(acfg, g, true, enrich.New(false, false), nil, bgp.NewDriver(acfg.BGP))
		client := NewAgentClient(acfg, ln, g, silentLog)
		go func() {
			_ = client.Run(ctx)
		}()
	}

	// Wait until both agents register
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(reg.OnlineRunners(nil)) == 2 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	if got := len(reg.OnlineRunners(nil)); got != 2 {
		t.Fatalf("expected 2 online agents, got %d", got)
	}

	runner, err := reg.GetRunner("hkg-01")
	if err != nil {
		t.Fatalf("GetRunner(hkg-01) failed: %v", err)
	}

	var samples int
	sum, err := runner.Ping(ctx, "8.8.8.8", probe.PingRequest{
		Target:     "127.0.0.1",
		TargetAddr: netip.MustParseAddr("127.0.0.1"),
		Protocol:   "tcp",
		Port:       port,
		Count:      3,
		TimeoutMs:  500,
		IntervalMs: 15,
	}, func(_ probe.PingSample) {
		samples++
	})
	if err != nil {
		t.Fatalf("remote tunneled Ping failed: %v", err)
	}
	if sum.Recv != 3 || samples != 3 {
		t.Fatalf("unexpected tunneled Ping result: %+v (samples=%d)", sum, samples)
	}

	bgpRes, err := runner.BGPRoute(ctx, "1.1.1.1")
	if err != nil || len(bgpRes.Paths) == 0 {
		t.Fatalf("remote tunneled BGPRoute failed: %+v err=%v", bgpRes, err)
	}
}
