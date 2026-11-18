package guard

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/cloudapp3/vmlg/internal/config"
)

type fakeResolver struct {
	records map[string][]netip.Addr
}

func (f fakeResolver) LookupNetIP(_ context.Context, _ string, host string) ([]netip.Addr, error) {
	addrs, ok := f.records[host]
	if !ok {
		return nil, errors.New("nxdomain")
	}
	return addrs, nil
}

func TestSSRFAndReservedProtection(t *testing.T) {
	cfg := config.Default(config.ModeStandalone).Guard
	cfg.AllowedReservedCIDRs = []string{"172.20.0.0/14"} // e.g., DN42
	cfg.BlockedCIDRs = []string{"1.1.1.1/32"}

	g, err := New(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.SetResolver(fakeResolver{
		records: map[string][]netip.Addr{
			"ok.example.com":     {netip.MustParseAddr("8.8.8.8")},
			"rebind.example.com": {netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("127.0.0.1")},
			"metadata.internal":  {netip.MustParseAddr("169.254.169.254")},
		},
	})

	blockedTargets := []string{
		"127.0.0.1",
		"::1",
		"::ffff:169.254.169.254",
		"169.254.169.254",
		"10.0.0.1",
		"192.168.1.1",
		"100.64.0.1",
		"fc00::1",
		"1.1.1.1", // explicitly blocked
		"rebind.example.com",
		"metadata.internal",
	}
	for _, target := range blockedTargets {
		if _, err := g.ResolveAndValidateTarget(context.Background(), target, "auto"); !errors.Is(err, ErrForbiddenTarget) {
			t.Errorf("target %q: expected ErrForbiddenTarget, got %v", target, err)
		}
	}

	invalidTargets := []string{
		"",
		"-c 10 8.8.8.8",
		"8.8.8.8;id",
		"http://8.8.8.8",
		"localhost",
	}
	for _, target := range invalidTargets {
		if _, err := g.ResolveAndValidateTarget(context.Background(), target, "auto"); !errors.Is(err, ErrInvalidTarget) {
			t.Errorf("target %q: expected ErrInvalidTarget, got %v", target, err)
		}
	}

	// Allowed public IP and whitelisted DN42 IP
	if _, err := g.ResolveAndValidateTarget(context.Background(), "8.8.8.8", "v4"); err != nil {
		t.Fatalf("expected 8.8.8.8 allowed, got %v", err)
	}
	if _, err := g.ResolveAndValidateTarget(context.Background(), "172.20.10.1", "v4"); err != nil {
		t.Fatalf("expected whitelisted DN42 IP allowed, got %v", err)
	}
	if _, err := g.ResolveAndValidateTarget(context.Background(), "ok.example.com", "v4"); err != nil {
		t.Fatalf("expected ok.example.com allowed, got %v", err)
	}
}

func TestPerTargetSubnetAmplificationLimit(t *testing.T) {
	cfg := config.Default(config.ModeStandalone).Guard
	cfg.ClientRPM = 600
	cfg.ClientBurst = 100
	cfg.TargetRPM = 60
	cfg.TargetBurst = 2
	cfg.MaxConcurrentProbes = 10

	g, err := New(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Three distinct attacker client IPs hitting the same victim /24 subnet!
	c1 := netip.MustParseAddr("93.184.216.1")
	c2 := netip.MustParseAddr("93.184.216.2")
	c3 := netip.MustParseAddr("93.184.216.3")

	rel1, err := g.AcquireProbe(c1, netip.MustParseAddr("8.8.4.10"))
	if err != nil {
		t.Fatalf("probe 1 failed: %v", err)
	}
	defer rel1()

	rel2, err := g.AcquireProbe(c2, netip.MustParseAddr("8.8.4.99"))
	if err != nil {
		t.Fatalf("probe 2 failed: %v", err)
	}
	defer rel2()

	if _, err := g.AcquireProbe(c3, netip.MustParseAddr("8.8.4.200")); !errors.Is(err, ErrTargetRateLimited) {
		t.Fatalf("expected 3rd probe into same /24 victim subnet to hit ErrTargetRateLimited, got %v", err)
	}
}

func TestTrustedProxyExtractionAndRedaction(t *testing.T) {
	cfg := config.Default(config.ModeStandalone).Guard
	g, err := New(cfg, []string{"127.0.0.1/32", "10.0.0.0/8"})
	if err != nil {
		t.Fatal(err)
	}

	// Untrusted peer spoofing X-Forwarded-For should be ignored
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "1.2.3.4:54321"
	req1.Header.Set("X-Forwarded-For", "9.9.9.9")
	if got := g.ClientIP(req1).String(); got != "1.2.3.4" {
		t.Fatalf("expected untrusted peer IP 1.2.3.4, got %s", got)
	}

	// Trusted proxy peer should extract rightmost untrusted XFF hop
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "127.0.0.1:43210"
	req2.Header.Set("X-Forwarded-For", "6.6.6.6, 8.8.8.8, 10.0.0.5")
	if got := g.ClientIP(req2).String(); got != "8.8.8.8" {
		t.Fatalf("expected 8.8.8.8 from XFF chain, got %s", got)
	}

	if got := RedactIP("8.8.8.8"); got != "8.8.8.x" {
		t.Fatalf("expected 8.8.8.x, got %s", got)
	}
	if got := RedactIP("2400:3200:baba::1"); got != "2400:3200:baba::x" {
		t.Fatalf("expected 2400:3200:baba::x, got %s", got)
	}
}
