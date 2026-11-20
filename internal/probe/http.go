package probe

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

// MaxHTTPBodyBytes is the hard cap on response body downloaded for timing
// purposes. The body itself is NEVER exposed in results (XSS / open-proxy line).
const MaxHTTPBodyBytes = 1 << 20 // 1 MiB

const maxHTTPRedirects = 10

// AddrPolicyFunc validates a resolved destination address (SSRF guard hook).
type AddrPolicyFunc func(netip.Addr) error

type HTTPRequest struct {
	URL       string `json:"url"`
	TimeoutMs int    `json:"timeout_ms"`

	insecureTLS bool // test-only: skip TLS verification (never JSON-serialized)
}

type HTTPRedirect struct {
	URL        string  `json:"url"`      // hop request URL
	StatusCode int     `json:"status"`   // hop response status
	Location   string  `json:"location"` // resolved next-hop URL
	ElapsedMs  float64 `json:"elapsed_ms"`
}

type HTTPTLSCert struct {
	Issuer    string   `json:"issuer"`
	Subject   string   `json:"subject"`
	NotBefore string   `json:"not_before"`
	NotAfter  string   `json:"not_after"`
	SANs      []string `json:"sans"` // DNS names + IP addresses
}

type HTTPResult struct {
	URL        string         `json:"url"` // final URL after redirects
	StatusCode int            `json:"status_code"`
	Status     string         `json:"status"` // e.g. "200 OK"
	Proto      string         `json:"proto"`  // e.g. "HTTP/2.0"
	RemoteAddr string         `json:"remote_addr,omitempty"`
	TTFBMs     float64        `json:"ttfb_ms"` // final hop: request start → response headers
	TotalMs    float64        `json:"total_ms"`
	Redirects  []HTTPRedirect `json:"redirects,omitempty"`
	TLS        *HTTPTLSCert   `json:"tls,omitempty"`
	Headers    http.Header    `json:"headers"` // final response headers (body never included)
	BytesRead  int64          `json:"bytes_read"`
	Truncated  bool           `json:"truncated,omitempty"` // body exceeded MaxHTTPBodyBytes
}

// ParseHTTPURL validates a probe target URL: only http/https with a host.
func ParseHTTPURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 2048 {
		return nil, fmt.Errorf("url is required (max 2048 chars)")
	}
	if strings.ContainsAny(raw, " \t\r\n") {
		return nil, fmt.Errorf("url must not contain whitespace")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("only http/https urls are supported (got %q)", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("url host is required")
	}
	if u.User != nil {
		return nil, fmt.Errorf("userinfo in url is not allowed")
	}
	return u, nil
}

// HTTPRateTargetAddr resolves the URL host to a single address for the
// per-target rate-limit bucket (first resolved record, loopback fallback).
func HTTPRateTargetAddr(ctx context.Context, rawURL string) netip.Addr {
	u, err := ParseHTTPURL(rawURL)
	if err != nil {
		return netip.MustParseAddr("127.0.0.1")
	}
	host := u.Hostname()
	if ip, err := netip.ParseAddr(host); err == nil {
		return ip.Unmap()
	}
	addrs, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(addrs) == 0 {
		return netip.MustParseAddr("127.0.0.1")
	}
	return addrs[0].Unmap()
}

// RunHTTP performs a controlled GET probe: manual redirect following with
// per-hop address policy checks, 1 MiB body cap (counted, never returned),
// TTFB/total timing, and a TLS certificate summary. policy may be nil.
func RunHTTP(ctx context.Context, req HTTPRequest, policy AddrPolicyFunc) (HTTPResult, error) {
	u, err := ParseHTTPURL(req.URL)
	if err != nil {
		return HTTPResult{}, err
	}
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var dialedAddr string
	transport := &http.Transport{
		DisableKeepAlives: true,
		Proxy:             nil, // never honor environment proxies
		ForceAttemptHTTP2: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: req.insecureTLS},
		DialContext: func(dctx context.Context, network, addr string) (net.Conn, error) {
			conn, derr := policyDial(dctx, network, addr, policy)
			if derr == nil {
				dialedAddr = conn.RemoteAddr().String()
			}
			return conn, derr
		},
	}
	client := &http.Client{
		Transport:     transport,
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	start := time.Now()
	res := HTTPResult{URL: u.String(), Redirects: []HTTPRedirect{}, RemoteAddr: dialedAddr}
	current := *u

	for hop := 0; ; hop++ {
		hopStart := time.Now()
		r, rerr := client.Get(current.String())
		if rerr != nil {
			return HTTPResult{}, fmt.Errorf("http probe failed: %w", rerr)
		}

		if r.StatusCode >= 300 && r.StatusCode <= 399 && r.Header.Get("Location") != "" && hop < maxHTTPRedirects {
			next, perr := current.Parse(r.Header.Get("Location"))
			if perr != nil {
				r.Body.Close()
				return HTTPResult{}, fmt.Errorf("invalid redirect location %q: %w", r.Header.Get("Location"), perr)
			}
			if next.Scheme != "http" && next.Scheme != "https" {
				r.Body.Close()
				return HTTPResult{}, fmt.Errorf("redirect to non-http scheme %q rejected", next.Scheme)
			}
			res.Redirects = append(res.Redirects, HTTPRedirect{
				URL:        current.String(),
				StatusCode: r.StatusCode,
				Location:   next.String(),
				ElapsedMs:  msSince(hopStart),
			})
			io.Copy(io.Discard, io.LimitReader(r.Body, 4<<10)) // drain a little for the connection
			r.Body.Close()
			current = *next
			continue
		}

		// Final response
		res.TTFBMs = msSince(hopStart)
		res.StatusCode = r.StatusCode
		res.Status = r.Status
		res.Proto = r.Proto
		res.URL = current.String()
		res.RemoteAddr = dialedAddr
		res.Headers = r.Header.Clone()
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			res.TLS = tlsSummary(r.TLS.PeerCertificates[0])
		}
		n, _ := io.Copy(io.Discard, io.LimitReader(r.Body, MaxHTTPBodyBytes+1))
		r.Body.Close()
		res.BytesRead = n
		res.Truncated = n > MaxHTTPBodyBytes
		if n > MaxHTTPBodyBytes {
			res.BytesRead = MaxHTTPBodyBytes
		}
		break
	}
	res.TotalMs = msSince(start)
	return res, nil
}

// policyDial resolves the target host, applies the policy to EVERY resolved
// record (DNS-rebinding defense), then dials the first address.
func policyDial(ctx context.Context, network, addr string, policy AddrPolicyFunc) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	var candidates []netip.Addr
	if ip, perr := netip.ParseAddr(host); perr == nil {
		candidates = []netip.Addr{ip.Unmap()}
	} else {
		addrs, lerr := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		if lerr != nil {
			return nil, lerr
		}
		for _, a := range addrs {
			candidates = append(candidates, a.Unmap())
		}
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no addresses resolved for %s", host)
	}

	chosen := candidates[0]
	if policy != nil {
		for _, c := range candidates {
			if perr := policy(c); perr != nil {
				return nil, perr
			}
		}
	}

	var d net.Dialer
	return d.DialContext(ctx, network, net.JoinHostPort(chosen.String(), port))
}

func tlsSummary(cert *x509.Certificate) *HTTPTLSCert {
	sans := make([]string, 0, len(cert.DNSNames)+len(cert.IPAddresses))
	sans = append(sans, cert.DNSNames...)
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	return &HTTPTLSCert{
		Issuer:    cert.Issuer.String(),
		Subject:   cert.Subject.String(),
		NotBefore: cert.NotBefore.UTC().Format(time.RFC3339),
		NotAfter:  cert.NotAfter.UTC().Format(time.RFC3339),
		SANs:      sans,
	}
}

func msSince(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000.0
}
