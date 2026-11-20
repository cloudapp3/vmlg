package bgp

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
)

var (
	ErrBGPDisabled   = errors.New("BGP looking glass is not enabled on this node")
	ErrInvalidPrefix = errors.New("invalid BGP query target: must be an IPv4/IPv6 address or CIDR prefix")
)

type ASHop struct {
	ASN  string `json:"asn"`
	Name string `json:"name,omitempty"`
}

type CommunityTag struct {
	Raw   string `json:"raw"`
	Label string `json:"label,omitempty"`
}

type RoutePath struct {
	Best        bool           `json:"best"`
	Protocol    string         `json:"protocol,omitempty"`
	NextHop     string         `json:"next_hop,omitempty"`
	Interface   string         `json:"interface,omitempty"`
	Since       string         `json:"since,omitempty"`
	LocalPref   int            `json:"local_pref,omitempty"`
	MED         int            `json:"med,omitempty"`
	Origin      string         `json:"origin,omitempty"`
	ASPath      []ASHop        `json:"as_path,omitempty"`
	Communities []CommunityTag `json:"communities,omitempty"`
}

type RouteLookupResult struct {
	Target    string        `json:"target"`
	Prefix    string        `json:"prefix,omitempty"`
	Driver    string        `json:"driver"`
	RouterASN string        `json:"router_asn,omitempty"`
	Paths     []RoutePath   `json:"paths"`
	RawOutput string        `json:"raw_output,omitempty"`
	RPKI      *RPKIValidity `json:"rpki,omitempty"` // only when [rpki] routinator_url is configured
}

type PeerSession struct {
	Name         string `json:"name"`
	Neighbor     string `json:"neighbor,omitempty"`
	ASN          string `json:"asn,omitempty"`
	Org          string `json:"org,omitempty"`
	State        string `json:"state"`
	Since        string `json:"since,omitempty"`
	PrefixesRcvd int    `json:"prefixes_rcvd,omitempty"`
	Info         string `json:"info,omitempty"`
}

type SummaryResult struct {
	Driver    string        `json:"driver"`
	RouterASN string        `json:"router_asn,omitempty"`
	Peers     []PeerSession `json:"peers"`
	RawOutput string        `json:"raw_output,omitempty"`
}

type Driver interface {
	Name() string
	Available() (bool, string)
	LookupRoute(ctx context.Context, target string) (RouteLookupResult, error)
	Summary(ctx context.Context) (SummaryResult, error)
}

var wellKnownCommunities = map[string]string{
	"65535:65281": "NO_EXPORT",
	"65535:65282": "NO_ADVERTISE",
	"65535:65283": "NO_EXPORT_SUBCONFED",
	"65535:65284": "NOPEER",
	"65535:666":   "BLACKHOLE",
	"65535:0":     "GRACEFUL_SHUTDOWN",
}

func NewDriver(cfg config.BGPConfig) Driver {
	if !cfg.Enabled {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "bird":
		return newBirdDriver(cfg)
	case "frr":
		return newFRRDriver(cfg)
	case "gobgp":
		return newGoBGPDriver(cfg)
	case "mock":
		return newMockDriver(cfg)
	default:
		return nil
	}
}

// ParseQueryTarget validates an IP literal or CIDR prefix and returns normalized string + IPv6 flag.
func ParseQueryTarget(raw string) (normalized string, isV6 bool, err error) {
	s := strings.TrimSpace(raw)
	if s == "" || len(s) > 64 {
		return "", false, ErrInvalidPrefix
	}
	if strings.Contains(s, "/") {
		p, perr := netip.ParsePrefix(s)
		if perr != nil {
			return "", false, ErrInvalidPrefix
		}
		p = p.Masked()
		return p.String(), p.Addr().Is6(), nil
	}
	ip, ierr := netip.ParseAddr(s)
	if ierr != nil {
		return "", false, ErrInvalidPrefix
	}
	ip = ip.Unmap()
	return ip.String(), ip.Is6(), nil
}

func buildASHops(tokens []string) []ASHop {
	out := make([]ASHop, 0, len(tokens))
	for _, tok := range tokens {
		clean := strings.Trim(tok, "{}[]()")
		if clean == "" {
			continue
		}
		upper := strings.ToUpper(clean)
		if !strings.HasPrefix(upper, "AS") {
			// Verify numeric ASN
			allDigits := true
			for i := 0; i < len(upper); i++ {
				if upper[i] < '0' || upper[i] > '9' {
					allDigits = false
					break
				}
			}
			if !allDigits {
				continue
			}
			upper = "AS" + upper
		}
		out = append(out, ASHop{
			ASN:  upper,
			Name: enrich.LookupASNName(upper),
		})
	}
	return out
}

func decorateCommunity(raw string, custom map[string]string) CommunityTag {
	norm := strings.TrimSpace(raw)
	norm = strings.Trim(norm, "()")
	norm = strings.ReplaceAll(norm, ",", ":")
	norm = strings.Join(strings.Fields(norm), "")
	// Replace multiple colons if spaces existed around commas
	for strings.Contains(norm, "::") {
		norm = strings.ReplaceAll(norm, "::", ":")
	}

	label := ""
	if custom != nil {
		label = custom[norm]
	}
	if label == "" {
		label = wellKnownCommunities[norm]
	}
	return CommunityTag{
		Raw:   norm,
		Label: label,
	}
}

func normalizeRouterASN(asn string) string {
	asn = strings.ToUpper(strings.TrimSpace(asn))
	if asn == "" {
		return ""
	}
	if !strings.HasPrefix(asn, "AS") {
		return fmt.Sprintf("AS%s", asn)
	}
	return asn
}
