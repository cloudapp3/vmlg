package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cloudapp3/vmlg/internal/bgp"
	"github.com/cloudapp3/vmlg/internal/cluster"
	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
	"github.com/cloudapp3/vmlg/internal/guard"
)

func TestStandaloneWebMetaAndSSRFBlock(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "sgp-01"
	cfg.Node.Name = "Singapore 01"
	cfg.BGP.Enabled = true
	cfg.BGP.Driver = "mock"
	cfg.BGP.RouterASN = "AS64512"

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// 1. Embedded index.html serves cleanly
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "vmlg Looking Glass") {
		t.Fatalf("unexpected index.html response: status=%d", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Fatalf("expected X-Frame-Options SAMEORIGIN on normal view, got %q", got)
	}

	embedResp, err := http.Get(ts.URL + "/?embed=1")
	if err != nil {
		t.Fatal(err)
	}
	_ = embedResp.Body.Close()
	if got := embedResp.Header.Get("X-Frame-Options"); got != "" {
		t.Fatalf("expected X-Frame-Options omitted on ?embed=1 widget view, got %q", got)
	}

	// 2. /api/v1/meta returns node list
	metaResp, err := http.Get(ts.URL + "/api/v1/meta")
	if err != nil {
		t.Fatal(err)
	}
	var meta map[string]any
	_ = json.NewDecoder(metaResp.Body).Decode(&meta)
	_ = metaResp.Body.Close()
	if meta["mode"] != "standalone" {
		t.Fatalf("expected standalone mode, got %v", meta["mode"])
	}

	// 3. SSRF against 169.254.169.254 is blocked in SSE error_msg
	ssrfResp, err := http.Get(ts.URL + "/api/v1/probe/ping?target=169.254.169.254")
	if err != nil {
		t.Fatal(err)
	}
	sseBytes, _ := io.ReadAll(ssrfResp.Body)
	_ = ssrfResp.Body.Close()
	if !strings.Contains(string(sseBytes), "event: error_msg") || !strings.Contains(string(sseBytes), "reserved, private, or blocked") {
		t.Fatalf("expected SSRF block event, got: %s", string(sseBytes))
	}

	// 4. BGP Route lookup & /metrics
	bgpResp, err := http.Get(ts.URL + "/api/v1/bgp/route?target=1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	bgpBody, _ := io.ReadAll(bgpResp.Body)
	_ = bgpResp.Body.Close()
	if bgpResp.StatusCode != http.StatusOK || !strings.Contains(string(bgpBody), `"AS4809"`) {
		t.Fatalf("unexpected BGP route response: status=%d body=%s", bgpResp.StatusCode, string(bgpBody))
	}

	mResp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	mBody, _ := io.ReadAll(mResp.Body)
	_ = mResp.Body.Close()
	if mResp.StatusCode != http.StatusOK || !strings.Contains(string(mBody), `vmlg_guard_rejections_total{reason="ssrf_forbidden"} 1`) {
		t.Fatalf("unexpected /metrics output:\n%s", string(mBody))
	}
}

func TestMultiNodeCompareBroadcastSSE(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Hub server with local node + 1 remote Agent node
	hubCfg := config.Default(config.ModeHub)
	hubCfg.Node.ID = "hkg-hub"
	hubCfg.Node.Name = "Hong Kong Hub"
	hubCfg.Hub.AgentToken = "test-cluster-secret"
	hubCfg.Hub.EnableLocalNode = true
	hubCfg.Guard.AllowPrivateTargets = true
	if err := hubCfg.NormalizeAndValidate(); err != nil {
		t.Fatal(err)
	}

	srv, err := New(hubCfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// Connect remote Agent "lax-01"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	agentCfg := config.Default(config.ModeAgent)
	agentCfg.Node.ID = "lax-01"
	agentCfg.Node.Name = "Los Angeles 01"
	agentCfg.Agent.HubURL = ts.URL
	agentCfg.Agent.Token = "test-cluster-secret"
	agentCfg.Guard.AllowPrivateTargets = true
	g, err := guard.New(agentCfg.Guard, nil)
	if err != nil {
		t.Fatal(err)
	}
	ln := cluster.NewLocalNode(agentCfg, g, true, enrich.New(false, false), nil, bgp.NewDriver(agentCfg.BGP))
	agent := cluster.NewAgentClient(agentCfg, ln, g, silentLog)
	go func() { _ = agent.Run(ctx) }()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if len(srv.Registry().OnlineRunners(nil)) == 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	// Local TCP target for deterministic fast compare
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

	cmpURL := ts.URL + "/api/v1/probe/compare?target=127.0.0.1&protocol=tcp&port=" + strconv.Itoa(port) + "&include_route=true"
	resp, err := http.Get(cmpURL)
	if err != nil {
		t.Fatal(err)
	}
	rawSSE, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	out := string(rawSSE)

	if !strings.Contains(out, `"node_id":"hkg-hub"`) || !strings.Contains(out, `"node_id":"lax-01"`) || !strings.Contains(out, "event: complete") {
		t.Fatalf("expected multi-node compare events from both hkg-hub and lax-01, got:\n%s", out)
	}
}
