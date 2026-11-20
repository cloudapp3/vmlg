package bgp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RPKI states surfaced to API/UI. "unavailable" is transport-level degradation,
// never a routing verdict.
const (
	RPKIStateValid       = "valid"
	RPKIStateInvalid     = "invalid"
	RPKIStateNotFound    = "notfound"
	RPKIStateUnavailable = "unavailable"
)

type RPKIROA struct {
	ASN       string `json:"asn"`
	Prefix    string `json:"prefix"`
	MaxLength int    `json:"max_length"`
}

type RPKIValidity struct {
	State       string    `json:"state"`
	Origin      string    `json:"origin,omitempty"`
	Prefix      string    `json:"prefix,omitempty"`
	Reason      string    `json:"reason,omitempty"`
	Description string    `json:"description,omitempty"`
	VRPs        []RPKIROA `json:"vrps,omitempty"` // covering ROAs (or conflicting ones when Invalid)
}

// RPKIValidator queries a Routinator-compatible HTTP API
// (GET {base}/api/v1/validity/{asn}/{prefix}). Nil when unconfigured —
// callers must treat nil as "feature absent", per ADR-VMLG-003.
type RPKIValidator struct {
	baseURL string
	client  *http.Client
}

func NewRPKIValidator(rawURL string, timeout time.Duration) *RPKIValidator {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil
	}
	if timeout <= 0 || timeout > 3*time.Second {
		timeout = 3 * time.Second
	}
	return &RPKIValidator{
		baseURL: strings.TrimRight(rawURL, "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

// Validate never returns an error: transport failures degrade to
// State=unavailable so the BGP main result stays unaffected.
func (v *RPKIValidator) Validate(ctx context.Context, asn, prefix string) RPKIValidity {
	out := RPKIValidity{Origin: asn, Prefix: prefix}
	if v == nil {
		out.State = RPKIStateUnavailable
		return out
	}

	url := fmt.Sprintf("%s/api/v1/validity/%s/%s", v.baseURL, normalizeASNumber(asn), prefix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		out.State, out.Reason = RPKIStateUnavailable, err.Error()
		return out
	}
	resp, err := v.client.Do(req)
	if err != nil {
		out.State, out.Reason = RPKIStateUnavailable, err.Error()
		return out
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		out.State = RPKIStateUnavailable
		out.Reason = fmt.Sprintf("routinator returned %s", resp.Status)
		return out
	}

	// Routinator shape: {"validated_route":{"validity":{"state":"valid",
	//   "reason":null,"description":"...","VRPs":[{"asn":"AS13335",
	//   "prefix":"104.16.0.0/12","max_length":12}]}}, ...}
	var doc struct {
		ValidatedRoute struct {
			Validity struct {
				State       string `json:"state"`
				Reason      any    `json:"reason"`
				Description string `json:"description"`
				VRPs        []struct {
					ASN       string `json:"asn"`
					Prefix    string `json:"prefix"`
					MaxLength int    `json:"max_length"`
				} `json:"VRPs"`
			} `json:"validity"`
		} `json:"validated_route"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		out.State, out.Reason = RPKIStateUnavailable, err.Error()
		return out
	}

	switch strings.ToLower(doc.ValidatedRoute.Validity.State) {
	case "valid":
		out.State = RPKIStateValid
	case "invalid":
		out.State = RPKIStateInvalid
	default:
		out.State = RPKIStateNotFound
	}
	out.Description = doc.ValidatedRoute.Validity.Description
	if r, ok := doc.ValidatedRoute.Validity.Reason.(string); ok {
		out.Reason = r
	}
	for _, vrp := range doc.ValidatedRoute.Validity.VRPs {
		out.VRPs = append(out.VRPs, RPKIROA{ASN: vrp.ASN, Prefix: vrp.Prefix, MaxLength: vrp.MaxLength})
	}
	return out
}

// BestPathOriginASN extracts the origin ASN (rightmost AS_PATH entry of the
// best path) and the matched prefix for RPKI validation.
func BestPathOriginASN(res RouteLookupResult) (asn string, prefix string) {
	prefix = res.Prefix
	for _, p := range res.Paths {
		if !p.Best || len(p.ASPath) == 0 {
			continue
		}
		return p.ASPath[len(p.ASPath)-1].ASN, prefix
	}
	// No explicit best flag: fall back to the first path with an AS_PATH.
	for _, p := range res.Paths {
		if len(p.ASPath) > 0 {
			return p.ASPath[len(p.ASPath)-1].ASN, prefix
		}
	}
	return "", prefix
}

func normalizeASNumber(asn string) string {
	asn = strings.TrimSpace(asn)
	if asn == "" {
		return asn
	}
	if !strings.HasPrefix(strings.ToUpper(asn), "AS") {
		return "AS" + asn
	}
	return asn
}
