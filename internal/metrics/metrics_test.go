package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrometheusExpositionAndAuth(t *testing.T) {
	c := New("hub", "metrics-token-xyz", func() SnapshotGauge {
		return SnapshotGauge{OnlineNodes: 3, OfflineNodes: 1, ActiveProbes: 2}
	})
	c.IncRequest("ping")
	c.IncRequest("ping")
	c.IncGuardReject("ssrf_forbidden")
	c.AddDownloadBytes(1048576)

	// Unauthenticated should be 401
	r1 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w1 := httptest.NewRecorder()
	c.ServeHTTP(w1, r1)
	if w1.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without bearer token, got %d", w1.Code)
	}

	// Authenticated should render Prometheus text
	r2 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	r2.Header.Set("Authorization", "Bearer metrics-token-xyz")
	w2 := httptest.NewRecorder()
	c.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}

	out := w2.Body.String()
	for _, want := range []string{
		`vmlg_nodes_total{status="online"} 3`,
		`vmlg_requests_total{kind="ping"} 2`,
		`vmlg_guard_rejections_total{reason="ssrf_forbidden"} 1`,
		`vmlg_speedtest_bytes_total{direction="download"} 1048576`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing metric %q in:\n%s", want, out)
		}
	}
}
