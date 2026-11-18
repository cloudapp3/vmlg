package guard

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
)

var (
	ErrInvalidTarget     = errors.New("invalid target host or IP address")
	ErrForbiddenTarget   = errors.New("target address is reserved, private, or blocked")
	ErrClientRateLimited = errors.New("client rate limit exceeded, please slow down")
	ErrTargetRateLimited = errors.New("target network rate limit exceeded")
	ErrBusy              = errors.New("node concurrency limit reached, please retry shortly")
)

var reservedPrefixes = mustPrefixes(
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"100::/64",
	"2001:db8::/32",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
)

type Resolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type ResolvedTarget struct {
	Input    string     `json:"input"`
	Hostname string     `json:"hostname,omitempty"`
	Addr     netip.Addr `json:"addr"`
}

type tokenBucket struct {
	tokens   float64
	lastSeen time.Time
}

type Guard struct {
	allowPrivate   bool
	allowedCIDRs   []netip.Prefix
	blockedCIDRs   []netip.Prefix
	trustedProxies []netip.Prefix
	resolver       Resolver

	clientRate  float64 // tokens per second
	clientBurst float64
	targetRate  float64 // tokens per second
	targetBurst float64

	mu            sync.Mutex
	clientBuckets map[string]*tokenBucket
	targetBuckets map[string]*tokenBucket
	lastCleanup   time.Time
	nowFn         func() time.Time

	probeSem     chan struct{}
	speedtestSem chan struct{}
}

func New(cfg config.GuardConfig, trustedProxyCIDRs []string) (*Guard, error) {
	allowed, err := parsePrefixes(cfg.AllowedReservedCIDRs)
	if err != nil {
		return nil, err
	}
	blocked, err := parsePrefixes(cfg.BlockedCIDRs)
	if err != nil {
		return nil, err
	}
	trusted, err := parsePrefixes(trustedProxyCIDRs)
	if err != nil {
		return nil, err
	}

	maxProbes := max(cfg.MaxConcurrentProbes, 1)
	maxSpeed := max(cfg.MaxConcurrentSpeedtest, 1)
	clientRPM := max(cfg.ClientRPM, 1)
	clientBurst := max(cfg.ClientBurst, 1)
	targetRPM := max(cfg.TargetRPM, 1)
	targetBurst := max(cfg.TargetBurst, 1)

	return &Guard{
		allowPrivate:   cfg.AllowPrivateTargets,
		allowedCIDRs:   allowed,
		blockedCIDRs:   blocked,
		trustedProxies: trusted,
		resolver:       net.DefaultResolver,
		clientRate:     float64(clientRPM) / 60.0,
		clientBurst:    float64(clientBurst),
		targetRate:     float64(targetRPM) / 60.0,
		targetBurst:    float64(targetBurst),
		clientBuckets:  make(map[string]*tokenBucket),
		targetBuckets:  make(map[string]*tokenBucket),
		lastCleanup:    time.Now(),
		nowFn:          time.Now,
		probeSem:       make(chan struct{}, maxProbes),
		speedtestSem:   make(chan struct{}, maxSpeed),
	}, nil
}

func (g *Guard) SetResolver(r Resolver) {
	if r != nil {
		g.resolver = r
	}
}

func (g *Guard) ActiveProbes() int {
	return len(g.probeSem)
}

func (g *Guard) ActiveSpeedtests() int {
	return len(g.speedtestSem)
}

// ClientIP extracts the normalized visitor IP address, honoring proxy headers
// ONLY when the immediate peer belongs to configured trusted_proxies.
func (g *Guard) ClientIP(r *http.Request) netip.Addr {
	remoteAddr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		remoteAddr = host
	}
	peerIP, err := netip.ParseAddr(strings.TrimSpace(remoteAddr))
	if err != nil {
		return netip.Addr{}
	}
	peerIP = peerIP.Unmap()

	if !g.isTrustedProxy(peerIP) {
		return peerIP
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			candidate, err := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if err != nil {
				continue
			}
			candidate = candidate.Unmap()
			if !g.isTrustedProxy(candidate) {
				return candidate
			}
		}
	}

	for _, hdr := range []string{"CF-Connecting-IP", "X-Real-IP"} {
		if raw := strings.TrimSpace(r.Header.Get(hdr)); raw != "" {
			if ip, err := netip.ParseAddr(raw); err == nil {
				return ip.Unmap()
			}
		}
	}

	return peerIP
}

func (g *Guard) isTrustedProxy(ip netip.Addr) bool {
	if !ip.IsValid() {
		return false
	}
	for _, p := range g.trustedProxies {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

// ResolveAndValidateTarget validates syntax, resolves DNS if needed, and checks
// every candidate address against SSRF, reserved range, and blocklist policies.
func (g *Guard) ResolveAndValidateTarget(ctx context.Context, rawTarget string, ipVersion string) (ResolvedTarget, error) {
	target := strings.TrimSpace(rawTarget)
	if target == "" || len(target) > 253 {
		return ResolvedTarget{}, ErrInvalidTarget
	}
	if strings.HasPrefix(target, "-") || strings.ContainsAny(target, " \t\r\n/\\?#@;&|$`\"'<>(){}[]!*") {
		return ResolvedTarget{}, ErrInvalidTarget
	}

	ipVersion = normalizeIPVersion(ipVersion)

	// Direct IP literal?
	if ip, err := netip.ParseAddr(target); err == nil {
		ip = ip.Unmap()
		if ipVersion == "v4" && !ip.Is4() {
			return ResolvedTarget{}, fmt.Errorf("%w: target is IPv6, but IPv4 was requested", ErrInvalidTarget)
		}
		if ipVersion == "v6" && !ip.Is6() {
			return ResolvedTarget{}, fmt.Errorf("%w: target is IPv4, but IPv6 was requested", ErrInvalidTarget)
		}
		if err := g.CheckAddrAllowed(ip); err != nil {
			return ResolvedTarget{}, err
		}
		return ResolvedTarget{
			Input: target,
			Addr:  ip,
		}, nil
	}

	// Validate RFC 1123 hostname
	if !isValidHostname(target) {
		return ResolvedTarget{}, ErrInvalidTarget
	}

	lookupNet := "ip"
	switch ipVersion {
	case "v4":
		lookupNet = "ip4"
	case "v6":
		lookupNet = "ip6"
	}

	resolveCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	addrs, err := g.resolver.LookupNetIP(resolveCtx, lookupNet, target)
	if err != nil || len(addrs) == 0 {
		return ResolvedTarget{}, fmt.Errorf("%w: DNS lookup failed for %s", ErrInvalidTarget, target)
	}

	// Prevent DNS rebinding: ALL returned addresses for the requested family must pass policy.
	var chosen netip.Addr
	for _, a := range addrs {
		ip := a.Unmap()
		if ipVersion == "v4" && !ip.Is4() {
			continue
		}
		if ipVersion == "v6" && !ip.Is6() {
			continue
		}
		if err := g.CheckAddrAllowed(ip); err != nil {
			return ResolvedTarget{}, err
		}
		if !chosen.IsValid() {
			chosen = ip
		}
	}
	if !chosen.IsValid() {
		return ResolvedTarget{}, fmt.Errorf("%w: no matching %s address found for %s", ErrInvalidTarget, ipVersion, target)
	}

	return ResolvedTarget{
		Input:    target,
		Hostname: strings.TrimSuffix(strings.ToLower(target), "."),
		Addr:     chosen,
	}, nil
}

func (g *Guard) CheckAddrAllowed(addr netip.Addr) error {
	ip := addr.Unmap()
	if !ip.IsValid() {
		return ErrInvalidTarget
	}

	// Explicit blocklist always wins.
	for _, blocked := range g.blockedCIDRs {
		if blocked.Contains(ip) {
			return ErrForbiddenTarget
		}
	}

	if g.allowPrivate {
		return nil
	}

	// Explicit allowlist overrides default reserved prefixes (e.g., DN42 ranges).
	for _, allowed := range g.allowedCIDRs {
		if allowed.Contains(ip) {
			return nil
		}
	}

	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return ErrForbiddenTarget
	}
	for _, reserved := range reservedPrefixes {
		if reserved.Contains(ip) {
			return ErrForbiddenTarget
		}
	}
	return nil
}

// AcquireProbe enforces Per-Client rate limit, Per-Target subnet rate limit,
// and node probe concurrency semaphore.
func (g *Guard) AcquireProbe(clientIP netip.Addr, targetIP netip.Addr) (func(), error) {
	now := g.nowFn()
	clientKey := clientBucketKey(clientIP)
	targetKey := targetBucketKey(targetIP)

	g.mu.Lock()
	g.cleanupLocked(now)
	if clientKey != "" && !allowToken(g.clientBuckets, clientKey, g.clientRate, g.clientBurst, now) {
		g.mu.Unlock()
		return nil, ErrClientRateLimited
	}
	if targetKey != "" && !allowToken(g.targetBuckets, targetKey, g.targetRate, g.targetBurst, now) {
		g.mu.Unlock()
		return nil, ErrTargetRateLimited
	}
	g.mu.Unlock()

	select {
	case g.probeSem <- struct{}{}:
		var once sync.Once
		return func() {
			once.Do(func() { <-g.probeSem })
		}, nil
	default:
		return nil, ErrBusy
	}
}

// AcquireSpeedtest enforces Per-Client rate limit and speedtest concurrency limit.
func (g *Guard) AcquireSpeedtest(clientIP netip.Addr, checkRate bool) (func(), error) {
	if checkRate {
		now := g.nowFn()
		clientKey := clientBucketKey(clientIP)
		g.mu.Lock()
		g.cleanupLocked(now)
		if clientKey != "" && !allowToken(g.clientBuckets, clientKey, g.clientRate, g.clientBurst, now) {
			g.mu.Unlock()
			return nil, ErrClientRateLimited
		}
		g.mu.Unlock()
	}

	select {
	case g.speedtestSem <- struct{}{}:
		var once sync.Once
		return func() {
			once.Do(func() { <-g.speedtestSem })
		}, nil
	default:
		return nil, ErrBusy
	}
}

func allowToken(store map[string]*tokenBucket, key string, ratePerSec, burst float64, now time.Time) bool {
	b, ok := store[key]
	if !ok {
		store[key] = &tokenBucket{
			tokens:   burst - 1,
			lastSeen: now,
		}
		return true
	}
	elapsed := now.Sub(b.lastSeen).Seconds()
	if elapsed > 0 {
		b.tokens = min(burst, b.tokens+elapsed*ratePerSec)
		b.lastSeen = now
	}
	if b.tokens < 1.0 {
		return false
	}
	b.tokens -= 1.0
	return true
}

func (g *Guard) cleanupLocked(now time.Time) {
	if now.Sub(g.lastCleanup) < 2*time.Minute && len(g.clientBuckets) < 4096 && len(g.targetBuckets) < 4096 {
		return
	}
	g.lastCleanup = now
	cutoff := now.Add(-10 * time.Minute)
	for k, b := range g.clientBuckets {
		if b.lastSeen.Before(cutoff) {
			delete(g.clientBuckets, k)
		}
	}
	for k, b := range g.targetBuckets {
		if b.lastSeen.Before(cutoff) {
			delete(g.targetBuckets, k)
		}
	}
}

func clientBucketKey(ip netip.Addr) string {
	ip = ip.Unmap()
	if !ip.IsValid() {
		return ""
	}
	if ip.Is4() {
		return ip.String()
	}
	p, err := ip.Prefix(64)
	if err != nil {
		return ip.String()
	}
	return p.Masked().String()
}

func targetBucketKey(ip netip.Addr) string {
	ip = ip.Unmap()
	if !ip.IsValid() {
		return ""
	}
	bits := 24
	if ip.Is6() {
		bits = 48
	}
	p, err := ip.Prefix(bits)
	if err != nil {
		return ip.String()
	}
	return p.Masked().String()
}

// RedactIP masks the host portion of an IPv4/IPv6 address for safe public sharing.
func RedactIP(raw string) string {
	ip, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return raw
	}
	ip = ip.Unmap()
	if ip.Is4() {
		b := ip.As4()
		return fmt.Sprintf("%d.%d.%d.x", b[0], b[1], b[2])
	}
	p, err := ip.Prefix(48)
	if err != nil {
		return "xxxx::x"
	}
	base := strings.TrimSuffix(p.Masked().Addr().String(), "::")
	return base + "::x"
}

func normalizeIPVersion(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "4", "v4", "ipv4":
		return "v4"
	case "6", "v6", "ipv6":
		return "v6"
	default:
		return "auto"
	}
}

func isValidHostname(h string) bool {
	h = strings.TrimSuffix(h, ".")
	if h == "" || len(h) > 253 || !strings.Contains(h, ".") {
		return false
	}
	labels := strings.Split(h, ".")
	allNumericTLD := true
	tld := labels[len(labels)-1]
	for _, ch := range tld {
		if ch < '0' || ch > '9' {
			allNumericTLD = false
			break
		}
	}
	if allNumericTLD {
		return false
	}
	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			isAlphaNum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
			if !isAlphaNum && c != '-' {
				return false
			}
		}
	}
	return true
}

func parsePrefixes(list []string) ([]netip.Prefix, error) {
	out := make([]netip.Prefix, 0, len(list))
	for _, raw := range list {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.Contains(raw, "/") {
			p, err := netip.ParsePrefix(raw)
			if err != nil {
				return nil, err
			}
			out = append(out, p.Masked())
		} else {
			ip, err := netip.ParseAddr(raw)
			if err != nil {
				return nil, err
			}
			ip = ip.Unmap()
			out = append(out, netip.PrefixFrom(ip, ip.BitLen()))
		}
	}
	return out, nil
}

func mustPrefixes(cidrs ...string) []netip.Prefix {
	p, err := parsePrefixes(cidrs)
	if err != nil {
		panic(err)
	}
	return p
}
