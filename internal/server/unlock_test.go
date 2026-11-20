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

func TestUnlockDisabledByDefault(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone) // zero-config: unlock off
	cfg.Node.ID = "unlock-off"

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/unlock")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 disabled, got %d body=%s", resp.StatusCode, string(body))
	}
	var errDoc map[string]string
	if json.Unmarshal(body, &errDoc) != nil || !strings.Contains(errDoc["error"], "disabled") {
		t.Fatalf("expected explicit disabled error, got %s", string(body))
	}

	metaResp, err := http.Get(ts.URL + "/api/v1/meta")
	if err != nil {
		t.Fatal(err)
	}
	metaBody, _ := io.ReadAll(metaResp.Body)
	_ = metaResp.Body.Close()
	if !strings.Contains(string(metaBody), `"unlock_enabled":false`) {
		t.Fatalf("meta must advertise unlock_enabled=false: %s", string(metaBody))
	}
}

func TestUnlockEnabledCacheAndRateLimit(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "unlock-on"
	cfg.Unlock.Enabled = true
	cfg.Unlock.TimeoutMs = 1500 // keep the sandbox egress attempt bounded
	cfg.Guard.UnlockRPM = 2
	cfg.Guard.UnlockBurst = 2

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	fetchUnlock := func() (int, map[string]any) {
		resp, err := http.Get(ts.URL + "/api/v1/unlock?node=unlock-on")
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		var doc map[string]any
		if err := json.Unmarshal(body, &doc); err != nil {
			t.Fatalf("bad json: %v raw=%s", err, string(body))
		}
		return resp.StatusCode, doc
	}

	// First run: executes the five checks (verdicts may be error-status in a
	// sandbox without egress, but the structure must be the fixed five).
	status, first := fetchUnlock()
	if status != http.StatusOK {
		t.Fatalf("first run status = %d doc=%v", status, first)
	}
	results := first["results"].([]any)
	if len(results) != 5 {
		t.Fatalf("expected 5 fixed services, got %d", len(results))
	}
	services := map[string]bool{}
	for _, r := range results {
		services[r.(map[string]any)["service"].(string)] = true
	}
	for _, want := range []string{"youtube", "netflix", "disney", "chatgpt", "tiktok"} {
		if !services[want] {
			t.Fatalf("service %q missing: %v", want, services)
		}
	}
	if first["cached"].(bool) {
		t.Fatalf("first run must not be cached")
	}
	checkedAt := int64(first["checked_at"].(float64))

	// Second run within TTL: served from cache (identical checked_at proves no
	// re-execution — the "upstream hit count must not grow" semantics).
	status, second := fetchUnlock()
	if status != http.StatusOK || !second["cached"].(bool) {
		t.Fatalf("second run should hit cache: status=%d doc=%v", status, second)
	}
	if int64(second["checked_at"].(float64)) != checkedAt {
		t.Fatalf("cached checked_at changed: %v vs %v", second["checked_at"], checkedAt)
	}
	if age, ok := second["cache_age_seconds"].(float64); !ok || age < 0 {
		t.Fatalf("cache_age_seconds missing: %v", second["cache_age_seconds"])
	}

	// Third run: stricter-than-probe bucket (RPM=2/Burst=2) is exhausted → 429.
	status, third := fetchUnlock()
	if status != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exhausting unlock bucket, got %d doc=%v", status, third)
	}
}
