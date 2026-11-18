package bgp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
)

type gobgpDriver struct {
	cfg     config.BGPConfig
	binPath string
}

func newGoBGPDriver(cfg config.BGPConfig) *gobgpDriver {
	bin := cfg.GoBGPBinary
	if bin == "" {
		bin = "gobgp"
	}
	path, _ := exec.LookPath(bin)
	return &gobgpDriver{cfg: cfg, binPath: path}
}

func (d *gobgpDriver) Name() string { return "gobgp" }

func (d *gobgpDriver) Available() (bool, string) {
	if d.binPath == "" {
		return false, "gobgp binary not found"
	}
	return true, ""
}

func (d *gobgpDriver) LookupRoute(ctx context.Context, target string) (RouteLookupResult, error) {
	norm, isV6, err := ParseQueryTarget(target)
	if err != nil {
		return RouteLookupResult{}, err
	}
	if d.binPath == "" {
		return RouteLookupResult{}, fmt.Errorf("gobgp binary not found")
	}

	afi := "ipv4"
	if isV6 {
		afi = "ipv6"
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	out, err := exec.CommandContext(cctx, d.binPath, "-j", "global", "rib", "-a", afi, norm).Output()
	if err != nil {
		return RouteLookupResult{}, fmt.Errorf("gobgp query failed: %w", err)
	}

	res := RouteLookupResult{
		Target:    norm,
		Driver:    "gobgp",
		RouterASN: normalizeRouterASN(d.cfg.RouterASN),
		RawOutput: string(out),
	}

	var table map[string][]struct {
		Nlri struct {
			Prefix string `json:"prefix"`
		} `json:"nlri"`
		Best  bool   `json:"best"`
		Nh    string `json:"nexthop"`
		Attrs []struct {
			Type    int    `json:"type"`
			Nexthop string `json:"nexthop"`
			Metric  int    `json:"metric"`
			Value   any    `json:"value"`
			AsPaths []struct {
				Asns []uint32 `json:"asns"`
			} `json:"as_paths"`
		} `json:"attrs"`
	}
	if json.Unmarshal(out, &table) == nil {
		for pfx, paths := range table {
			if res.Prefix == "" {
				res.Prefix = pfx
			}
			for _, p := range paths {
				rp := RoutePath{
					Best:     p.Best,
					Protocol: "GoBGP",
					NextHop:  p.Nh,
				}
				for _, attr := range p.Attrs {
					if attr.Nexthop != "" && rp.NextHop == "" {
						rp.NextHop = attr.Nexthop
					}
					if attr.Metric > 0 {
						rp.MED = attr.Metric
					}
					for _, seg := range attr.AsPaths {
						toks := make([]string, 0, len(seg.Asns))
						for _, asn := range seg.Asns {
							toks = append(toks, strconv.FormatUint(uint64(asn), 10))
						}
						rp.ASPath = append(rp.ASPath, buildASHops(toks)...)
					}
				}
				res.Paths = append(res.Paths, rp)
			}
		}
	}
	return res, nil
}

func (d *gobgpDriver) Summary(ctx context.Context) (SummaryResult, error) {
	if d.binPath == "" {
		return SummaryResult{}, fmt.Errorf("gobgp binary not found")
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	out, err := exec.CommandContext(cctx, d.binPath, "-j", "neighbor").Output()
	if err != nil {
		return SummaryResult{}, fmt.Errorf("gobgp neighbor query failed: %w", err)
	}

	res := SummaryResult{
		Driver:    "gobgp",
		RouterASN: normalizeRouterASN(d.cfg.RouterASN),
		RawOutput: string(out),
	}

	var peers []struct {
		Conf struct {
			NeighborAddress string `json:"neighbor_address"`
			PeerAs          uint32 `json:"peer_as"`
			Description     string `json:"description"`
		} `json:"conf"`
		State struct {
			SessionState string `json:"session_state"`
		} `json:"state"`
	}
	if json.Unmarshal(out, &peers) == nil {
		for _, p := range peers {
			asn := fmt.Sprintf("AS%d", p.Conf.PeerAs)
			name := p.Conf.Description
			if name == "" {
				name = p.Conf.NeighborAddress
			}
			res.Peers = append(res.Peers, PeerSession{
				Name:     name,
				Neighbor: p.Conf.NeighborAddress,
				ASN:      asn,
				Org:      enrich.LookupASNName(asn),
				State:    p.State.SessionState,
			})
		}
	}
	return res, nil
}
