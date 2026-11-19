package probe

import (
	"net/netip"

	"github.com/cloudapp3/vmlg/internal/enrich"
)

type PingRequest struct {
	Target     string     `json:"target"`
	TargetAddr netip.Addr `json:"target_addr"`
	Protocol   string     `json:"protocol"` // "icmp" | "tcp"
	Port       int        `json:"port,omitempty"`
	Count      int        `json:"count"`
	TimeoutMs  int        `json:"timeout_ms"`
	IntervalMs int        `json:"interval_ms"`
}

type PingSample struct {
	Seq     int     `json:"seq"`
	Target  string  `json:"target"`
	Addr    string  `json:"addr"`
	Proto   string  `json:"proto"`
	TTL     int     `json:"ttl,omitempty"`
	RTTMs   float64 `json:"rtt_ms,omitempty"`
	Timeout bool    `json:"timeout,omitempty"`
	Error   string  `json:"error,omitempty"`
}

type PingSummary struct {
	Target   string       `json:"target"`
	Addr     string       `json:"addr"`
	Protocol string       `json:"protocol"`
	Engine   string       `json:"engine"`
	Sent     int          `json:"sent"`
	Recv     int          `json:"recv"`
	LossPct  float64      `json:"loss_pct"`
	MinMs    float64      `json:"min_ms"`
	AvgMs    float64      `json:"avg_ms"`
	MaxMs    float64      `json:"max_ms"`
	JitterMs float64      `json:"jitter_ms"`
	StDevMs  float64      `json:"stdev_ms"`
	Samples  []PingSample `json:"samples,omitempty"`
}

type DigRequest struct {
	Name      string `json:"name"`
	RRType    string `json:"rrtype"`             // A | AAAA | CNAME | MX | NS | TXT | PTR
	Resolver  string `json:"resolver,omitempty"` // IP[:port]; empty = node system resolver
	TimeoutMs int    `json:"timeout_ms"`
}

type DigRecord struct {
	Name  string `json:"name"`
	TTL   uint32 `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type DigResult struct {
	Name      string      `json:"name"`
	RRType    string      `json:"rrtype"`
	Resolver  string      `json:"resolver"`            // effective resolver: "system" or IP[:port]
	Transport string      `json:"transport"`           // udp | tcp
	RCode     string      `json:"rcode"`               // NOERROR, NXDOMAIN, SERVFAIL, ...
	AD        bool        `json:"ad,omitempty"`        // DNSSEC authenticated data flag passthrough
	Truncated bool        `json:"truncated,omitempty"` // TC bit set on UDP response
	Answers   []DigRecord `json:"answers"`             // full answer section incl. CNAME chain
	ElapsedMs float64     `json:"elapsed_ms"`
}

type MTRRequest struct {
	Target     string     `json:"target"`
	TargetAddr netip.Addr `json:"target_addr"`
	Cycles     int        `json:"cycles"`
	MaxHops    int        `json:"max_hops"`
	TimeoutMs  int        `json:"timeout_ms"`
}

type MTRHop struct {
	Hop     int     `json:"hop"`
	Addr    string  `json:"addr,omitempty"`
	PTR     string  `json:"ptr,omitempty"`
	ASN     string  `json:"asn,omitempty"`
	Org     string  `json:"org,omitempty"`
	LineTag string  `json:"line_tag,omitempty"`
	Tier    string  `json:"tier,omitempty"`
	IXP     string  `json:"ixp,omitempty"`
	Sent    int     `json:"sent"`
	Recv    int     `json:"recv"`
	LossPct float64 `json:"loss_pct"`
	LastMs  float64 `json:"last_ms"`
	AvgMs   float64 `json:"avg_ms"`
	BestMs  float64 `json:"best_ms"`
	WrstMs  float64 `json:"wrst_ms"`
	StDevMs float64 `json:"stdev_ms"`
}

type MTRSnapshot struct {
	Target    string              `json:"target"`
	Addr      string              `json:"addr"`
	Engine    string              `json:"engine"`
	Cycle     int                 `json:"cycle"`
	MaxCycles int                 `json:"max_cycles"`
	Completed bool                `json:"completed"`
	Route     enrich.RouteVerdict `json:"route"`
	Hops      []MTRHop            `json:"hops"`
}
