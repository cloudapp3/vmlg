package bgp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
)

type frrDriver struct {
	cfg     config.BGPConfig
	binPath string
}

func newFRRDriver(cfg config.BGPConfig) *frrDriver {
	bin := cfg.VtyshBinary
	if bin == "" {
		bin = "vtysh"
	}
	path, _ := exec.LookPath(bin)
	return &frrDriver{cfg: cfg, binPath: path}
}

func (d *frrDriver) Name() string { return "frr" }

func (d *frrDriver) Available() (bool, string) {
	if d.binPath == "" {
		return false, "vtysh binary not found"
	}
	return true, ""
}

func (d *frrDriver) LookupRoute(ctx context.Context, target string) (RouteLookupResult, error) {
	norm, isV6, err := ParseQueryTarget(target)
	if err != nil {
		return RouteLookupResult{}, err
	}
	if d.binPath == "" {
		return RouteLookupResult{}, fmt.Errorf("vtysh binary not found")
	}

	afi := "ipv4"
	if isV6 {
		afi = "ipv6"
	}
	vtyCmd := fmt.Sprintf("show bgp %s unicast %s json", afi, norm)

	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	out, err := exec.CommandContext(cctx, d.binPath, "-c", vtyCmd).Output()
	if err != nil {
		return RouteLookupResult{}, fmt.Errorf("vtysh query failed: %w", err)
	}

	return parseFRRRouteJSON(norm, d.cfg.RouterASN, d.cfg.Communities, out)
}

func parseFRRRouteJSON(target, routerASN string, customComm map[string]string, rawJSON []byte) (RouteLookupResult, error) {
	var doc struct {
		Prefix string `json:"prefix"`
		Paths  []struct {
			LocalPref int    `json:"locPrf"`
			Metric    int    `json:"metric"`
			Origin    string `json:"origin"`
			Valid     bool   `json:"valid"`
			Bestpath  *struct {
				Overall bool `json:"overall"`
			} `json:"bestpath"`
			ASPath *struct {
				String string `json:"string"`
			} `json:"aspath"`
			Nexthops []struct {
				IP            string `json:"ip"`
				InterfaceName string `json:"interfaceName"`
			} `json:"nexthops"`
			Community *struct {
				List []string `json:"list"`
			} `json:"community"`
			LargeCommunity *struct {
				String string `json:"string"`
			} `json:"largeCommunity"`
		} `json:"paths"`
	}

	res := RouteLookupResult{
		Target:    target,
		Driver:    "frr",
		RouterASN: normalizeRouterASN(routerASN),
		RawOutput: string(rawJSON),
	}

	if err := json.Unmarshal(rawJSON, &doc); err != nil {
		return res, nil
	}
	res.Prefix = doc.Prefix

	for _, p := range doc.Paths {
		rp := RoutePath{
			Best:      p.Bestpath != nil && p.Bestpath.Overall,
			Protocol:  "BGP",
			LocalPref: p.LocalPref,
			MED:       p.Metric,
			Origin:    p.Origin,
		}
		if len(p.Nexthops) > 0 {
			rp.NextHop = p.Nexthops[0].IP
			rp.Interface = p.Nexthops[0].InterfaceName
		}
		if p.ASPath != nil && p.ASPath.String != "" && p.ASPath.String != "Local" {
			rp.ASPath = buildASHops(strings.Fields(p.ASPath.String))
		}
		if p.Community != nil {
			for _, c := range p.Community.List {
				rp.Communities = append(rp.Communities, decorateCommunity(c, customComm))
			}
		}
		if p.LargeCommunity != nil && p.LargeCommunity.String != "" {
			for _, c := range strings.Fields(p.LargeCommunity.String) {
				rp.Communities = append(rp.Communities, decorateCommunity(c, customComm))
			}
		}
		res.Paths = append(res.Paths, rp)
	}
	return res, nil
}

func (d *frrDriver) Summary(ctx context.Context) (SummaryResult, error) {
	if d.binPath == "" {
		return SummaryResult{}, fmt.Errorf("vtysh binary not found")
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	out, err := exec.CommandContext(cctx, d.binPath, "-c", "show bgp summary json").Output()
	if err != nil {
		return SummaryResult{}, fmt.Errorf("vtysh summary failed: %w", err)
	}

	var doc map[string]struct {
		AS    int `json:"as"`
		Peers map[string]struct {
			RemoteAs   int    `json:"remoteAs"`
			State      string `json:"state"`
			PeerUptime string `json:"peerUptime"`
			PfxRcd     int    `json:"pfxRcd"`
			Desc       string `json:"desc"`
		} `json:"peers"`
	}
	res := SummaryResult{
		Driver:    "frr",
		RouterASN: normalizeRouterASN(d.cfg.RouterASN),
		RawOutput: string(out),
	}
	if json.Unmarshal(out, &doc) == nil {
		for afi, section := range doc {
			if res.RouterASN == "" && section.AS > 0 {
				res.RouterASN = "AS" + strconv.Itoa(section.AS)
			}
			for addr, peer := range section.Peers {
				asn := "AS" + strconv.Itoa(peer.RemoteAs)
				name := peer.Desc
				if name == "" {
					name = afi + ":" + addr
				}
				res.Peers = append(res.Peers, PeerSession{
					Name:         name,
					Neighbor:     addr,
					ASN:          asn,
					Org:          enrich.LookupASNName(asn),
					State:        peer.State,
					Since:        peer.PeerUptime,
					PrefixesRcvd: peer.PfxRcd,
				})
			}
		}
	}
	return res, nil
}
