package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/server"
	"github.com/cloudapp3/vmlg/internal/version"
)

const (
	protocolVersion = "2025-06-18"
	serverName      = "vmlg"
)

type handlerTransport struct {
	h http.Handler
}

func (t handlerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	if req.RemoteAddr == "" {
		req.RemoteAddr = "127.0.0.1:0"
	}
	t.h.ServeHTTP(rec, req)
	return rec.Result(), nil
}

type Server struct {
	baseURL  string
	client   *http.Client
	localSrv *server.Server
	out      io.Writer
	errOut   io.Writer
	writeMu  sync.Mutex
}

func NewLocal(cfg config.Config, stdout, stderr io.Writer) (*Server, error) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv, err := server.New(cfg, logger)
	if err != nil {
		return nil, err
	}
	if stderr == nil {
		stderr = io.Discard
	}
	return &Server{
		baseURL:  "http://vmlg.local",
		client:   &http.Client{Transport: handlerTransport{h: srv.Handler()}, Timeout: 65 * time.Second},
		localSrv: srv,
		out:      stdout,
		errOut:   stderr,
	}, nil
}

func NewRemote(hubHTTP string, stdout, stderr io.Writer) (*Server, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(hubHTTP), "/"))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("invalid hub HTTP URL %q", hubHTTP)
	}
	if stderr == nil {
		stderr = io.Discard
	}
	return &Server{
		baseURL: u.String(),
		client:  &http.Client{Timeout: 65 * time.Second},
		out:     stdout,
		errOut:  stderr,
	}, nil
}

func (s *Server) Close() {
	if s.localSrv != nil {
		s.localSrv.Close()
	}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *Server) Serve(ctx context.Context, r io.Reader) error {
	if r == nil || s.out == nil {
		return errors.New("mcp stdio streams must not be nil")
	}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		s.handleLine(ctx, []byte(line))
	}
	return scanner.Err()
}

func (s *Server) handleLine(ctx context.Context, raw []byte) {
	var req rpcRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		s.writeError(json.RawMessage("null"), -32700, "parse error")
		return
	}
	// Notifications have no ID and produce no response.
	if len(req.ID) == 0 {
		return
	}

	switch req.Method {
	case "initialize":
		s.writeResult(req.ID, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    serverName,
				"version": version.Version,
			},
		})
	case "ping":
		s.writeResult(req.ID, map[string]any{})
	case "tools/list":
		s.writeResult(req.ID, map[string]any{"tools": toolDefinitions()})
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			s.writeError(req.ID, -32602, "invalid tools/call params")
			return
		}
		res, isErr := s.callTool(ctx, params.Name, params.Arguments)
		s.writeResult(req.ID, res.toMCP(isErr))
	default:
		s.writeError(req.ID, -32601, "method not found: "+req.Method)
	}
}

type toolResult struct {
	Text       string
	Structured any
}

func (tr toolResult) toMCP(isError bool) map[string]any {
	out := map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": tr.Text},
		},
		"isError": isError,
	}
	if tr.Structured != nil {
		out["structuredContent"] = tr.Structured
	}
	return out
}

func (s *Server) callTool(ctx context.Context, name string, rawArgs json.RawMessage) (toolResult, bool) {
	var args map[string]any
	if len(rawArgs) > 0 {
		_ = json.Unmarshal(rawArgs, &args)
	}
	strArg := func(k string) string {
		if v, ok := args[k].(string); ok {
			return strings.TrimSpace(v)
		}
		return ""
	}
	intArg := func(k string, def int) int {
		if v, ok := args[k].(float64); ok {
			return int(v)
		}
		return def
	}

	switch name {
	case "vmlg_list_nodes":
		var doc map[string]any
		if err := s.getJSON(ctx, "/api/v1/nodes", &doc); err != nil {
			return toolResult{Text: err.Error()}, true
		}
		b, _ := json.MarshalIndent(doc, "", "  ")
		return toolResult{Text: string(b), Structured: doc}, false

	case "vmlg_ping":
		target := strArg("target")
		if target == "" {
			return toolResult{Text: "target is required"}, true
		}
		q := url.Values{}
		q.Set("target", target)
		if v := strArg("node"); v != "" {
			q.Set("node", v)
		}
		if v := strArg("protocol"); v != "" {
			q.Set("protocol", v)
		}
		if v := strArg("ip_version"); v != "" {
			q.Set("ip_version", v)
		}
		if c := intArg("count", 0); c > 0 {
			q.Set("count", strconv.Itoa(c))
		}
		if p := intArg("port", 0); p > 0 {
			q.Set("port", strconv.Itoa(p))
		}

		events, err := s.readSSE(ctx, "/api/v1/probe/ping?"+q.Encode())
		if err != nil {
			return toolResult{Text: err.Error()}, true
		}
		if errPayload, ok := events["error_msg"]; ok {
			return toolResult{Text: string(errPayload[0])}, true
		}
		if sumList := events["summary"]; len(sumList) > 0 {
			var sum map[string]any
			_ = json.Unmarshal(sumList[len(sumList)-1], &sum)
			b, _ := json.MarshalIndent(sum, "", "  ")
			return toolResult{Text: string(b), Structured: sum}, false
		}
		return toolResult{Text: "no ping summary returned"}, true

	case "vmlg_mtr":
		target := strArg("target")
		if target == "" {
			return toolResult{Text: "target is required"}, true
		}
		q := url.Values{}
		q.Set("target", target)
		if v := strArg("node"); v != "" {
			q.Set("node", v)
		}
		if v := strArg("ip_version"); v != "" {
			q.Set("ip_version", v)
		}
		if c := intArg("cycles", 0); c > 0 {
			q.Set("cycles", strconv.Itoa(c))
		}

		events, err := s.readSSE(ctx, "/api/v1/probe/mtr?"+q.Encode())
		if err != nil {
			return toolResult{Text: err.Error()}, true
		}
		if errPayload, ok := events["error_msg"]; ok {
			return toolResult{Text: string(errPayload[0])}, true
		}
		for _, evName := range []string{"done", "snapshot"} {
			if list := events[evName]; len(list) > 0 {
				var snap map[string]any
				_ = json.Unmarshal(list[len(list)-1], &snap)
				b, _ := json.MarshalIndent(snap, "", "  ")
				return toolResult{Text: string(b), Structured: snap}, false
			}
		}
		return toolResult{Text: "no MTR snapshot returned"}, true

	case "vmlg_dig":
		target := strArg("target")
		if target == "" {
			return toolResult{Text: "target is required"}, true
		}
		q := url.Values{}
		q.Set("target", target)
		if v := strArg("rrtype"); v != "" {
			q.Set("rrtype", v)
		}
		if v := strArg("resolver"); v != "" {
			q.Set("resolver", v)
		}
		if v := strArg("node"); v != "" {
			q.Set("node", v)
		}

		var res map[string]any
		if err := s.getJSON(ctx, "/api/v1/probe/dig?"+q.Encode(), &res); err != nil {
			return toolResult{Text: err.Error()}, true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return toolResult{Text: string(b), Structured: res}, false

	case "vmlg_http":
		target := strArg("url")
		if target == "" {
			return toolResult{Text: "url is required"}, true
		}
		q := url.Values{}
		q.Set("url", target)
		if v := strArg("node"); v != "" {
			q.Set("node", v)
		}

		var res map[string]any
		if err := s.getJSON(ctx, "/api/v1/probe/http?"+q.Encode(), &res); err != nil {
			return toolResult{Text: err.Error()}, true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return toolResult{Text: string(b), Structured: res}, false

	case "vmlg_compare":
		target := strArg("target")
		if target == "" {
			return toolResult{Text: "target is required"}, true
		}
		q := url.Values{}
		q.Set("target", target)
		if v := strArg("ip_version"); v != "" {
			q.Set("ip_version", v)
		}
		if v, ok := args["include_route"].(bool); ok && !v {
			q.Set("include_route", "false")
		}
		if v := strArg("protocol"); v != "" {
			q.Set("protocol", v)
		}
		if p := intArg("port", 0); p > 0 {
			q.Set("port", strconv.Itoa(p))
		}

		events, err := s.readSSE(ctx, "/api/v1/probe/compare?"+q.Encode())
		if err != nil {
			return toolResult{Text: err.Error()}, true
		}
		if errPayload, ok := events["error_msg"]; ok {
			return toolResult{Text: string(errPayload[0])}, true
		}

		byNode := map[string]map[string]any{}
		ensure := func(nid string) map[string]any {
			if m, ok := byNode[nid]; ok {
				return m
			}
			m := map[string]any{"node_id": nid}
			byNode[nid] = m
			return m
		}
		for _, raw := range events["node_ping"] {
			var p struct {
				NodeID  string `json:"node_id"`
				Summary any    `json:"summary"`
			}
			if json.Unmarshal(raw, &p) == nil {
				ensure(p.NodeID)["ping"] = p.Summary
			}
		}
		for _, raw := range events["node_route"] {
			var r struct {
				NodeID string `json:"node_id"`
				MTR    any    `json:"mtr"`
			}
			if json.Unmarshal(raw, &r) == nil {
				ensure(r.NodeID)["mtr"] = r.MTR
			}
		}
		for _, raw := range events["node_error"] {
			var e struct {
				NodeID string `json:"node_id"`
				Error  string `json:"error"`
			}
			if json.Unmarshal(raw, &e) == nil {
				ensure(e.NodeID)["error"] = e.Error
			}
		}

		structured := map[string]any{"target": target, "results": byNode}
		b, _ := json.MarshalIndent(structured, "", "  ")
		return toolResult{Text: string(b), Structured: structured}, false

	case "vmlg_bgp_route":
		target := strArg("target")
		if target == "" {
			return toolResult{Text: "target IP or CIDR prefix is required"}, true
		}
		q := url.Values{}
		q.Set("target", target)
		if v := strArg("node"); v != "" {
			q.Set("node", v)
		}
		var doc map[string]any
		if err := s.getJSON(ctx, "/api/v1/bgp/route?"+q.Encode(), &doc); err != nil {
			return toolResult{Text: err.Error()}, true
		}
		b, _ := json.MarshalIndent(doc, "", "  ")
		return toolResult{Text: string(b), Structured: doc}, false

	default:
		return toolResult{Text: "unknown tool: " + name}, true
	}
}

func (s *Server) getJSON(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, dst)
}

func (s *Server) readSSE(ctx context.Context, path string) (map[string][]json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	events := make(map[string][]json.RawMessage)
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	var curEvent string
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "event:") {
			curEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if curEvent != "" && data != "" {
				events[curEvent] = append(events[curEvent], json.RawMessage(data))
			}
		}
	}
	return events, sc.Err()
}

func (s *Server) writeResult(id json.RawMessage, result any) {
	s.writeJSON(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (s *Server) writeError(id json.RawMessage, code int, msg string) {
	s.writeJSON(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	})
}

func (s *Server) writeJSON(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, _ = s.out.Write(append(b, '\n'))
}

func toolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"name":        "vmlg_list_nodes",
			"description": "List all Looking Glass nodes, their online status, regions, and capabilities (Ping, MTR, Speedtest, iPerf3, BGP).",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			"name":        "vmlg_ping",
			"description": "Run ICMP or TCP Ping from a Looking Glass node to a target IP or hostname.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target":     map[string]any{"type": "string", "description": "Target IPv4, IPv6, or hostname"},
					"node":       map[string]any{"type": "string", "description": "Optional node ID"},
					"protocol":   map[string]any{"type": "string", "enum": []string{"icmp", "tcp"}},
					"port":       map[string]any{"type": "integer", "description": "TCP port when protocol is tcp (default 443)"},
					"ip_version": map[string]any{"type": "string", "enum": []string{"auto", "v4", "v6"}},
					"count":      map[string]any{"type": "integer", "description": "Probe packet count (1-10)"},
				},
				"required": []string{"target"},
			},
		},
		{
			"name":        "vmlg_mtr",
			"description": "Run multi-cycle MTR route trace with ASN, carrier backbone (CN2 GIA / CU 9929 / CMIN2 / 163), and IXP classification.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target":     map[string]any{"type": "string", "description": "Target IPv4, IPv6, or hostname"},
					"node":       map[string]any{"type": "string", "description": "Optional node ID"},
					"ip_version": map[string]any{"type": "string", "enum": []string{"auto", "v4", "v6"}},
					"cycles":     map[string]any{"type": "integer", "description": "MTR probe cycles (1-10)"},
				},
				"required": []string{"target"},
			},
		},
		{
			"name":        "vmlg_dig",
			"description": "Run a DNS query (dig) from a Looking Glass node: full answer records with TTLs and CNAME chains via the node's system resolver or a user-specified resolver.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target":   map[string]any{"type": "string", "description": "Domain name to query, or an IP for reverse PTR lookups"},
					"rrtype":   map[string]any{"type": "string", "enum": []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "PTR"}, "description": "Record type (default A)"},
					"resolver": map[string]any{"type": "string", "description": "Optional resolver IP literal (e.g. 1.1.1.1); defaults to the node's system resolver"},
					"node":     map[string]any{"type": "string", "description": "Optional node ID"},
				},
				"required": []string{"target"},
			},
		},
		{
			"name":        "vmlg_http",
			"description": "Probe an HTTP/HTTPS URL from a Looking Glass node: status code, TTFB, redirect chain, TLS certificate summary, and response headers. Response body is never returned.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url":  map[string]any{"type": "string", "description": "Absolute http:// or https:// URL to probe"},
					"node": map[string]any{"type": "string", "description": "Optional node ID"},
				},
				"required": []string{"url"},
			},
		},
		{
			"name":        "vmlg_compare",
			"description": "Broadcast concurrent Ping and return-route classification across all online Looking Glass nodes to compare regional latency.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target":        map[string]any{"type": "string", "description": "Target IPv4, IPv6, or hostname"},
					"ip_version":    map[string]any{"type": "string", "enum": []string{"auto", "v4", "v6"}},
					"include_route": map[string]any{"type": "boolean", "description": "Also trace return route on each node (default true)"},
					"protocol":      map[string]any{"type": "string", "enum": []string{"icmp", "tcp"}},
					"port":          map[string]any{"type": "integer"},
				},
				"required": []string{"target"},
			},
		},
		{
			"name":        "vmlg_bgp_route",
			"description": "Query the BGP routing table (RIB) on a node for an IP or CIDR prefix, returning AS-Path, LocalPref, NextHop, and BGP Communities.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target": map[string]any{"type": "string", "description": "Target IP or CIDR prefix (e.g. 1.1.1.1 or 1.1.1.0/24)"},
					"node":   map[string]any{"type": "string", "description": "Optional node ID"},
				},
				"required": []string{"target"},
			},
		},
	}
}
