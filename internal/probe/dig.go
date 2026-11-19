package probe

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// digRRTypes is the fixed v1 whitelist of queryable record types.
var digRRTypes = map[string]uint16{
	"A":     dns.TypeA,
	"AAAA":  dns.TypeAAAA,
	"CNAME": dns.TypeCNAME,
	"MX":    dns.TypeMX,
	"NS":    dns.TypeNS,
	"TXT":   dns.TypeTXT,
	"PTR":   dns.TypePTR,
}

// NormalizeDigRRType validates and canonicalizes a record type string (default "A").
func NormalizeDigRRType(raw string) (string, uint16, error) {
	t := strings.ToUpper(strings.TrimSpace(raw))
	if t == "" {
		t = "A"
	}
	code, ok := digRRTypes[t]
	if !ok {
		return "", 0, fmt.Errorf("unsupported record type %q (allowed: A, AAAA, CNAME, MX, NS, TXT, PTR)", t)
	}
	return t, code, nil
}

// SystemResolverAddr returns the first nameserver from /etc/resolv.conf,
// falling back to 127.0.0.1 when unreadable or empty.
func SystemResolverAddr() string {
	if cc, err := dns.ClientConfigFromFile("/etc/resolv.conf"); err == nil {
		for _, s := range cc.Servers {
			if s != "" {
				return net.JoinHostPort(s, "53")
			}
		}
	}
	return "127.0.0.1:53"
}

// DigRateTargetAddr derives the per-target rate-limit key address for a dig
// request: the explicit resolver when given, otherwise the node's system
// resolver. Guard policy is NOT applied here (the HTTP layer owns SSRF checks).
func DigRateTargetAddr(req DigRequest) netip.Addr {
	if host := resolverHost(req.Resolver); host != "" {
		if ip, err := netip.ParseAddr(host); err == nil {
			return ip.Unmap()
		}
	}
	if host := resolverHost(SystemResolverAddr()); host != "" {
		if ip, err := netip.ParseAddr(host); err == nil {
			return ip.Unmap()
		}
	}
	return netip.MustParseAddr("127.0.0.1")
}

// resolverHost extracts the host portion of "host" / "host:port" / "[v6]:port".
func resolverHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if ip, err := netip.ParseAddr(raw); err == nil {
		return ip.Unmap().String()
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		return host
	}
	return raw
}

// NormalizeDigResolver canonicalizes a user-supplied resolver into "ip:port"
// form. Only IP literals are accepted (no hostname resolvers).
func NormalizeDigResolver(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	ip, err := netip.ParseAddr(resolverHost(raw))
	if err != nil {
		return "", fmt.Errorf("resolver must be an IPv4/IPv6 literal, got %q", raw)
	}
	ip = ip.Unmap()
	if _, port, err := net.SplitHostPort(raw); err == nil {
		if p, perr := strconv.Atoi(port); perr != nil || p < 1 || p > 65535 {
			return "", fmt.Errorf("resolver port must be 1-65535, got %q", port)
		}
		return net.JoinHostPort(ip.String(), port), nil
	}
	return net.JoinHostPort(ip.String(), "53"), nil
}

// RunDig performs a single DNS query with full dig semantics: the complete
// answer section (including CNAME chains) is returned with per-record TTLs.
// UDP is used first with automatic TCP retry on truncation. RCODE failures
// (NXDOMAIN, SERVFAIL, ...) are returned as a normal result carrying RCode;
// only transport-level failures return an error.
func RunDig(ctx context.Context, req DigRequest) (DigResult, error) {
	rrName, rrCode, err := NormalizeDigRRType(req.RRType)
	if err != nil {
		return DigResult{}, err
	}

	name := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(req.Name), "."))
	if name == "" {
		return DigResult{}, fmt.Errorf("query name is required")
	}
	if rrName == "PTR" {
		// Accept a bare IP and reverse it; explicit *.in-addr.arpa names pass through.
		if ip, perr := netip.ParseAddr(name); perr == nil {
			if rev, rerr := dns.ReverseAddr(ip.Unmap().String()); rerr == nil {
				name = strings.TrimSuffix(strings.ToLower(rev), ".")
			}
		}
	}

	resolver, err := NormalizeDigResolver(req.Resolver)
	if err != nil {
		return DigResult{}, err
	}
	if resolver == "" {
		resolver = SystemResolverAddr()
	}

	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), rrCode)
	m.RecursionDesired = true

	transport := "udp"
	start := time.Now()
	udpClient := &dns.Client{Net: "udp", Timeout: timeout}
	reply, _, err := udpClient.ExchangeContext(ctx, m, resolver)
	if err == nil && reply.Truncated {
		transport = "tcp"
		tcpClient := &dns.Client{Net: "tcp", Timeout: timeout}
		reply, _, err = tcpClient.ExchangeContext(ctx, m, resolver)
	}
	elapsed := time.Since(start)
	if err != nil {
		return DigResult{}, fmt.Errorf("dns query via %s failed: %w", resolver, err)
	}

	res := DigResult{
		Name:      name,
		RRType:    rrName,
		Resolver:  resolver,
		Transport: transport,
		ElapsedMs: float64(elapsed.Microseconds()) / 1000.0,
		RCode:     dns.RcodeToString[reply.Rcode],
		AD:        reply.AuthenticatedData,
		Truncated: reply.Truncated,
		Answers:   make([]DigRecord, 0, len(reply.Answer)),
	}
	for _, rr := range reply.Answer {
		if rec, ok := digRecordFromRR(rr); ok {
			res.Answers = append(res.Answers, rec)
		}
	}
	return res, nil
}

func digRecordFromRR(rr dns.RR) (DigRecord, bool) {
	h := rr.Header()
	rec := DigRecord{
		Name: strings.TrimSuffix(strings.ToLower(h.Name), "."),
		TTL:  h.Ttl,
		Type: dns.TypeToString[h.Rrtype],
	}
	switch v := rr.(type) {
	case *dns.A:
		rec.Value = v.A.String()
	case *dns.AAAA:
		rec.Value = v.AAAA.String()
	case *dns.CNAME:
		rec.Value = strings.TrimSuffix(v.Target, ".")
	case *dns.NS:
		rec.Value = strings.TrimSuffix(v.Ns, ".")
	case *dns.PTR:
		rec.Value = strings.TrimSuffix(v.Ptr, ".")
	case *dns.MX:
		rec.Value = fmt.Sprintf("%d %s", v.Preference, strings.TrimSuffix(v.Mx, "."))
	case *dns.TXT:
		rec.Value = strings.Join(v.Txt, "")
	default:
		return DigRecord{}, false
	}
	return rec, true
}
