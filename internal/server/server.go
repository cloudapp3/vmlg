package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudapp3/vmlg/internal/bgp"
	"github.com/cloudapp3/vmlg/internal/cluster"
	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
	"github.com/cloudapp3/vmlg/internal/guard"
	"github.com/cloudapp3/vmlg/internal/metrics"
	"github.com/cloudapp3/vmlg/internal/probe"
	"github.com/cloudapp3/vmlg/internal/speedtest"
	"github.com/cloudapp3/vmlg/internal/version"
	"github.com/cloudapp3/vmlg/internal/web"
)

type Server struct {
	cfg       config.Config
	guard     *guard.Guard
	enricher  *enrich.Enricher
	registry  *cluster.Registry
	iperf     *speedtest.IPerfManager
	speedtest *speedtest.Handler
	metrics   *metrics.Collector
	logger    *slog.Logger
	handler   http.Handler
}

func New(cfg config.Config, logger *slog.Logger) (*Server, error) {
	if logger == nil {
		logger = slog.Default()
	}

	g, err := guard.New(cfg.Guard, cfg.Server.TrustedProxies)
	if err != nil {
		return nil, err
	}
	enricher := enrich.New(cfg.Probe.EnableDNSASN, cfg.Probe.EnablePTR)

	var (
		localRunner cluster.NodeRunner
		ipm         *speedtest.IPerfManager
		spHandler   *speedtest.Handler
	)

	if cfg.Mode == config.ModeStandalone || (cfg.Mode == config.ModeHub && cfg.Hub.EnableLocalNode) {
		ipm = speedtest.NewIPerfManager(cfg.IPerf3)
		bgpDrv := bgp.NewDriver(cfg.BGP)
		localRunner = cluster.NewLocalNode(cfg, g, false, enricher, ipm, bgpDrv)
		spHandler = speedtest.NewHandler(
			cfg.Speedtest,
			g,
			cfg.Node.ID,
			cfg.Hub.SigningSecret,
			false,
			cfg.Server.CORSOrigins,
		)
	}

	ttl := time.Duration(max(cfg.Hub.TokenTTLSeconds, 60)) * time.Second
	reg := cluster.NewRegistry(cfg.Hub.AgentToken, cfg.Hub.SigningSecret, ttl, localRunner, logger)

	s := &Server{
		cfg:       cfg,
		guard:     g,
		enricher:  enricher,
		registry:  reg,
		iperf:     ipm,
		speedtest: spHandler,
		logger:    logger,
	}

	s.metrics = metrics.New(cfg.Mode, cfg.Server.MetricsToken, func() metrics.SnapshotGauge {
		nodes := reg.ListNodes(false)
		online, offline := 0, 0
		for _, n := range nodes {
			if n.Online {
				online++
			} else {
				offline++
			}
		}
		activeIPerf := 0
		if ipm != nil {
			activeIPerf = ipm.ActiveCount()
		}
		return metrics.SnapshotGauge{
			OnlineNodes:       online,
			OfflineNodes:      offline,
			ActiveProbes:      g.ActiveProbes(),
			ActiveSpeedtests:  g.ActiveSpeedtests(),
			ActiveIPerfLeases: activeIPerf,
		}
	})

	if spHandler != nil {
		spHandler.SetTelemetry(
			func(kind string) { s.metrics.IncRequest(kind) },
			func(dir string, n int64) {
				if dir == "download" {
					s.metrics.AddDownloadBytes(n)
				} else {
					s.metrics.AddUploadBytes(n)
				}
			},
		)
	}

	mux, err := s.buildMux()
	if err != nil {
		return nil, err
	}
	s.handler = withSecurityHeaders(mux, cfg.Server.AllowEmbed)
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.handler
}

func (s *Server) Guard() *guard.Guard {
	return s.guard
}

func (s *Server) Registry() *cluster.Registry {
	return s.registry
}

func (s *Server) Close() {
	if s.iperf != nil {
		s.iperf.Close()
	}
}

func (s *Server) buildMux() (*http.ServeMux, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealth)
	if s.cfg.Server.EnableMetrics {
		mux.Handle("GET /metrics", s.metrics)
	}
	mux.HandleFunc("GET /api/v1/meta", s.handleMeta)
	mux.HandleFunc("GET /api/v1/nodes", s.handleNodes)
	mux.HandleFunc("GET /api/v1/probe/ping", s.handlePingSSE)
	mux.HandleFunc("GET /api/v1/probe/mtr", s.handleMTRSSE)
	mux.HandleFunc("GET /api/v1/probe/dig", s.handleDig)
	mux.HandleFunc("GET /api/v1/probe/compare", s.handleCompareSSE)
	mux.HandleFunc("GET /api/v1/bgp/route", s.handleBGPRoute)
	mux.HandleFunc("GET /api/v1/bgp/summary", s.handleBGPSummary)
	mux.HandleFunc("POST /api/v1/iperf3/lease", s.handleIPerfLease)

	if s.cfg.Mode == config.ModeHub {
		mux.HandleFunc("/api/v1/cluster/ws", s.registry.ServeAgentWS)
	}

	if s.speedtest != nil {
		mux.HandleFunc("/api/v1/speedtest/ping", s.speedtest.ServePing)
		mux.HandleFunc("/api/v1/speedtest/download", s.speedtest.ServeDownload)
		mux.HandleFunc("/api/v1/speedtest/upload", s.speedtest.ServeUpload)
		mux.HandleFunc("/speedtest/files/{name}", s.speedtest.ServeTestFile)
	}

	staticFS, err := web.StaticFS()
	if err != nil {
		return nil, err
	}
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	return mux, nil
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"mode":   s.cfg.Mode,
	})
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	clientIP := s.guard.ClientIP(r)
	visitor := map[string]any{}
	if clientIP.IsValid() {
		ipStr := clientIP.String()
		family := "v4"
		if clientIP.Is6() {
			family = "v6"
		}
		meta := s.enricher.Lookup(r.Context(), clientIP)
		visitor = map[string]any{
			"ip":          ipStr,
			"redacted_ip": guard.RedactIP(ipStr),
			"family":      family,
			"asn":         meta.ASN,
			"org":         meta.Org,
			"line_tag":    meta.LineTag,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"title":          s.cfg.Server.Title,
		"subtitle":       s.cfg.Server.Subtitle,
		"mode":           s.cfg.Mode,
		"default_locale": s.cfg.Server.DefaultLocale,
		"version":        version.Version,
		"allow_embed":    s.cfg.Server.AllowEmbed,
		"links": map[string]string{
			"repo_url":        s.cfg.Server.RepoURL,
			"community_url":   s.cfg.Server.CommunityURL,
			"community_label": s.cfg.Server.CommunityLabel,
			"brand_url":       s.cfg.Server.BrandURL,
			"brand_label":     s.cfg.Server.BrandLabel,
		},
		"visitor": visitor,
		"defaults": map[string]any{
			"ping_count":     s.cfg.Probe.PingDefaultCount,
			"ping_max_count": s.cfg.Probe.PingMaxCount,
			"mtr_cycles":     s.cfg.Probe.MTRDefaultCycles,
			"mtr_max_cycles": s.cfg.Probe.MTRMaxCycles,
		},
		"nodes": s.registry.ListNodes(true),
	})
}

func (s *Server) handleNodes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"nodes": s.registry.ListNodes(true),
	})
}

func (s *Server) handlePingSSE(w http.ResponseWriter, r *http.Request) {
	sse, ok := newSSEWriter(w)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	nodeID := strings.TrimSpace(q.Get("node"))
	runner, err := s.registry.GetRunner(nodeID)
	if err != nil {
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}

	s.metrics.IncRequest("ping")
	resolved, err := s.guard.ResolveAndValidateTarget(r.Context(), q.Get("target"), q.Get("ip_version"))
	if err != nil {
		s.recordGuardError(err)
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}

	clientIP := s.guard.ClientIP(r)
	release, err := s.guard.AcquireProbe(clientIP, resolved.Addr)
	if err != nil {
		s.recordGuardError(err)
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}
	defer release()

	count := clampInt(q.Get("count"), s.cfg.Probe.PingDefaultCount, 1, s.cfg.Probe.PingMaxCount)
	port := clampInt(q.Get("port"), 443, 1, 65535)
	proto := strings.ToLower(strings.TrimSpace(q.Get("protocol")))
	if proto != "tcp" {
		proto = "icmp"
	}

	sse.Send("target", map[string]any{
		"input":    resolved.Input,
		"hostname": resolved.Hostname,
		"addr":     resolved.Addr.String(),
		"node_id":  runner.Descriptor().ID,
	})

	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()

	sum, err := runner.Ping(ctx, clientIP.String(), probe.PingRequest{
		Target:     resolved.Input,
		TargetAddr: resolved.Addr,
		Protocol:   proto,
		Port:       port,
		Count:      count,
		TimeoutMs:  s.cfg.Probe.PingTimeoutMs,
		IntervalMs: s.cfg.Probe.PingIntervalMs,
	}, func(sample probe.PingSample) {
		sse.Send("sample", sample)
	})
	if err != nil {
		if ctx.Err() == nil {
			sse.Send("error_msg", map[string]string{"error": err.Error()})
		}
		return
	}
	sse.Send("summary", sum)
}

func (s *Server) handleMTRSSE(w http.ResponseWriter, r *http.Request) {
	sse, ok := newSSEWriter(w)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	runner, err := s.registry.GetRunner(strings.TrimSpace(q.Get("node")))
	if err != nil {
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}

	s.metrics.IncRequest("mtr")
	resolved, err := s.guard.ResolveAndValidateTarget(r.Context(), q.Get("target"), q.Get("ip_version"))
	if err != nil {
		s.recordGuardError(err)
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}

	clientIP := s.guard.ClientIP(r)
	release, err := s.guard.AcquireProbe(clientIP, resolved.Addr)
	if err != nil {
		s.recordGuardError(err)
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}
	defer release()

	cycles := clampInt(q.Get("cycles"), s.cfg.Probe.MTRDefaultCycles, 1, s.cfg.Probe.MTRMaxCycles)

	sse.Send("target", map[string]any{
		"input":    resolved.Input,
		"hostname": resolved.Hostname,
		"addr":     resolved.Addr.String(),
		"node_id":  runner.Descriptor().ID,
	})

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	snap, err := runner.MTR(ctx, clientIP.String(), probe.MTRRequest{
		Target:     resolved.Input,
		TargetAddr: resolved.Addr,
		Cycles:     cycles,
		MaxHops:    s.cfg.Probe.MTRMaxHops,
		TimeoutMs:  s.cfg.Probe.PingTimeoutMs,
	}, func(update probe.MTRSnapshot) {
		sse.Send("snapshot", update)
	})
	if err != nil {
		if ctx.Err() == nil {
			sse.Send("error_msg", map[string]string{"error": err.Error()})
		}
		return
	}
	sse.Send("done", snap)
}

func (s *Server) handleDig(w http.ResponseWriter, r *http.Request) {
	s.metrics.IncRequest("dig")
	q := r.URL.Query()

	name := strings.TrimSpace(q.Get("target"))
	if err := guard.ValidateDigQuery(name); err != nil {
		s.metrics.IncGuardReject("invalid_target")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	rrName, _, err := probe.NormalizeDigRRType(q.Get("rrtype"))
	if err != nil {
		s.metrics.IncGuardReject("invalid_target")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	resolver, err := probe.NormalizeDigResolver(q.Get("resolver"))
	if err != nil {
		s.metrics.IncGuardReject("invalid_target")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if resolver != "" {
		if ip, perr := netip.ParseAddr(resolverHostPort(resolver)); perr == nil {
			if aerr := s.guard.CheckAddrAllowed(ip); aerr != nil {
				s.recordGuardError(aerr)
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": aerr.Error()})
				return
			}
		}
	}

	runner, err := s.registry.GetRunner(strings.TrimSpace(q.Get("node")))
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, cluster.ErrNodeOffline) {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	req := probe.DigRequest{
		Name:      name,
		RRType:    rrName,
		Resolver:  resolver,
		TimeoutMs: s.cfg.Probe.DigTimeoutMs,
	}

	clientIP := s.guard.ClientIP(r)
	release, gerr := s.guard.AcquireProbe(clientIP, probe.DigRateTargetAddr(req))
	if gerr != nil {
		s.recordGuardError(gerr)
		status := http.StatusTooManyRequests
		if errors.Is(gerr, guard.ErrBusy) {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]string{"error": gerr.Error()})
		return
	}
	defer release()

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	res, err := runner.Dig(ctx, clientIP.String(), req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// resolverHostPort splits an already-normalized "ip:port" resolver into its host part.
func resolverHostPort(normalized string) string {
	if host, _, err := net.SplitHostPort(normalized); err == nil {
		return host
	}
	return normalized
}

func (s *Server) handleCompareSSE(w http.ResponseWriter, r *http.Request) {
	sse, ok := newSSEWriter(w)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	s.metrics.IncRequest("compare")
	resolved, err := s.guard.ResolveAndValidateTarget(r.Context(), q.Get("target"), q.Get("ip_version"))
	if err != nil {
		s.recordGuardError(err)
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}

	clientIP := s.guard.ClientIP(r)
	release, err := s.guard.AcquireProbe(clientIP, resolved.Addr)
	if err != nil {
		s.recordGuardError(err)
		sse.Send("error_msg", map[string]string{"error": err.Error()})
		return
	}
	defer release()

	var filterIDs []string
	if rawNodes := strings.TrimSpace(q.Get("nodes")); rawNodes != "" {
		filterIDs = strings.Split(rawNodes, ",")
	}
	runners := s.registry.OnlineRunners(filterIDs)
	if len(runners) == 0 {
		sse.Send("error_msg", map[string]string{"error": "no online nodes available"})
		return
	}
	if len(runners) > 16 {
		runners = runners[:16]
	}

	includeRoute := q.Get("include_route") != "false"
	proto := strings.ToLower(strings.TrimSpace(q.Get("protocol")))
	if proto != "tcp" {
		proto = "icmp"
	}
	port := clampInt(q.Get("port"), 443, 1, 65535)

	ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
	defer cancel()

	sse.Send("start", map[string]any{
		"target":     resolved.Input,
		"addr":       resolved.Addr.String(),
		"node_count": len(runners),
	})

	var wg sync.WaitGroup
	for _, runner := range runners {
		wg.Add(1)
		go func(rn cluster.NodeRunner) {
			defer wg.Done()
			nid := rn.Descriptor().ID

			sum, perr := rn.Ping(ctx, clientIP.String(), probe.PingRequest{
				Target:     resolved.Input,
				TargetAddr: resolved.Addr,
				Protocol:   proto,
				Port:       port,
				Count:      min(3, s.cfg.Probe.PingMaxCount),
				TimeoutMs:  s.cfg.Probe.PingTimeoutMs,
				IntervalMs: 200,
			}, nil)
			if perr != nil {
				if ctx.Err() == nil {
					sse.Send("node_error", map[string]string{"node_id": nid, "error": perr.Error()})
				}
				return
			}
			sse.Send("node_ping", map[string]any{
				"node_id": nid,
				"summary": sum,
			})

			if includeRoute && ctx.Err() == nil {
				mtrSnap, merr := rn.MTR(ctx, clientIP.String(), probe.MTRRequest{
					Target:     resolved.Input,
					TargetAddr: resolved.Addr,
					Cycles:     min(2, s.cfg.Probe.MTRMaxCycles),
					MaxHops:    s.cfg.Probe.MTRMaxHops,
					TimeoutMs:  min(1000, s.cfg.Probe.PingTimeoutMs),
				}, func(partial probe.MTRSnapshot) {
					sse.Send("node_route", map[string]any{
						"node_id": nid,
						"mtr":     partial,
					})
				})
				if merr == nil {
					sse.Send("node_route", map[string]any{
						"node_id": nid,
						"mtr":     mtrSnap,
					})
				}
			}
		}(runner)
	}

	wg.Wait()
	if ctx.Err() == nil {
		sse.Send("complete", map[string]bool{"completed": true})
	}
}

func (s *Server) handleIPerfLease(w http.ResponseWriter, r *http.Request) {
	s.metrics.IncRequest("iperf_lease")
	nodeID := strings.TrimSpace(r.URL.Query().Get("node"))
	runner, err := s.registry.GetRunner(nodeID)
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, cluster.ErrNodeOffline) {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	clientIP := s.guard.ClientIP(r)
	// Enforce client rate limit on iperf lease requests
	rel, gerr := s.guard.AcquireSpeedtest(clientIP, true)
	if gerr != nil {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": gerr.Error()})
		return
	}
	rel()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	lease, err := runner.AcquireIPerf(ctx, clientIP.String())
	if err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, speedtest.ErrIPerfBusy) {
			status = http.StatusTooManyRequests
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, lease)
}

func (s *Server) handleBGPRoute(w http.ResponseWriter, r *http.Request) {
	s.metrics.IncRequest("bgp_route")
	q := r.URL.Query()
	target := strings.TrimSpace(q.Get("target"))
	norm, _, err := bgp.ParseQueryTarget(target)
	if err != nil {
		s.metrics.IncGuardReject("invalid_target")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	runner, err := s.registry.GetRunner(strings.TrimSpace(q.Get("node")))
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, cluster.ErrNodeOffline) {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	clientIP := s.guard.ClientIP(r)
	rel, gerr := s.guard.AcquireSpeedtest(clientIP, true)
	if gerr != nil {
		s.recordGuardError(gerr)
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": gerr.Error()})
		return
	}
	rel()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := runner.BGPRoute(ctx, norm)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, bgp.ErrBGPDisabled) {
			status = http.StatusNotImplemented
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleBGPSummary(w http.ResponseWriter, r *http.Request) {
	s.metrics.IncRequest("bgp_summary")
	runner, err := s.registry.GetRunner(strings.TrimSpace(r.URL.Query().Get("node")))
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, cluster.ErrNodeOffline) {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	clientIP := s.guard.ClientIP(r)
	rel, gerr := s.guard.AcquireSpeedtest(clientIP, true)
	if gerr != nil {
		s.recordGuardError(gerr)
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": gerr.Error()})
		return
	}
	rel()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	res, err := runner.BGPSummary(ctx)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, bgp.ErrBGPDisabled) {
			status = http.StatusNotImplemented
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) recordGuardError(err error) {
	switch {
	case errors.Is(err, guard.ErrForbiddenTarget):
		s.metrics.IncGuardReject("ssrf_forbidden")
	case errors.Is(err, guard.ErrInvalidTarget):
		s.metrics.IncGuardReject("invalid_target")
	case errors.Is(err, guard.ErrClientRateLimited):
		s.metrics.IncGuardReject("client_rate_limit")
	case errors.Is(err, guard.ErrTargetRateLimited):
		s.metrics.IncGuardReject("target_rate_limit")
	case errors.Is(err, guard.ErrBusy):
		s.metrics.IncGuardReject("busy")
	}
}

type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	mu      sync.Mutex
}

func newSSEWriter(w http.ResponseWriter) (*sseWriter, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	return &sseWriter{w: w, flusher: flusher}, true
}

func (s *sseWriter) Send(event string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, data)
	s.flusher.Flush()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func clampInt(raw string, def, low, high int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return min(max(v, low), high)
}

func withSecurityHeaders(next http.Handler, allowEmbed bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if !allowEmbed || r.URL.Query().Get("embed") != "1" {
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		}
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
