package cluster

import (
	"context"
	"net/netip"
	"time"

	"github.com/cloudapp3/vmlg/internal/bgp"
	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
	"github.com/cloudapp3/vmlg/internal/guard"
	"github.com/cloudapp3/vmlg/internal/probe"
	"github.com/cloudapp3/vmlg/internal/speedtest"
	"github.com/cloudapp3/vmlg/internal/version"
)

type LocalNode struct {
	cfg          config.Config
	guard        *guard.Guard
	enforceGuard bool
	enricher     *enrich.Enricher
	iperf        *speedtest.IPerfManager
	bgpDrv       bgp.Driver
	startedAt    time.Time
}

func NewLocalNode(cfg config.Config, g *guard.Guard, enforceGuard bool, e *enrich.Enricher, ipm *speedtest.IPerfManager, bgpDrv bgp.Driver) *LocalNode {
	return &LocalNode{
		cfg:          cfg,
		guard:        g,
		enforceGuard: enforceGuard,
		enricher:     e,
		iperf:        ipm,
		bgpDrv:       bgpDrv,
		startedAt:    time.Now().UTC(),
	}
}

func (n *LocalNode) Descriptor() NodeDescriptor {
	iperfOK, iperfReason := false, ""
	if n.iperf != nil {
		iperfOK, iperfReason = n.iperf.Available()
	}
	bgpOK, bgpDriver, bgpNote := false, "", ""
	if n.bgpDrv != nil {
		bgpDriver = n.bgpDrv.Name()
		bgpOK, bgpNote = n.bgpDrv.Available()
	}
	routerASN := n.cfg.BGP.RouterASN
	if routerASN == "" {
		routerASN = n.cfg.Node.ASN
	}
	var testFiles []string
	speedEnabled := n.cfg.Speedtest.Enabled
	if n.cfg.Mode == config.ModeAgent && n.cfg.Node.PublicURL == "" {
		speedEnabled = false
	}
	if speedEnabled {
		testFiles = n.cfg.Speedtest.TestFiles
	}

	activeProbes, activeSpeed := 0, 0
	if n.guard != nil {
		activeProbes = n.guard.ActiveProbes()
		activeSpeed = n.guard.ActiveSpeedtests()
	}

	pubHost := n.cfg.Node.PublicHost
	if pubHost == "" {
		if n.cfg.Node.IPv4 != "" {
			pubHost = n.cfg.Node.IPv4
		} else if n.cfg.Node.IPv6 != "" {
			pubHost = n.cfg.Node.IPv6
		}
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	return NodeDescriptor{
		ID:               n.cfg.Node.ID,
		Name:             n.cfg.Node.Name,
		Location:         n.cfg.Node.Location,
		Region:           n.cfg.Node.Region,
		Provider:         n.cfg.Node.Provider,
		ASN:              n.cfg.Node.ASN,
		IPv4:             n.cfg.Node.IPv4,
		IPv6:             n.cfg.Node.IPv6,
		PublicURL:        n.cfg.Node.PublicURL,
		PublicHost:       pubHost,
		Tags:             n.cfg.Node.Tags,
		Version:          version.Version,
		IsLocal:          n.cfg.Mode != config.ModeAgent,
		Online:           true,
		ConnectedAt:      n.startedAt.Format(time.RFC3339),
		LastSeen:         nowStr,
		ActiveProbes:     activeProbes,
		ActiveSpeedtests: activeSpeed,
		Capabilities: NodeCapabilities{
			Ping:      true,
			MTR:       true,
			Speedtest: speedEnabled,
			TestFiles: testFiles,
			IPerf3:    iperfOK,
			IPerfNote: iperfReason,
			BGP:       bgpOK,
			BGPDriver: bgpDriver,
			RouterASN: routerASN,
			BGPNote:   bgpNote,
		},
	}
}

func (n *LocalNode) Ping(ctx context.Context, clientIP string, req probe.PingRequest, onSample func(probe.PingSample)) (probe.PingSummary, error) {
	if n.guard != nil && n.enforceGuard {
		cip, _ := netip.ParseAddr(clientIP)
		release, err := n.guard.AcquireProbe(cip, req.TargetAddr)
		if err != nil {
			return probe.PingSummary{}, err
		}
		defer release()
	}
	return probe.RunPing(ctx, req, onSample)
}

func (n *LocalNode) MTR(ctx context.Context, clientIP string, req probe.MTRRequest, onUpdate func(probe.MTRSnapshot)) (probe.MTRSnapshot, error) {
	if n.guard != nil && n.enforceGuard {
		cip, _ := netip.ParseAddr(clientIP)
		release, err := n.guard.AcquireProbe(cip, req.TargetAddr)
		if err != nil {
			return probe.MTRSnapshot{}, err
		}
		defer release()
	}
	return probe.RunMTR(ctx, req, n.enricher, onUpdate)
}

func (n *LocalNode) AcquireIPerf(_ context.Context, clientKey string) (speedtest.IPerfLease, error) {
	if n.iperf == nil {
		return speedtest.IPerfLease{}, speedtest.ErrIPerfDisabled
	}
	desc := n.Descriptor()
	return n.iperf.Acquire(clientKey, desc.PublicHost, desc.ID)
}

func (n *LocalNode) BGPRoute(ctx context.Context, target string) (bgp.RouteLookupResult, error) {
	if n.bgpDrv == nil {
		return bgp.RouteLookupResult{}, bgp.ErrBGPDisabled
	}
	return n.bgpDrv.LookupRoute(ctx, target)
}

func (n *LocalNode) BGPSummary(ctx context.Context) (bgp.SummaryResult, error) {
	if n.bgpDrv == nil {
		return bgp.SummaryResult{}, bgp.ErrBGPDisabled
	}
	return n.bgpDrv.Summary(ctx)
}
