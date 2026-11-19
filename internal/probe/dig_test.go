package probe

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/miekg/dns"
)

// startFakeDNS spins a local authoritative-ish DNS server used by dig tests.
func startFakeDNS(t *testing.T, handler dns.HandlerFunc) string {
	t.Helper()
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	server := &dns.Server{PacketConn: conn, Handler: handler}
	go func() { _ = server.ActivateAndServe() }()
	t.Cleanup(func() { _ = server.Shutdown() })
	return conn.LocalAddr().String()
}

func TestRunDigCNAMEChainAndTTL(t *testing.T) {
	srv := startFakeDNS(t, func(w dns.ResponseWriter, r *dns.Msg) {
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
		} else {
			m.Rcode = dns.RcodeNameError
		}
		_ = w.WriteMsg(m)
	})

	res, err := RunDig(context.Background(), DigRequest{
		Name: "WWW.Example.COM.", RRType: "a", Resolver: srv, TimeoutMs: 2000,
	})
	if err != nil {
		t.Fatalf("RunDig: %v", err)
	}
	if res.RCode != "NOERROR" {
		t.Fatalf("rcode = %s, want NOERROR", res.RCode)
	}
	if len(res.Answers) != 2 {
		t.Fatalf("answers = %d, want 2 (CNAME chain + A)", len(res.Answers))
	}
	cname, a := res.Answers[0], res.Answers[1]
	if cname.Type != "CNAME" || cname.Value != "cdn.example.com" || cname.TTL != 120 {
		t.Errorf("cname record = %+v", cname)
	}
	if a.Type != "A" || a.Value != "93.184.216.34" || a.TTL != 60 {
		t.Errorf("a record = %+v", a)
	}
	if res.Name != "www.example.com" {
		t.Errorf("name = %s, want lowercased www.example.com", res.Name)
	}
	if res.Transport != "udp" {
		t.Errorf("transport = %s, want udp", res.Transport)
	}
	if res.ElapsedMs <= 0 {
		t.Errorf("elapsed_ms = %v, want > 0", res.ElapsedMs)
	}
	if res.Resolver != srv {
		t.Errorf("resolver = %s, want %s", res.Resolver, srv)
	}
}

func TestRunDigNXDOMAINIsResultNotError(t *testing.T) {
	srv := startFakeDNS(t, func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Rcode = dns.RcodeNameError
		_ = w.WriteMsg(m)
	})

	res, err := RunDig(context.Background(), DigRequest{Name: "nope.example.com", Resolver: srv, TimeoutMs: 2000})
	if err != nil {
		t.Fatalf("NXDOMAIN should be a result, got error: %v", err)
	}
	if res.RCode != "NXDOMAIN" || len(res.Answers) != 0 {
		t.Fatalf("res = %+v", res)
	}
}

func TestRunDigPTRReverseAndTXTMXFormatting(t *testing.T) {
	srv := startFakeDNS(t, func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		if len(r.Question) != 1 {
			return
		}
		switch r.Question[0].Name {
		case "4.3.2.1.in-addr.arpa.":
			m.Answer = []dns.RR{
				&dns.PTR{
					Hdr: dns.RR_Header{Name: "4.3.2.1.in-addr.arpa.", Rrtype: dns.TypePTR, Class: dns.ClassINET, Ttl: 300},
					Ptr: "host.example.com.",
				},
			}
		case "example.com.":
			if r.Question[0].Qtype == dns.TypeTXT {
				m.Answer = []dns.RR{
					&dns.TXT{
						Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 3600},
						Txt: []string{"v=spf1 ", "include:_spf.example.com ~all"},
					},
				}
			} else if r.Question[0].Qtype == dns.TypeMX {
				m.Answer = []dns.RR{
					&dns.MX{
						Hdr:        dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 600},
						Preference: 10,
						Mx:         "mail.example.com.",
					},
				}
			}
		}
		_ = w.WriteMsg(m)
	})

	ptr, err := RunDig(context.Background(), DigRequest{Name: "1.2.3.4", RRType: "PTR", Resolver: srv, TimeoutMs: 2000})
	if err != nil {
		t.Fatalf("PTR: %v", err)
	}
	if ptr.Name != "4.3.2.1.in-addr.arpa" || len(ptr.Answers) != 1 ||
		ptr.Answers[0].Value != "host.example.com" || ptr.Answers[0].TTL != 300 {
		t.Errorf("ptr result = %+v", ptr)
	}

	txt, err := RunDig(context.Background(), DigRequest{Name: "example.com", RRType: "TXT", Resolver: srv, TimeoutMs: 2000})
	if err != nil {
		t.Fatalf("TXT: %v", err)
	}
	if txt.Answers[0].Value != "v=spf1 include:_spf.example.com ~all" {
		t.Errorf("txt value = %q", txt.Answers[0].Value)
	}

	mx, err := RunDig(context.Background(), DigRequest{Name: "example.com", RRType: "MX", Resolver: srv, TimeoutMs: 2000})
	if err != nil {
		t.Fatalf("MX: %v", err)
	}
	if mx.Answers[0].Value != "10 mail.example.com" {
		t.Errorf("mx value = %q", mx.Answers[0].Value)
	}
}

func TestNormalizeDigRRTypeAndResolver(t *testing.T) {
	if name, code, err := NormalizeDigRRType(" aaaa "); err != nil || name != "AAAA" || code != dns.TypeAAAA {
		t.Errorf("normalize aaaa = %s %d %v", name, code, err)
	}
	if _, _, err := NormalizeDigRRType("AXFR"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("AXFR should be rejected, got %v", err)
	}
	if r, err := NormalizeDigResolver("1.1.1.1"); err != nil || r != "1.1.1.1:53" {
		t.Errorf("resolver 1.1.1.1 = %q %v", r, err)
	}
	if r, err := NormalizeDigResolver("[2001:db8::1]:5353"); err != nil || r != "[2001:db8::1]:5353" {
		t.Errorf("resolver v6+port = %q %v", r, err)
	}
	if _, err := NormalizeDigResolver("dns.example.com"); err == nil {
		t.Errorf("hostname resolver should be rejected")
	}
	if _, err := NormalizeDigResolver("1.1.1.1:0"); err == nil {
		t.Errorf("port 0 should be rejected")
	}
	if addr := DigRateTargetAddr(DigRequest{Resolver: "8.8.8.8"}); addr.String() != "8.8.8.8" {
		t.Errorf("DigRateTargetAddr = %s, want 8.8.8.8", addr)
	}
}
