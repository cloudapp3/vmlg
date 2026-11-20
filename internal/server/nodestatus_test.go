package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudapp3/vmlg/internal/config"
)

func TestNodeStatusPanelEndpoint(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "status-01"

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/nodes")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	var doc struct {
		Nodes []map[string]any `json:"nodes"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil || len(doc.Nodes) == 0 {
		t.Fatalf("nodes decode failed: %v raw=%s", err, string(raw))
	}
	node := doc.Nodes[0]
	for _, key := range []string{"online", "version", "loadavg_1", "loadavg_5", "loadavg_15", "mem_used_percent", "uptime_seconds"} {
		if _, ok := node[key]; !ok {
			t.Fatalf("node status field %q missing: %v", key, node)
		}
	}
	if node["version"].(string) == "" {
		t.Fatalf("version should be non-empty")
	}

	// Prometheus: per-node liveness gauge.
	mResp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	mBody, _ := io.ReadAll(mResp.Body)
	_ = mResp.Body.Close()
	want := `vmlg_node_up{node_id="status-01"} 1`
	if !strings.Contains(string(mBody), want) {
		t.Fatalf("missing %q in /metrics:\n%s", want, string(mBody))
	}
}

func TestNodeStatusPrivacyToggle(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "status-02"
	cfg.NodeStatus.Public = false

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/nodes")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	body := string(raw)
	for _, banned := range []string{"loadavg_1", "loadavg_5", "loadavg_15", "mem_used_percent", "uptime_seconds", "active_probes"} {
		if strings.Contains(body, banned) {
			t.Fatalf("privacy mode leaked %q: %s", banned, body)
		}
	}
	if !strings.Contains(body, `"online":true`) || !strings.Contains(body, `"version"`) {
		t.Fatalf("privacy mode must keep online flag + version: %s", body)
	}

	// Liveness gauge stays available to the operator in /metrics.
	mResp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	mBody, _ := io.ReadAll(mResp.Body)
	_ = mResp.Body.Close()
	if !strings.Contains(string(mBody), `vmlg_node_up{node_id="status-02"} 1`) {
		t.Fatalf("vmlg_node_up missing in privacy mode metrics:\n%s", string(mBody))
	}
}
