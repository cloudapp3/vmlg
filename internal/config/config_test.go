package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultStandaloneValid(t *testing.T) {
	cfg := Default(ModeStandalone)
	if err := cfg.NormalizeAndValidate(); err != nil {
		t.Fatalf("expected default standalone config valid, got: %v", err)
	}
}

func TestHubRequiresToken(t *testing.T) {
	cfg := Default(ModeHub)
	if err := cfg.NormalizeAndValidate(); err == nil {
		t.Fatal("expected hub config without agent_token to fail")
	}
	cfg.Hub.AgentToken = "secret-token-123"
	if err := cfg.NormalizeAndValidate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Hub.SigningSecret == "" {
		t.Fatal("expected auto-generated signing_secret")
	}
}

func TestUnknownKeyRejected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(p, []byte("mode = \"standalone\"\nunknown_field = 123\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(p, ModeStandalone)
	if err == nil || !strings.Contains(err.Error(), "unknown config keys") {
		t.Fatalf("expected unknown config keys error, got: %v", err)
	}
}
