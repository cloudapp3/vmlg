package mcp

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/cloudapp3/vmlg/internal/config"
)

func TestMCPStdioServerInitializeAndToolsCall(t *testing.T) {
	cfg := config.Default(config.ModeStandalone)
	cfg.Node.ID = "mcp-test-node"
	cfg.BGP.Enabled = true
	cfg.BGP.Driver = "mock"
	cfg.BGP.RouterASN = "AS64512"

	var stdout bytes.Buffer
	srv, err := NewLocal(cfg, &stdout, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vmlg_list_nodes","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vmlg_bgp_route","arguments":{"target":"1.1.1.1"}}}`,
	}, "\n") + "\n"

	if err := srv.Serve(context.Background(), strings.NewReader(input)); err != nil {
		t.Fatalf("MCP Serve failed: %v", err)
	}

	out := stdout.String()
	for _, want := range []string{
		`"protocolVersion":"2025-06-18"`,
		`"vmlg_compare"`,
		`"mcp-test-node"`,
		`"AS4809"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing expected MCP output %q in:\n%s", want, out)
		}
	}
}
