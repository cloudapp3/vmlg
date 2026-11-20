package probe

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

var errPolicyDenied = errors.New("policy denied")

func TestParseHTTPURL(t *testing.T) {
	for _, bad := range []string{"", "ftp://example.com/x", "file:///etc/passwd", "http://", "https://user:pass@example.com/", "http://exa mple.com/"} {
		if _, err := ParseHTTPURL(bad); err == nil {
			t.Errorf("expected rejection for %q", bad)
		}
	}
	for _, ok := range []string{"http://example.com/", "https://example.com/path?q=1", "http://1.1.1.1:8080/x"} {
		if _, err := ParseHTTPURL(ok); err != nil {
			t.Errorf("expected accept for %q: %v", ok, err)
		}
	}
}

func TestRunHTTPBasics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Probe-Marker", "yes")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.Repeat("v", 100)))
	}))
	defer srv.Close()

	res, err := RunHTTP(context.Background(), HTTPRequest{URL: srv.URL, TimeoutMs: 5000}, nil)
	if err != nil {
		t.Fatalf("RunHTTP: %v", err)
	}
	if res.StatusCode != 200 || res.Status != "200 OK" {
		t.Fatalf("status = %d %q", res.StatusCode, res.Status)
	}
	if res.TTFBMs <= 0 || res.TotalMs <= 0 || res.TotalMs < res.TTFBMs {
		t.Fatalf("timing = ttfb %v total %v", res.TTFBMs, res.TotalMs)
	}
	if got := res.Headers.Get("X-Probe-Marker"); got != "yes" {
		t.Fatalf("X-Probe-Marker = %q", got)
	}
	if res.BytesRead != 100 || res.Truncated {
		t.Fatalf("bytes_read = %d truncated = %v", res.BytesRead, res.Truncated)
	}
	if !strings.Contains(res.RemoteAddr, "127.0.0.1") {
		t.Fatalf("remote_addr = %q", res.RemoteAddr)
	}
	if len(res.Redirects) != 0 {
		t.Fatalf("unexpected redirects: %+v", res.Redirects)
	}
	if res.TLS != nil {
		t.Fatalf("plain http response should not carry TLS summary")
	}
}

func TestRunHTTPRedirectChain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a":
			w.Header().Set("Location", "/b")
			w.WriteHeader(http.StatusMovedPermanently)
		case "/b":
			w.Header().Set("Location", "/c")
			w.WriteHeader(http.StatusFound)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	res, err := RunHTTP(context.Background(), HTTPRequest{URL: srv.URL + "/a", TimeoutMs: 5000}, nil)
	if err != nil {
		t.Fatalf("RunHTTP: %v", err)
	}
	if res.StatusCode != 200 || !strings.HasSuffix(res.URL, "/c") {
		t.Fatalf("final = %d %s", res.StatusCode, res.URL)
	}
	if len(res.Redirects) != 2 {
		t.Fatalf("redirects = %+v", res.Redirects)
	}
	if res.Redirects[0].StatusCode != 301 || res.Redirects[1].StatusCode != 302 {
		t.Fatalf("redirect codes = %+v", res.Redirects)
	}
	if !strings.HasSuffix(res.Redirects[0].Location, "/b") || !strings.HasSuffix(res.Redirects[1].Location, "/c") {
		t.Fatalf("redirect locations = %+v", res.Redirects)
	}
	if res.Redirects[0].ElapsedMs <= 0 {
		t.Fatalf("redirect elapsed = %v", res.Redirects[0].ElapsedMs)
	}
}

func TestRunHTTPTruncation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(make([]byte, 2*MaxHTTPBodyBytes))
	}))
	defer srv.Close()

	res, err := RunHTTP(context.Background(), HTTPRequest{URL: srv.URL, TimeoutMs: 8000}, nil)
	if err != nil {
		t.Fatalf("RunHTTP: %v", err)
	}
	if !res.Truncated || res.BytesRead != MaxHTTPBodyBytes {
		t.Fatalf("truncated = %v bytes_read = %d (want 1MiB cap)", res.Truncated, res.BytesRead)
	}
}

func TestRunHTTPTLSSummary(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	req := HTTPRequest{URL: srv.URL, TimeoutMs: 5000}
	req.insecureTLS = true // test-only toggle; production probes always verify
	res, err := RunHTTP(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("RunHTTP tls: %v", err)
	}
	if res.TLS == nil {
		t.Fatalf("expected TLS summary on https response")
	}
	if res.TLS.Issuer == "" || res.TLS.NotAfter <= res.TLS.NotBefore {
		t.Fatalf("tls summary = %+v", res.TLS)
	}
	foundSAN := false
	for _, s := range res.TLS.SANs {
		if s == "127.0.0.1" {
			foundSAN = true
		}
	}
	if !foundSAN {
		t.Fatalf("SANs = %v, want 127.0.0.1 present", res.TLS.SANs)
	}
}

func TestRunHTTPPolicyRejectsLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	reject := func(netip.Addr) error { return errPolicyDenied }
	if _, err := RunHTTP(context.Background(), HTTPRequest{URL: srv.URL, TimeoutMs: 5000}, reject); err == nil ||
		!strings.Contains(err.Error(), "policy denied") {
		t.Fatalf("expected policy rejection, got %v", err)
	}
}
