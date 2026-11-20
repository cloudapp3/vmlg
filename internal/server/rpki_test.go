package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudapp3/vmlg/internal/config"
)

func bgpMockCfg(nodeID string) config.Config {
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = nodeID
	cfg.BGP.Enabled = true
	cfg.BGP.Driver = "mock"
	cfg.BGP.RouterASN = "AS64512"
	return cfg
}

func TestRPKIAbsentWhenUnconfigured(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := bgpMockCfg("rpki-off")

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/bgp/route?target=1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	if strings.Contains(string(body), `"rpki"`) {
		t.Fatalf("rpki field must be absent when routinator unconfigured: %s", string(body))
	}
}

func TestRPKIBadgeWithRoutinator(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	routinator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo-style: any /api/v1/validity/{asn}/{prefix} query is VALID with one ROA.
		if !strings.HasPrefix(r.URL.Path, "/api/v1/validity/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"validated_route":{"validity":{"state":"valid","description":"covered",` +
			`"VRPs":[{"asn":"AS15169","prefix":"1.1.1.0/24","max_length":24}]}}}`))
	}))
	defer routinator.Close()

	cfg := bgpMockCfg("rpki-on")
	cfg.RPKI.RoutinatorURL = routinator.URL
	cfg.RPKI.TimeoutMs = 2000

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/bgp/route?target=1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(body))
	}
	for _, want := range []string{`"rpki"`, `"state":"valid"`, `"max_length":24`, `"AS15169"`} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("missing %q in BGP response: %s", want, string(body))
		}
	}
}

func TestRPKIDegradesWithoutBlocking(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	// A server that is closed immediately: connection refused, fast.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	deadURL := dead.URL
	dead.Close()

	cfg := bgpMockCfg("rpki-dead")
	cfg.RPKI.RoutinatorURL = deadURL
	cfg.RPKI.TimeoutMs = 1000

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/bgp/route?target=1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	// Main BGP result must survive.
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), `"paths"`) {
		t.Fatalf("BGP main result broken by dead routinator: status=%d body=%s", resp.StatusCode, string(body))
	}
	// RPKI column degrades to unavailable.
	if !strings.Contains(string(body), `"rpki"`) || !strings.Contains(string(body), `"state":"unavailable"`) {
		t.Fatalf("expected unavailable rpki state: %s", string(body))
	}
}
