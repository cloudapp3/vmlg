package enrich

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"
)

type HopMeta struct {
	PTR     string `json:"ptr,omitempty"`
	ASN     string `json:"asn,omitempty"`
	Org     string `json:"org,omitempty"`
	LineTag string `json:"line_tag,omitempty"`
	Tier    string `json:"tier,omitempty"` // "premium" | "standard" | "ixp" | "global"
	IXP     string `json:"ixp,omitempty"`
}

type RouteVerdict struct {
	Code   string   `json:"code"`
	Label  string   `json:"label"`
	Tier   string   `json:"tier"` // "premium" | "standard" | "global" | "direct"
	Badges []string `json:"badges,omitempty"`
}

type staticRule struct {
	prefix  netip.Prefix
	asn     string
	org     string
	lineTag string
	tier    string
	ixp     string
}

type asnMeta struct {
	org     string
	lineTag string
	tier    string
}

var staticPrefixes = buildStaticPrefixes()

var knownASNs = map[string]asnMeta{
	"AS4809":  {org: "China Telecom CN2", lineTag: "CN2", tier: "premium"},
	"AS4134":  {org: "China Telecom Backbone", lineTag: "CT 163", tier: "standard"},
	"AS9929":  {org: "China Unicom CUII", lineTag: "CU 9929", tier: "premium"},
	"AS10099": {org: "China Unicom Global", lineTag: "CU 10099", tier: "premium"},
	"AS4837":  {org: "China Unicom 169", lineTag: "CU 4837", tier: "standard"},
	"AS58807": {org: "China Mobile CMIN2", lineTag: "CMIN2", tier: "premium"},
	"AS58453": {org: "China Mobile International", lineTag: "CMI", tier: "standard"},
	"AS9808":  {org: "China Mobile Backbone", lineTag: "CM 9808", tier: "standard"},
	"AS4538":  {org: "CERNET", lineTag: "CERNET", tier: "standard"},
	"AS7497":  {org: "CSTNET", lineTag: "CSTNET", tier: "standard"},
	"AS2497":  {org: "IIJ Japan", lineTag: "IIJ", tier: "global"},
	"AS17676": {org: "SoftBank Corp.", lineTag: "SoftBank", tier: "global"},
	"AS2914":  {org: "NTT Global IP Network", lineTag: "NTT", tier: "global"},
	"AS2516":  {org: "KDDI Corporation", lineTag: "KDDI", tier: "global"},
	"AS3491":  {org: "PCCW Global", lineTag: "PCCW", tier: "global"},
	"AS1299":  {org: "Arelion (Telia)", lineTag: "Arelion", tier: "global"},
	"AS174":   {org: "Cogent Communications", lineTag: "Cogent", tier: "global"},
	"AS3356":  {org: "Lumen (Level3)", lineTag: "Lumen", tier: "global"},
	"AS6939":  {org: "Hurricane Electric", lineTag: "HE", tier: "global"},
	"AS7473":  {org: "Singtel", lineTag: "Singtel", tier: "global"},
	"AS4637":  {org: "Telstra Global", lineTag: "Telstra", tier: "global"},
	"AS3462":  {org: "Chunghwa Telecom HiNet", lineTag: "HiNet", tier: "global"},
	"AS4766":  {org: "Korea Telecom", lineTag: "KT", tier: "global"},
	"AS13335": {org: "Cloudflare, Inc.", lineTag: "Cloudflare", tier: "global"},
	"AS15169": {org: "Google LLC", lineTag: "Google", tier: "global"},
	"AS16509": {org: "Amazon.com, Inc.", lineTag: "AWS", tier: "global"},
	"AS8075":  {org: "Microsoft Corporation", lineTag: "Azure", tier: "global"},
}

type cacheEntry struct {
	meta      HopMeta
	expiresAt time.Time
}

type Enricher struct {
	enableDNSASN bool
	enablePTR    bool
	resolver     *net.Resolver

	mu    sync.RWMutex
	cache map[netip.Addr]cacheEntry
}

func New(enableDNSASN, enablePTR bool) *Enricher {
	return &Enricher{
		enableDNSASN: enableDNSASN,
		enablePTR:    enablePTR,
		resolver:     net.DefaultResolver,
		cache:        make(map[netip.Addr]cacheEntry),
	}
}

// LookupASNName returns the known carrier/organization name for an ASN (e.g. "AS4809").
func LookupASNName(asn string) string {
	norm := strings.ToUpper(strings.TrimSpace(asn))
	if !strings.HasPrefix(norm, "AS") {
		norm = "AS" + norm
	}
	if m, ok := knownASNs[norm]; ok {
		return m.org
	}
	return ""
}

func (e *Enricher) Lookup(ctx context.Context, addr netip.Addr) HopMeta {
	ip := addr.Unmap()
	if !ip.IsValid() {
		return HopMeta{}
	}

	now := time.Now()
	e.mu.RLock()
	if entry, ok := e.cache[ip]; ok && now.Before(entry.expiresAt) {
		e.mu.RUnlock()
		return entry.meta
	}
	e.mu.RUnlock()

	meta := lookupStatic(ip)

	if !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
		var wg sync.WaitGroup
		var ptr string
		var dnsASN, dnsOrg string

		if e.enablePTR {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cctx, cancel := context.WithTimeout(ctx, 600*time.Millisecond)
				defer cancel()
				names, err := e.resolver.LookupAddr(cctx, ip.String())
				if err == nil && len(names) > 0 {
					ptr = strings.TrimSuffix(names[0], ".")
				}
			}()
		}

		if e.enableDNSASN && meta.ASN == "" && meta.IXP == "" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cctx, cancel := context.WithTimeout(ctx, 700*time.Millisecond)
				defer cancel()
				dnsASN, dnsOrg = e.lookupCymruASN(cctx, ip)
			}()
		}

		wg.Wait()

		if ptr != "" {
			meta.PTR = ptr
			enrichFromPTR(&meta, ptr)
		}
		if meta.ASN == "" && dnsASN != "" {
			meta.ASN = dnsASN
		}
		if meta.Org == "" && dnsOrg != "" {
			meta.Org = dnsOrg
		}
		if meta.ASN != "" {
			if known, ok := knownASNs[meta.ASN]; ok {
				if meta.Org == "" {
					meta.Org = known.org
				}
				if meta.LineTag == "" {
					meta.LineTag = known.lineTag
				}
				if meta.Tier == "" {
					meta.Tier = known.tier
				}
			}
		}
	}

	e.mu.Lock()
	if len(e.cache) >= 4096 {
		for k, v := range e.cache {
			if now.After(v.expiresAt) {
				delete(e.cache, k)
			}
		}
		if len(e.cache) >= 4096 {
			clear(e.cache)
		}
	}
	e.cache[ip] = cacheEntry{
		meta:      meta,
		expiresAt: now.Add(30 * time.Minute),
	}
	e.mu.Unlock()

	return meta
}

func lookupStatic(ip netip.Addr) HopMeta {
	for _, r := range staticPrefixes {
		if r.prefix.Contains(ip) {
			return HopMeta{
				ASN:     r.asn,
				Org:     r.org,
				LineTag: r.lineTag,
				Tier:    r.tier,
				IXP:     r.ixp,
			}
		}
	}
	return HopMeta{}
}

func enrichFromPTR(m *HopMeta, ptr string) {
	lower := strings.ToLower(ptr)
	switch {
	case strings.Contains(lower, "chinatelecom.com") || strings.Contains(lower, "chinanet"):
		if m.LineTag == "" {
			m.LineTag = "CT 163"
			m.Tier = "standard"
		}
	case strings.Contains(lower, "cndata") || strings.Contains(lower, "cn2"):
		m.LineTag = "CN2"
		m.Tier = "premium"
	case strings.Contains(lower, "cuii") || strings.Contains(lower, "unicom-cug"):
		m.LineTag = "CU 9929"
		m.Tier = "premium"
	case strings.Contains(lower, "china169"):
		if m.LineTag == "" {
			m.LineTag = "CU 4837"
			m.Tier = "standard"
		}
	case strings.Contains(lower, "cmi.chinamobile"):
		if m.LineTag == "" {
			m.LineTag = "CMI"
			m.Tier = "standard"
		}
	case strings.Contains(lower, "iij.net"):
		if m.LineTag == "" {
			m.LineTag = "IIJ"
			m.Tier = "global"
		}
	case strings.Contains(lower, "bbtec.net") || strings.Contains(lower, "softbank"):
		if m.LineTag == "" {
			m.LineTag = "SoftBank"
			m.Tier = "global"
		}
	case strings.Contains(lower, "ntt.net"):
		if m.LineTag == "" {
			m.LineTag = "NTT"
			m.Tier = "global"
		}
	case strings.Contains(lower, "hkix.net"):
		m.IXP = "HKIX"
		m.LineTag = "HKIX"
		m.Tier = "ixp"
	case strings.Contains(lower, "de-cix.net"):
		m.IXP = "DE-CIX"
		m.LineTag = "DE-CIX"
		m.Tier = "ixp"
	}
}

func (e *Enricher) lookupCymruASN(ctx context.Context, ip netip.Addr) (string, string) {
	var queryName string
	if ip.Is4() {
		b := ip.As4()
		queryName = fmt.Sprintf("%d.%d.%d.%d.origin.asn.cymru.com", b[3], b[2], b[1], b[0])
	} else {
		b := ip.As16()
		var nibbles [64]byte
		idx := 0
		const hexChars = "0123456789abcdef"
		for i := 15; i >= 0; i-- {
			nibbles[idx] = hexChars[b[i]&0x0f]
			nibbles[idx+1] = '.'
			nibbles[idx+2] = hexChars[b[i]>>4]
			nibbles[idx+3] = '.'
			idx += 4
		}
		queryName = string(nibbles[:]) + "origin6.asn.cymru.com"
	}

	txts, err := e.resolver.LookupTXT(ctx, queryName)
	if err != nil || len(txts) == 0 {
		return "", ""
	}
	// Example TXT: "13335 | 1.1.1.0/24 | AU | apnic | 2011-08-11"
	parts := strings.Split(txts[0], "|")
	if len(parts) == 0 {
		return "", ""
	}
	asnFields := strings.Fields(parts[0])
	if len(asnFields) == 0 {
		return "", ""
	}
	asnNum := asnFields[0]
	asn := "AS" + asnNum

	if known, ok := knownASNs[asn]; ok {
		return asn, known.org
	}

	orgTxts, err := e.resolver.LookupTXT(ctx, fmt.Sprintf("AS%s.asn.cymru.com", asnNum))
	if err == nil && len(orgTxts) > 0 {
		// Example TXT: "13335 | US | arin | 2010-07-14 | CLOUDFLARENET, US"
		orgParts := strings.Split(orgTxts[0], "|")
		if len(orgParts) >= 5 {
			return asn, strings.TrimSpace(orgParts[4])
		}
	}
	return asn, ""
}

// SummarizeRoute inspects enriched hops of a trace and determines the primary
// carrier route classification (e.g. CN2 GIA, CN2 GT, CU 9929, CMIN2, CT 163, etc.).
func SummarizeRoute(hops []HopMeta) RouteVerdict {
	var (
		has5943  bool
		has163   bool
		has9929  bool
		has10099 bool
		has4837  bool
		hasCMIN2 bool
		hasCMI   bool
		has9808  bool
	)
	seenBadges := make([]string, 0, 4)
	seenMap := make(map[string]bool)
	addBadge := func(tag string) {
		if tag == "" || seenMap[tag] {
			return
		}
		seenMap[tag] = true
		seenBadges = append(seenBadges, tag)
	}

	for _, h := range hops {
		switch h.LineTag {
		case "CN2 GIA", "CN2":
			has5943 = true
		case "CT 163":
			has163 = true
		case "CU 9929":
			has9929 = true
		case "CU 10099":
			has10099 = true
		case "CU 4837":
			has4837 = true
		case "CMIN2":
			hasCMIN2 = true
		case "CMI":
			hasCMI = true
		case "CM 9808":
			has9808 = true
		}
		if h.IXP != "" {
			addBadge(h.IXP)
		} else if h.LineTag != "" {
			addBadge(h.LineTag)
		}
	}

	switch {
	case has5943 && !has163:
		return RouteVerdict{Code: "CN2_GIA", Label: "CN2 GIA", Tier: "premium", Badges: seenBadges}
	case has5943 && has163:
		return RouteVerdict{Code: "CN2_GT", Label: "CN2 GT", Tier: "premium", Badges: seenBadges}
	case has9929:
		label := "CU 9929"
		if has10099 {
			label = "CU 10099+9929"
		}
		return RouteVerdict{Code: "CU_9929", Label: label, Tier: "premium", Badges: seenBadges}
	case hasCMIN2:
		return RouteVerdict{Code: "CMIN2", Label: "CMIN2", Tier: "premium", Badges: seenBadges}
	case has10099 && has4837:
		return RouteVerdict{Code: "CU_10099_4837", Label: "CU 10099+4837", Tier: "standard", Badges: seenBadges}
	case has4837:
		return RouteVerdict{Code: "CU_4837", Label: "CU 4837", Tier: "standard", Badges: seenBadges}
	case has163:
		return RouteVerdict{Code: "CT_163", Label: "CT 163", Tier: "standard", Badges: seenBadges}
	case hasCMI || has9808:
		label := "CMI"
		if !hasCMI {
			label = "CM 9808"
		}
		return RouteVerdict{Code: "CMI", Label: label, Tier: "standard", Badges: seenBadges}
	}

	if len(seenBadges) > 0 {
		return RouteVerdict{
			Code:   strings.ToUpper(strings.ReplaceAll(seenBadges[0], " ", "_")),
			Label:  seenBadges[0],
			Tier:   "global",
			Badges: seenBadges,
		}
	}
	return RouteVerdict{Code: "BGP_DIRECT", Label: "BGP / Direct", Tier: "direct"}
}

func buildStaticPrefixes() []staticRule {
	raw := []struct {
		cidr, asn, org, tag, tier, ixp string
	}{
		// Premium Chinese Carrier Backbones
		{"59.43.0.0/16", "AS4809", "China Telecom CN2", "CN2 GIA", "premium", ""},
		{"202.97.0.0/16", "AS4134", "China Telecom 163 Backbone", "CT 163", "standard", ""},
		{"218.105.0.0/16", "AS9929", "China Unicom CUII (9929)", "CU 9929", "premium", ""},
		{"210.51.0.0/16", "AS9929", "China Unicom CUII (9929)", "CU 9929", "premium", ""},
		{"210.78.0.0/16", "AS10099", "China Unicom Global (CUG)", "CU 10099", "premium", ""},
		{"219.158.0.0/16", "AS4837", "China Unicom 169 Backbone", "CU 4837", "standard", ""},
		{"223.118.0.0/15", "AS58807", "China Mobile CMIN2", "CMIN2", "premium", ""},
		{"223.120.0.0/15", "AS58453", "China Mobile International", "CMI", "standard", ""},
		{"221.176.0.0/13", "AS9808", "China Mobile Backbone", "CM 9808", "standard", ""},
		{"221.183.0.0/16", "AS9808", "China Mobile Backbone", "CM 9808", "standard", ""},

		// Major Internet Exchange Peering LANs (IXPs)
		{"123.255.88.0/21", "", "Hong Kong Internet Exchange", "HKIX", "ixp", "HKIX"},
		{"80.81.192.0/21", "", "DE-CIX Frankfurt", "DE-CIX", "ixp", "DE-CIX"},
		{"80.249.208.0/21", "", "Amsterdam Internet Exchange", "AMS-IX", "ixp", "AMS-IX"},
		{"195.66.224.0/22", "", "London Internet Exchange", "LINX", "ixp", "LINX"},
		{"103.16.102.0/23", "", "Equinix Singapore IX", "Equinix-SG", "ixp", "Equinix-SG"},
		{"203.190.230.0/23", "", "Equinix Tokyo IX", "Equinix-TY", "ixp", "Equinix-TY"},
		{"210.171.224.0/23", "", "JPIX Tokyo", "JPIX", "ixp", "JPIX"},
		{"218.100.6.0/23", "", "BBIX Tokyo", "BBIX", "ixp", "BBIX"},
		{"206.72.210.0/23", "", "CoreSite Any2IX Los Angeles", "Any2IX", "ixp", "Any2IX"},

		// Well-known Anycast / Public DNS for instant zero-latency tags
		{"1.1.1.0/24", "AS13335", "Cloudflare, Inc.", "Cloudflare", "global", ""},
		{"1.0.0.0/24", "AS13335", "Cloudflare, Inc.", "Cloudflare", "global", ""},
		{"8.8.8.0/24", "AS15169", "Google LLC", "Google", "global", ""},
		{"8.8.4.0/24", "AS15169", "Google LLC", "Google", "global", ""},
	}
	out := make([]staticRule, 0, len(raw))
	for _, item := range raw {
		out = append(out, staticRule{
			prefix:  netip.MustParsePrefix(item.cidr),
			asn:     item.asn,
			org:     item.org,
			lineTag: item.tag,
			tier:    item.tier,
			ixp:     item.ixp,
		})
	}
	return out
}
