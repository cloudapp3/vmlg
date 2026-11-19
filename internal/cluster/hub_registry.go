package cluster

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudapp3/vmlg/internal/bgp"
	"github.com/cloudapp3/vmlg/internal/probe"
	"github.com/cloudapp3/vmlg/internal/speedtest"
	"github.com/gorilla/websocket"
)

var (
	ErrNodeNotFound = errors.New("requested node does not exist")
	ErrNodeOffline  = errors.New("requested node is currently offline")
)

type wsSafeConn struct {
	conn    *websocket.Conn
	writeMu sync.Mutex
}

func (w *wsSafeConn) Send(env Envelope) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	_ = w.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return w.conn.WriteJSON(env)
}

func (w *wsSafeConn) Close() error {
	return w.conn.Close()
}

type remoteSession struct {
	ws      *wsSafeConn
	desc    NodeDescriptor
	mu      sync.Mutex
	pending map[string]chan Envelope
	closed  bool
}

func (s *remoteSession) addPending(reqID string) (chan Envelope, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, ErrNodeOffline
	}
	ch := make(chan Envelope, 32)
	s.pending[reqID] = ch
	return ch, nil
}

func (s *remoteSession) removePending(reqID string) {
	s.mu.Lock()
	delete(s.pending, reqID)
	s.mu.Unlock()
}

func (s *remoteSession) dispatch(env Envelope) {
	if env.ReqID == "" {
		return
	}
	s.mu.Lock()
	ch, ok := s.pending[env.ReqID]
	s.mu.Unlock()
	if ok {
		select {
		case ch <- env:
		default:
		}
	}
}

func (s *remoteSession) closeAll() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	for id, ch := range s.pending {
		close(ch)
		delete(s.pending, id)
	}
	s.mu.Unlock()
	_ = s.ws.Close()
}

type remoteNodeRunner struct {
	reg     *Registry
	nodeID  string
	session *remoteSession
}

func (r *remoteNodeRunner) Descriptor() NodeDescriptor {
	r.session.mu.Lock()
	d := r.session.desc
	r.session.mu.Unlock()
	return d
}

func (r *remoteNodeRunner) Ping(ctx context.Context, clientIP string, req probe.PingRequest, onSample func(probe.PingSample)) (probe.PingSummary, error) {
	reqID := r.reg.nextReqID()
	ch, err := r.session.addPending(reqID)
	if err != nil {
		return probe.PingSummary{}, err
	}
	defer r.session.removePending(reqID)

	startEnv, err := makeEnvelope(MsgProbeStart, reqID, ProbeStartPayload{
		Kind:     "ping",
		ClientIP: clientIP,
		Ping:     &req,
	}, "")
	if err != nil {
		return probe.PingSummary{}, err
	}
	if err := r.session.ws.Send(startEnv); err != nil {
		return probe.PingSummary{}, ErrNodeOffline
	}

	for {
		select {
		case <-ctx.Done():
			cancelEnv, _ := makeEnvelope(MsgProbeCancel, reqID, nil, "")
			_ = r.session.ws.Send(cancelEnv)
			return probe.PingSummary{}, ctx.Err()
		case env, ok := <-ch:
			if !ok {
				return probe.PingSummary{}, ErrNodeOffline
			}
			switch env.Type {
			case MsgProbeSample:
				var sample probe.PingSample
				if json.Unmarshal(env.Payload, &sample) == nil && onSample != nil {
					onSample(sample)
				}
			case MsgProbeDone:
				var sum probe.PingSummary
				if err := json.Unmarshal(env.Payload, &sum); err != nil {
					return probe.PingSummary{}, err
				}
				return sum, nil
			case MsgProbeError:
				return probe.PingSummary{}, errors.New(env.Error)
			}
		}
	}
}

func (r *remoteNodeRunner) MTR(ctx context.Context, clientIP string, req probe.MTRRequest, onUpdate func(probe.MTRSnapshot)) (probe.MTRSnapshot, error) {
	reqID := r.reg.nextReqID()
	ch, err := r.session.addPending(reqID)
	if err != nil {
		return probe.MTRSnapshot{}, err
	}
	defer r.session.removePending(reqID)

	startEnv, err := makeEnvelope(MsgProbeStart, reqID, ProbeStartPayload{
		Kind:     "mtr",
		ClientIP: clientIP,
		MTR:      &req,
	}, "")
	if err != nil {
		return probe.MTRSnapshot{}, err
	}
	if err := r.session.ws.Send(startEnv); err != nil {
		return probe.MTRSnapshot{}, ErrNodeOffline
	}

	for {
		select {
		case <-ctx.Done():
			cancelEnv, _ := makeEnvelope(MsgProbeCancel, reqID, nil, "")
			_ = r.session.ws.Send(cancelEnv)
			return probe.MTRSnapshot{}, ctx.Err()
		case env, ok := <-ch:
			if !ok {
				return probe.MTRSnapshot{}, ErrNodeOffline
			}
			switch env.Type {
			case MsgProbeMTR:
				var snap probe.MTRSnapshot
				if json.Unmarshal(env.Payload, &snap) == nil && onUpdate != nil {
					onUpdate(snap)
				}
			case MsgProbeDone:
				var snap probe.MTRSnapshot
				if err := json.Unmarshal(env.Payload, &snap); err != nil {
					return probe.MTRSnapshot{}, err
				}
				return snap, nil
			case MsgProbeError:
				return probe.MTRSnapshot{}, errors.New(env.Error)
			}
		}
	}
}

func (r *remoteNodeRunner) Dig(ctx context.Context, clientIP string, req probe.DigRequest) (probe.DigResult, error) {
	reqID := r.reg.nextReqID()
	ch, err := r.session.addPending(reqID)
	if err != nil {
		return probe.DigResult{}, err
	}
	defer r.session.removePending(reqID)

	env, err := makeEnvelope(MsgProbeStart, reqID, ProbeStartPayload{
		Kind:     "dig",
		ClientIP: clientIP,
		Dig:      &req,
	}, "")
	if err != nil {
		return probe.DigResult{}, err
	}
	if err := r.session.ws.Send(env); err != nil {
		return probe.DigResult{}, ErrNodeOffline
	}

	select {
	case <-ctx.Done():
		cancelEnv, _ := makeEnvelope(MsgProbeCancel, reqID, nil, "")
		_ = r.session.ws.Send(cancelEnv)
		return probe.DigResult{}, ctx.Err()
	case resp, ok := <-ch:
		if !ok {
			return probe.DigResult{}, ErrNodeOffline
		}
		if resp.Error != "" {
			return probe.DigResult{}, errors.New(resp.Error)
		}
		var res probe.DigResult
		if err := json.Unmarshal(resp.Payload, &res); err != nil {
			return probe.DigResult{}, err
		}
		return res, nil
	}
}

func (r *remoteNodeRunner) AcquireIPerf(ctx context.Context, clientKey string) (speedtest.IPerfLease, error) {
	reqID := r.reg.nextReqID()
	ch, err := r.session.addPending(reqID)
	if err != nil {
		return speedtest.IPerfLease{}, err
	}
	defer r.session.removePending(reqID)

	env, err := makeEnvelope(MsgIPerfAcquire, reqID, IPerfRequestPayload{ClientKey: clientKey}, "")
	if err != nil {
		return speedtest.IPerfLease{}, err
	}
	if err := r.session.ws.Send(env); err != nil {
		return speedtest.IPerfLease{}, ErrNodeOffline
	}

	select {
	case <-ctx.Done():
		return speedtest.IPerfLease{}, ctx.Err()
	case resp, ok := <-ch:
		if !ok {
			return speedtest.IPerfLease{}, ErrNodeOffline
		}
		if resp.Error != "" {
			return speedtest.IPerfLease{}, errors.New(resp.Error)
		}
		var lease speedtest.IPerfLease
		if err := json.Unmarshal(resp.Payload, &lease); err != nil {
			return speedtest.IPerfLease{}, err
		}
		return lease, nil
	}
}

func (r *remoteNodeRunner) BGPRoute(ctx context.Context, target string) (bgp.RouteLookupResult, error) {
	reqID := r.reg.nextReqID()
	ch, err := r.session.addPending(reqID)
	if err != nil {
		return bgp.RouteLookupResult{}, err
	}
	defer r.session.removePending(reqID)

	env, err := makeEnvelope(MsgBGPRoute, reqID, BGPRoutePayload{Target: target}, "")
	if err != nil {
		return bgp.RouteLookupResult{}, err
	}
	if err := r.session.ws.Send(env); err != nil {
		return bgp.RouteLookupResult{}, ErrNodeOffline
	}

	select {
	case <-ctx.Done():
		return bgp.RouteLookupResult{}, ctx.Err()
	case resp, ok := <-ch:
		if !ok {
			return bgp.RouteLookupResult{}, ErrNodeOffline
		}
		if resp.Error != "" {
			return bgp.RouteLookupResult{}, errors.New(resp.Error)
		}
		var res bgp.RouteLookupResult
		if err := json.Unmarshal(resp.Payload, &res); err != nil {
			return bgp.RouteLookupResult{}, err
		}
		return res, nil
	}
}

func (r *remoteNodeRunner) BGPSummary(ctx context.Context) (bgp.SummaryResult, error) {
	reqID := r.reg.nextReqID()
	ch, err := r.session.addPending(reqID)
	if err != nil {
		return bgp.SummaryResult{}, err
	}
	defer r.session.removePending(reqID)

	env, err := makeEnvelope(MsgBGPSummary, reqID, nil, "")
	if err != nil {
		return bgp.SummaryResult{}, err
	}
	if err := r.session.ws.Send(env); err != nil {
		return bgp.SummaryResult{}, ErrNodeOffline
	}

	select {
	case <-ctx.Done():
		return bgp.SummaryResult{}, ctx.Err()
	case resp, ok := <-ch:
		if !ok {
			return bgp.SummaryResult{}, ErrNodeOffline
		}
		if resp.Error != "" {
			return bgp.SummaryResult{}, errors.New(resp.Error)
		}
		var res bgp.SummaryResult
		if err := json.Unmarshal(resp.Payload, &res); err != nil {
			return bgp.SummaryResult{}, err
		}
		return res, nil
	}
}

type Registry struct {
	agentToken    string
	signingSecret string
	ticketTTL     time.Duration
	logger        *slog.Logger
	localNode     NodeRunner

	mu        sync.RWMutex
	sessions  map[string]*remoteSession
	lastKnown map[string]NodeDescriptor

	reqSeq atomic.Uint64
}

func NewRegistry(agentToken, signingSecret string, ticketTTL time.Duration, localNode NodeRunner, logger *slog.Logger) *Registry {
	if logger == nil {
		logger = slog.Default()
	}
	if ticketTTL <= 0 {
		ticketTTL = 5 * time.Minute
	}
	return &Registry{
		agentToken:    agentToken,
		signingSecret: signingSecret,
		ticketTTL:     ticketTTL,
		logger:        logger,
		localNode:     localNode,
		sessions:      make(map[string]*remoteSession),
		lastKnown:     make(map[string]NodeDescriptor),
	}
}

func (r *Registry) nextReqID() string {
	return fmt.Sprintf("req-%d", r.reqSeq.Add(1))
}

func (r *Registry) ListNodes(withTicket bool) []NodeDescriptor {
	r.mu.RLock()
	out := make([]NodeDescriptor, 0, len(r.lastKnown)+1)
	if r.localNode != nil {
		d := r.localNode.Descriptor()
		if withTicket && r.signingSecret != "" {
			d.SpeedtestTicket = speedtest.SignTicket(r.signingSecret, d.ID, r.ticketTTL)
		}
		out = append(out, d)
	}
	for id, known := range r.lastKnown {
		if r.localNode != nil && id == r.localNode.Descriptor().ID {
			continue
		}
		if sess, online := r.sessions[id]; online {
			sess.mu.Lock()
			d := sess.desc
			sess.mu.Unlock()
			d.Online = true
			if withTicket && r.signingSecret != "" && d.Capabilities.Speedtest {
				d.SpeedtestTicket = speedtest.SignTicket(r.signingSecret, d.ID, r.ticketTTL)
			}
			out = append(out, d)
		} else {
			d := known
			d.Online = false
			d.ActiveProbes = 0
			d.ActiveSpeedtests = 0
			out = append(out, d)
		}
	}
	r.mu.RUnlock()

	sort.Slice(out, func(i, j int) bool {
		if out[i].IsLocal != out[j].IsLocal {
			return out[i].IsLocal
		}
		if out[i].Region != out[j].Region {
			return out[i].Region < out[j].Region
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (r *Registry) GetRunner(nodeID string) (NodeRunner, error) {
	if nodeID == "" && r.localNode != nil {
		return r.localNode, nil
	}
	if r.localNode != nil && r.localNode.Descriptor().ID == nodeID {
		return r.localNode, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if nodeID == "" {
		// Pick first online remote node when Hub has no local node
		for id, sess := range r.sessions {
			return &remoteNodeRunner{reg: r, nodeID: id, session: sess}, nil
		}
		return nil, ErrNodeOffline
	}

	if sess, ok := r.sessions[nodeID]; ok {
		return &remoteNodeRunner{reg: r, nodeID: nodeID, session: sess}, nil
	}
	if _, known := r.lastKnown[nodeID]; known {
		return nil, ErrNodeOffline
	}
	return nil, ErrNodeNotFound
}

func (r *Registry) OnlineRunners(filterIDs []string) []NodeRunner {
	allow := make(map[string]bool, len(filterIDs))
	for _, id := range filterIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			allow[id] = true
		}
	}

	var runners []NodeRunner
	if r.localNode != nil {
		id := r.localNode.Descriptor().ID
		if len(allow) == 0 || allow[id] {
			runners = append(runners, r.localNode)
		}
	}

	r.mu.RLock()
	ids := make([]string, 0, len(r.sessions))
	for id := range r.sessions {
		if len(allow) == 0 || allow[id] {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		runners = append(runners, &remoteNodeRunner{
			reg:     r,
			nodeID:  id,
			session: r.sessions[id],
		})
	}
	r.mu.RUnlock()
	return runners
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin:     func(_ *http.Request) bool { return true },
}

func (r *Registry) ServeAgentWS(w http.ResponseWriter, req *http.Request) {
	if r.agentToken == "" {
		http.Error(w, "cluster hub not enabled", http.StatusNotFound)
		return
	}

	gotToken := strings.TrimSpace(strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer"))
	if subtle.ConstantTimeCompare([]byte(gotToken), []byte(r.agentToken)) != 1 {
		http.Error(w, "unauthorized agent token", http.StatusUnauthorized)
		return
	}

	rawConn, err := wsUpgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	rawConn.SetReadLimit(1 << 20)
	_ = rawConn.SetReadDeadline(time.Now().Add(45 * time.Second))
	rawConn.SetPongHandler(func(string) error {
		return rawConn.SetReadDeadline(time.Now().Add(45 * time.Second))
	})

	ws := &wsSafeConn{conn: rawConn}

	// First frame must be MsgRegister
	var regEnv Envelope
	if err := rawConn.ReadJSON(&regEnv); err != nil || regEnv.Type != MsgRegister {
		_ = ws.Close()
		return
	}
	var desc NodeDescriptor
	if err := json.Unmarshal(regEnv.Payload, &desc); err != nil || desc.ID == "" {
		_ = ws.Close()
		return
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)
	desc.IsLocal = false
	desc.Online = true
	desc.ConnectedAt = nowStr
	desc.LastSeen = nowStr

	sess := &remoteSession{
		ws:      ws,
		desc:    desc,
		pending: make(map[string]chan Envelope),
	}

	r.mu.Lock()
	if prev, exists := r.sessions[desc.ID]; exists {
		go prev.closeAll()
	}
	r.sessions[desc.ID] = sess
	r.lastKnown[desc.ID] = desc
	r.mu.Unlock()

	r.logger.Info("agent registered", "node_id", desc.ID, "name", desc.Name, "region", desc.Region)

	ackEnv, _ := makeEnvelope(MsgRegistered, "", map[string]string{"status": "ok"}, "")
	_ = ws.Send(ackEnv)

	defer func() {
		sess.closeAll()
		r.mu.Lock()
		if r.sessions[desc.ID] == sess {
			delete(r.sessions, desc.ID)
		}
		r.mu.Unlock()
		r.logger.Info("agent disconnected", "node_id", desc.ID)
	}()

	for {
		var env Envelope
		if err := rawConn.ReadJSON(&env); err != nil {
			return
		}
		_ = rawConn.SetReadDeadline(time.Now().Add(45 * time.Second))

		switch env.Type {
		case MsgHeartbeat:
			var hb HeartbeatPayload
			_ = json.Unmarshal(env.Payload, &hb)
			sess.mu.Lock()
			sess.desc.LastSeen = time.Now().UTC().Format(time.RFC3339)
			sess.desc.ActiveProbes = hb.ActiveProbes
			sess.desc.ActiveSpeedtests = hb.ActiveSpeedtests
			updated := sess.desc
			sess.mu.Unlock()

			r.mu.Lock()
			r.lastKnown[desc.ID] = updated
			r.mu.Unlock()

		case MsgProbeSample, MsgProbeMTR, MsgProbeDone, MsgProbeError, MsgIPerfResult, MsgBGPResult:
			sess.dispatch(env)
		}
	}
}
