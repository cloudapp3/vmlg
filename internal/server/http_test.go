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

func TestHTTPProbeEndpointSuccess(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hop" {
			w.Header().Set("Location", "/final")
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
		w.Header().Set("X-Backend", "vmlg-test")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("this body must never appear in probe results"))
	}))
	defer backend.Close()

	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "http-01"
	cfg.Guard.AllowedReservedCIDRs = []string{"127.0.0.0/8"}

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/probe/http?url=" + backend.URL + "/hop")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d body=%s", resp.StatusCode, string(raw))
	}

	var res map[string]any
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatal(err)
	}
	if res["status_code"].(float64) != 200 {
		t.Fatalf("status_code = %v", res["status_code"])
	}
	if redirects, ok := res["redirects"].([]any); !ok || len(redirects) != 1 {
		t.Fatalf("redirects = %v", res["redirects"])
	}
	hdrs, _ := res["headers"].(map[string]any)
	if hdrs["X-Backend"] == nil {
		t.Fatalf("headers missing X-Backend: %v", res["headers"])
	}
	// The response body must NEVER leak into the probe result.
	if body, _ := res["body"].(string); body != "" {
		t.Fatalf("body field present in result: %q", body)
	}
	if string(raw) != "" && strings.Contains(string(raw), "this body must never appear") {
		t.Fatalf("response body leaked into probe result")
	}
	if res["ttfb_ms"].(float64) <= 0 {
		t.Fatalf("ttfb_ms = %v", res["ttfb_ms"])
	}
}

func TestHTTPProbeEndpointGuardRejections(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "http-02"

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	cases := []struct {
		name   string
		query  string
		status int
	}{
		{"loopback target blocked", "url=http://127.0.0.1:8080/x", http.StatusBadRequest},
		{"localhost resolves to loopback", "url=http://localhost/x", http.StatusBadRequest},
		{"metadata host blocked", "url=http://169.254.169.254/latest/meta-data", http.StatusBadRequest},
		{"file scheme rejected", "url=file:///etc/passwd", http.StatusBadRequest},
		{"ftp scheme rejected", "url=ftp://example.com/f", http.StatusBadRequest},
		{"missing url rejected", "node=x", http.StatusBadRequest},
	}
	for _, tc := range cases {
		resp, err := http.Get(ts.URL + "/api/v1/probe/http?" + tc.query)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode != tc.status {
			t.Fatalf("%s: status = %d (want %d), body=%s", tc.name, resp.StatusCode, tc.status, string(body))
		}
		var errDoc map[string]string
		if json.Unmarshal(body, &errDoc) != nil || errDoc["error"] == "" {
			t.Fatalf("%s: expected JSON error body, got %s", tc.name, string(body))
		}
	}
}

func TestHTTPProbeEndpointClientRateLimit(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "http-03"
	cfg.Guard.ClientRPM = 1
	cfg.Guard.ClientBurst = 1
	cfg.Guard.TargetRPM = 1
	cfg.Guard.TargetBurst = 1
	cfg.Guard.AllowedReservedCIDRs = []string{"127.0.0.0/8"}

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	first, err := http.Get(ts.URL + "/api/v1/probe/http?url=" + backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = first.Body.Close()

	second, err := http.Get(ts.URL + "/api/v1/probe/http?url=" + backend.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(second.Body)
	_ = second.Body.Close()
	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on hammering, got %d body=%s", second.StatusCode, string(body))
	}
}
