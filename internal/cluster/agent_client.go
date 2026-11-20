package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/guard"
	"github.com/cloudapp3/vmlg/internal/probe"
	"github.com/cloudapp3/vmlg/internal/speedtest"
	"github.com/gorilla/websocket"
)

type AgentClient struct {
	cfg       config.Config
	localNode *LocalNode
	guard     *guard.Guard
	logger    *slog.Logger
}

func NewAgentClient(cfg config.Config, localNode *LocalNode, g *guard.Guard, logger *slog.Logger) *AgentClient {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentClient{
		cfg:       cfg,
		localNode: localNode,
		guard:     g,
		logger:    logger,
	}
}

func (a *AgentClient) Run(ctx context.Context) error {
	var speedSrv *http.Server
	if a.cfg.Agent.SpeedtestListen != "" && a.cfg.Speedtest.Enabled {
		speedSrv = a.startSpeedtestHTTPServer()
		defer func() {
			sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = speedSrv.Shutdown(sctx)
		}()
	}

	wsURL, err := normalizeHubWSURL(a.cfg.Agent.HubURL)
	if err != nil {
		return err
	}

	backoff := time.Second
	maxBackoff := time.Duration(max(a.cfg.Agent.ReconnectMaxSeconds, 5)) * time.Second

	for {
		if ctx.Err() != nil {
			return nil
		}

		connectedAt := time.Now()
		err := a.connectAndServeOnce(ctx, wsURL)
		if ctx.Err() != nil {
			return nil
		}

		if time.Since(connectedAt) > 30*time.Second {
			backoff = time.Second
		}
		a.logger.Warn("hub tunnel disconnected, retrying", "err", err, "backoff", backoff.String())

		jitter := time.Duration(rand.IntN(400)) * time.Millisecond
		timer := time.NewTimer(backoff + jitter)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

func (a *AgentClient) startSpeedtestHTTPServer() *http.Server {
	mux := http.NewServeMux()
	sp := speedtest.NewHandler(
		a.cfg.Speedtest,
		a.guard,
		a.cfg.Node.ID,
		a.cfg.Agent.SigningSecret,
		a.cfg.Agent.RequireSignedSpeed,
		a.cfg.Server.CORSOrigins,
	)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/speedtest/ping", sp.ServePing)
	mux.HandleFunc("/api/v1/speedtest/download", sp.ServeDownload)
	mux.HandleFunc("/api/v1/speedtest/upload", sp.ServeUpload)
	mux.HandleFunc("/speedtest/files/{name}", sp.ServeTestFile)

	srv := &http.Server{
		Addr:              a.cfg.Agent.SpeedtestListen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		a.logger.Info("agent direct speedtest HTTP server listening", "addr", a.cfg.Agent.SpeedtestListen)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Error("agent speedtest server error", "err", err)
		}
	}()
	return srv
}

const AgentReadyFile = "/tmp/vmlg-agent.ready"

func touchAgentReady() {
	_ = os.WriteFile(AgentReadyFile, []byte("ok\n"), 0o644)
}

func clearAgentReady() {
	_ = os.Remove(AgentReadyFile)
}

func (a *AgentClient) connectAndServeOnce(ctx context.Context, wsURL string) error {
	hdr := http.Header{}
	hdr.Set("Authorization", "Bearer "+a.cfg.Agent.Token)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	rawConn, _, err := dialer.DialContext(ctx, wsURL, hdr)
	if err != nil {
		return err
	}
	rawConn.SetReadLimit(1 << 20)
	ws := &wsSafeConn{conn: rawConn}

	connCtx, cancelConn := context.WithCancel(ctx)
	defer func() {
		cancelConn()
		_ = ws.Close()
	}()

	// Send registration
	regEnv, err := makeEnvelope(MsgRegister, "", a.localNode.Descriptor(), "")
	if err != nil {
		return err
	}
	if err := ws.Send(regEnv); err != nil {
		return err
	}

	var ack Envelope
	_ = rawConn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err := rawConn.ReadJSON(&ack); err != nil || ack.Type != MsgRegistered {
		if err != nil {
			return err
		}
		return errors.New("unexpected hub handshake response")
	}
	_ = rawConn.SetReadDeadline(time.Time{})

	a.logger.Info("connected to hub", "hub_url", wsURL, "node_id", a.cfg.Node.ID)
	touchAgentReady()
	defer clearAgentReady()

	var jobsMu sync.Mutex
	activeJobs := make(map[string]context.CancelFunc)

	defer func() {
		jobsMu.Lock()
		for _, cancel := range activeJobs {
			cancel()
		}
		jobsMu.Unlock()
	}()

	// Periodic heartbeat + WS ping
	hbInterval := time.Duration(max(a.cfg.Agent.HeartbeatSeconds, 5)) * time.Second
	go func() {
		ticker := time.NewTicker(hbInterval)
		defer ticker.Stop()
		for {
			select {
			case <-connCtx.Done():
				return
			case <-ticker.C:
				desc := a.localNode.Descriptor()
				env, _ := makeEnvelope(MsgHeartbeat, "", HeartbeatPayload{
					ActiveProbes:     desc.ActiveProbes,
					ActiveSpeedtests: desc.ActiveSpeedtests,
				}, "")
				if err := ws.Send(env); err != nil {
					cancelConn()
					return
				}
				touchAgentReady()
			}
		}
	}()

	go func() {
		<-connCtx.Done()
		_ = ws.Close()
	}()

	for {
		var env Envelope
		if err := rawConn.ReadJSON(&env); err != nil {
			return err
		}

		switch env.Type {
		case MsgProbeStart:
			var p ProbeStartPayload
			if err := json.Unmarshal(env.Payload, &p); err != nil {
				continue
			}
			jobCtx, jobCancel := context.WithCancel(connCtx)
			jobsMu.Lock()
			activeJobs[env.ReqID] = jobCancel
			jobsMu.Unlock()

			go func(reqID string, payload ProbeStartPayload) {
				defer func() {
					jobCancel()
					jobsMu.Lock()
					delete(activeJobs, reqID)
					jobsMu.Unlock()
				}()
				a.handleProbeJob(jobCtx, ws, reqID, payload)
			}(env.ReqID, p)

		case MsgProbeCancel:
			jobsMu.Lock()
			if cancel, ok := activeJobs[env.ReqID]; ok {
				cancel()
			}
			jobsMu.Unlock()

		case MsgIPerfAcquire:
			var p IPerfRequestPayload
			_ = json.Unmarshal(env.Payload, &p)
			go func(reqID, clientKey string) {
				lease, err := a.localNode.AcquireIPerf(connCtx, clientKey)
				if err != nil {
					resp, _ := makeEnvelope(MsgIPerfResult, reqID, nil, err.Error())
					_ = ws.Send(resp)
					return
				}
				resp, _ := makeEnvelope(MsgIPerfResult, reqID, lease, "")
				_ = ws.Send(resp)
			}(env.ReqID, p.ClientKey)

		case MsgBGPRoute:
			var p BGPRoutePayload
			_ = json.Unmarshal(env.Payload, &p)
			go func(reqID, target string) {
				res, err := a.localNode.BGPRoute(connCtx, target)
				if err != nil {
					resp, _ := makeEnvelope(MsgBGPResult, reqID, nil, err.Error())
					_ = ws.Send(resp)
					return
				}
				resp, _ := makeEnvelope(MsgBGPResult, reqID, res, "")
				_ = ws.Send(resp)
			}(env.ReqID, p.Target)

		case MsgBGPSummary:
			go func(reqID string) {
				res, err := a.localNode.BGPSummary(connCtx)
				if err != nil {
					resp, _ := makeEnvelope(MsgBGPResult, reqID, nil, err.Error())
					_ = ws.Send(resp)
					return
				}
				resp, _ := makeEnvelope(MsgBGPResult, reqID, res, "")
				_ = ws.Send(resp)
			}(env.ReqID)
		}
	}
}

func (a *AgentClient) handleProbeJob(ctx context.Context, ws *wsSafeConn, reqID string, p ProbeStartPayload) {
	switch p.Kind {
	case "ping":
		if p.Ping == nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, "missing ping request")
			_ = ws.Send(errEnv)
			return
		}
		sum, err := a.localNode.Ping(ctx, p.ClientIP, *p.Ping, func(s probe.PingSample) {
			ev, _ := makeEnvelope(MsgProbeSample, reqID, s, "")
			_ = ws.Send(ev)
		})
		if err != nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, err.Error())
			_ = ws.Send(errEnv)
			return
		}
		doneEnv, _ := makeEnvelope(MsgProbeDone, reqID, sum, "")
		_ = ws.Send(doneEnv)

	case "mtr":
		if p.MTR == nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, "missing mtr request")
			_ = ws.Send(errEnv)
			return
		}
		snap, err := a.localNode.MTR(ctx, p.ClientIP, *p.MTR, func(s probe.MTRSnapshot) {
			ev, _ := makeEnvelope(MsgProbeMTR, reqID, s, "")
			_ = ws.Send(ev)
		})
		if err != nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, err.Error())
			_ = ws.Send(errEnv)
			return
		}
		doneEnv, _ := makeEnvelope(MsgProbeDone, reqID, snap, "")
		_ = ws.Send(doneEnv)

	case "dig":
		if p.Dig == nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, "missing dig request")
			_ = ws.Send(errEnv)
			return
		}
		res, err := a.localNode.Dig(ctx, p.ClientIP, *p.Dig)
		if err != nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, err.Error())
			_ = ws.Send(errEnv)
			return
		}
		doneEnv, _ := makeEnvelope(MsgProbeDone, reqID, res, "")
		_ = ws.Send(doneEnv)

	case "http":
		if p.HTTP == nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, "missing http request")
			_ = ws.Send(errEnv)
			return
		}
		res, err := a.localNode.HTTP(ctx, p.ClientIP, *p.HTTP)
		if err != nil {
			errEnv, _ := makeEnvelope(MsgProbeError, reqID, nil, err.Error())
			_ = ws.Send(errEnv)
			return
		}
		doneEnv, _ := makeEnvelope(MsgProbeDone, reqID, res, "")
		_ = ws.Send(doneEnv)
	}
}

func normalizeHubWSURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	case "ws", "wss":
	default:
		return "", errors.New("hub_url scheme must be ws, wss, http, or https")
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = "/api/v1/cluster/ws"
	}
	return u.String(), nil
}
