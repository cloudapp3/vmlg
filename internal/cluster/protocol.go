package cluster

import (
	"context"
	"encoding/json"

	"github.com/cloudapp3/vmlg/internal/bgp"
	"github.com/cloudapp3/vmlg/internal/probe"
	"github.com/cloudapp3/vmlg/internal/speedtest"
)

const (
	MsgRegister     = "register"
	MsgRegistered   = "registered"
	MsgHeartbeat    = "heartbeat"
	MsgProbeStart   = "probe_start"
	MsgProbeSample  = "probe_sample"
	MsgProbeMTR     = "probe_mtr"
	MsgProbeDone    = "probe_done"
	MsgProbeError   = "probe_error"
	MsgProbeCancel  = "probe_cancel"
	MsgIPerfAcquire = "iperf_acquire"
	MsgIPerfResult  = "iperf_result"
	MsgBGPRoute     = "bgp_route"
	MsgBGPSummary   = "bgp_summary"
	MsgBGPResult    = "bgp_result"
)

type NodeCapabilities struct {
	Ping      bool     `json:"ping"`
	MTR       bool     `json:"mtr"`
	Speedtest bool     `json:"speedtest"`
	TestFiles []string `json:"test_files,omitempty"`
	IPerf3    bool     `json:"iperf3"`
	IPerfNote string   `json:"iperf_note,omitempty"`
	BGP       bool     `json:"bgp"`
	BGPDriver string   `json:"bgp_driver,omitempty"`
	RouterASN string   `json:"router_asn,omitempty"`
	BGPNote   string   `json:"bgp_note,omitempty"`
}

type NodeDescriptor struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Location         string           `json:"location"`
	Region           string           `json:"region"`
	Provider         string           `json:"provider,omitempty"`
	ASN              string           `json:"asn,omitempty"`
	IPv4             string           `json:"ipv4,omitempty"`
	IPv6             string           `json:"ipv6,omitempty"`
	PublicURL        string           `json:"public_url,omitempty"`
	PublicHost       string           `json:"public_host,omitempty"`
	Tags             []string         `json:"tags,omitempty"`
	Version          string           `json:"version"`
	IsLocal          bool             `json:"is_local,omitempty"`
	Online           bool             `json:"online"`
	ConnectedAt      string           `json:"connected_at,omitempty"`
	LastSeen         string           `json:"last_seen,omitempty"`
	ActiveProbes     int              `json:"active_probes"`
	ActiveSpeedtests int              `json:"active_speedtests"`
	SpeedtestTicket  string           `json:"speedtest_ticket,omitempty"`
	Capabilities     NodeCapabilities `json:"capabilities"`
}

type HeartbeatPayload struct {
	ActiveProbes     int `json:"active_probes"`
	ActiveSpeedtests int `json:"active_speedtests"`
}

type ProbeStartPayload struct {
	Kind     string             `json:"kind"` // "ping" | "mtr"
	ClientIP string             `json:"client_ip,omitempty"`
	Ping     *probe.PingRequest `json:"ping,omitempty"`
	MTR      *probe.MTRRequest  `json:"mtr,omitempty"`
}

type IPerfRequestPayload struct {
	ClientKey string `json:"client_key"`
}

type BGPRoutePayload struct {
	Target string `json:"target"`
}

type Envelope struct {
	Type    string          `json:"type"`
	ReqID   string          `json:"req_id,omitempty"`
	Error   string          `json:"error,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func makeEnvelope(msgType, reqID string, payload any, errStr string) (Envelope, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return Envelope{}, err
		}
		raw = b
	}
	return Envelope{
		Type:    msgType,
		ReqID:   reqID,
		Error:   errStr,
		Payload: raw,
	}, nil
}

type NodeRunner interface {
	Descriptor() NodeDescriptor
	Ping(ctx context.Context, clientIP string, req probe.PingRequest, onSample func(probe.PingSample)) (probe.PingSummary, error)
	MTR(ctx context.Context, clientIP string, req probe.MTRRequest, onUpdate func(probe.MTRSnapshot)) (probe.MTRSnapshot, error)
	AcquireIPerf(ctx context.Context, clientKey string) (speedtest.IPerfLease, error)
	BGPRoute(ctx context.Context, target string) (bgp.RouteLookupResult, error)
	BGPSummary(ctx context.Context) (bgp.SummaryResult, error)
}
