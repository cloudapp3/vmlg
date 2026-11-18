package bgp

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/cloudapp3/vmlg/internal/config"
)

type mockDriver struct {
	cfg config.BGPConfig
}

func newMockDriver(cfg config.BGPConfig) *mockDriver {
	return &mockDriver{cfg: cfg}
}

func (d *mockDriver) Name() string              { return "mock" }
func (d *mockDriver) Available() (bool, string) { return true, "" }

func (d *mockDriver) LookupRoute(_ context.Context, target string) (RouteLookupResult, error) {
	norm, isV6, err := ParseQueryTarget(target)
	if err != nil {
		return RouteLookupResult{}, err
	}

	routerASN := normalizeRouterASN(d.cfg.RouterASN)
	if routerASN == "" {
		routerASN = "AS64512"
	}

	prefix := norm
	if ip, perr := netip.ParseAddr(norm); perr == nil {
		bits := 24
		if isV6 {
			bits = 48
		}
		if p, err := ip.Prefix(bits); err == nil {
			prefix = p.Masked().String()
		}
	}

	originASN := "13335"
	if norm == "8.8.8.8" || norm == "8.8.8.0/24" {
		originASN = "15169"
	}

	paths := []RoutePath{
		{
			Best:      true,
			Protocol:  "bgp_ct_cn2",
			NextHop:   "59.43.187.1",
			Interface: "eth0",
			Since:     "14d06h",
			LocalPref: 220,
			MED:       0,
			Origin:    "IGP",
			ASPath:    buildASHops([]string{"4809", originASN}),
			Communities: []CommunityTag{
				decorateCommunity("64512:220", d.cfg.Communities),
				decorateCommunity("65535:65281", d.cfg.Communities),
			},
		},
		{
			Best:      false,
			Protocol:  "bgp_ntt_transit",
			NextHop:   "129.250.2.9",
			Interface: "eth1",
			Since:     "31d12h",
			LocalPref: 100,
			MED:       50,
			Origin:    "IGP",
			ASPath:    buildASHops([]string{"2914", "174", originASN}),
			Communities: []CommunityTag{
				decorateCommunity("2914:420", d.cfg.Communities),
			},
		},
	}

	raw := fmt.Sprintf("Table master4:\n%s unicast [bgp_ct_cn2 14d06h] * (100) [AS%si]\n\tvia 59.43.187.1 on eth0\n\tBGP.as_path: 4809 %s\n\tBGP.local_pref: 220\n", prefix, originASN, originASN)
	return RouteLookupResult{
		Target:    norm,
		Prefix:    prefix,
		Driver:    "mock",
		RouterASN: routerASN,
		Paths:     paths,
		RawOutput: raw,
	}, nil
}

func (d *mockDriver) Summary(_ context.Context) (SummaryResult, error) {
	routerASN := normalizeRouterASN(d.cfg.RouterASN)
	if routerASN == "" {
		routerASN = "AS64512"
	}
	return SummaryResult{
		Driver:    "mock",
		RouterASN: routerASN,
		Peers: []PeerSession{
			{Name: "bgp_ct_cn2", Neighbor: "59.43.187.1", ASN: "AS4809", Org: "China Telecom CN2", State: "Established", Since: "14d06h", PrefixesRcvd: 942110},
			{Name: "bgp_hkix_rs1", Neighbor: "123.255.88.1", ASN: "AS4635", Org: "HKIX Route Server", State: "Established", Since: "45d02h", PrefixesRcvd: 185420},
			{Name: "bgp_ntt_transit", Neighbor: "129.250.2.9", ASN: "AS2914", Org: "NTT Global IP Network", State: "Established", Since: "31d12h", PrefixesRcvd: 968400},
		},
	}, nil
}
