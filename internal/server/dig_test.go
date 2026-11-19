package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/probe"
	"github.com/miekg/dns"
)

func startDigTestDNS(t *testing.T) string {
	t.Helper()
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	server := &dns.Server{PacketConn: conn, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		if len(r.Question) == 1 && r.Question[0].Name == "www.example.com." {
			m.Answer = []dns.RR{
				&dns.CNAME{
					Hdr:    dns.RR_Header{Name: "www.example.com.", Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: 120},
					Target: "cdn.example.com.",
				},
				&dns.A{
					Hdr: dns.RR_Header{Name: "cdn.example.com.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
					A:   net.ParseIP("93.184.216.34"),
				},
			}
		}
		_ = w.WriteMsg(m)
	})}
	go func() { _ = server.ActivateAndServe() }()
	t.Cleanup(func() { _ = server.Shutdown() })
	return conn.LocalAddr().String()
}

func TestDigEndpointSuccessAndCNAMEChain(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	dnsAddr := startDigTestDNS(t)

	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "dig-01"
	// Tests resolve against a loopback fake DNS server, so allow 127/8 for the
	// resolver target only (the queried name itself never goes through guard).
	cfg.Guard.AllowedReservedCIDRs = []string{"127.0.0.0/8"}

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/probe/dig?target=www.example.com&rrtype=A&resolver=" + dnsAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("dig status = %d, body=%s", resp.StatusCode, string(body))
	}
	var res probe.DigResult
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}
	if res.RCode != "NOERROR" || len(res.Answers) != 2 {
		t.Fatalf("dig result = %+v", res)
	}
	if res.Answers[0].Type != "CNAME" || res.Answers[0].TTL != 120 || res.Answers[1].Value != "93.184.216.34" {
		t.Fatalf("dig answers = %+v", res.Answers)
	}
	if res.Resolver != dnsAddr {
		t.Fatalf("resolver echo = %s, want %s", res.Resolver, dnsAddr)
	}
}

func TestDigEndpointGuardRejections(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "dig-02"

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
		{"private resolver blocked", "target=example.com&resolver=192.168.1.1", http.StatusBadRequest},
		{"metadata resolver blocked", "target=example.com&resolver=169.254.169.254", http.StatusBadRequest},
		{"hostname resolver rejected", "target=example.com&resolver=dns.example.com", http.StatusBadRequest},
		{"bad rrtype rejected", "target=example.com&rrtype=AXFR", http.StatusBadRequest},
		{"invalid query name rejected", "target=bad%20name%3B", http.StatusBadRequest},
	}
	for _, tc := range cases {
		resp, err := http.Get(ts.URL + "/api/v1/probe/dig?" + tc.query)
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

func TestDigEndpointClientRateLimit(t *testing.T) {
	silentLog := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "dig-03"
	cfg.Guard.ClientRPM = 1
	cfg.Guard.ClientBurst = 1
	cfg.Guard.TargetRPM = 1
	cfg.Guard.TargetBurst = 1

	srv, err := New(cfg, silentLog)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// First request consumes the single burst token (resolver 1.1.1.1 is
	// policy-allowed; its success/failure is irrelevant to the rate-limit check).
	first, err := http.Get(ts.URL + "/api/v1/probe/dig?target=example.com&resolver=1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	_ = first.Body.Close()

	second, err := http.Get(ts.URL + "/api/v1/probe/dig?target=example.com&resolver=1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(second.Body)
	_ = second.Body.Close()
	if second.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on hammering, got %d body=%s", second.StatusCode, string(body))
	}
}
