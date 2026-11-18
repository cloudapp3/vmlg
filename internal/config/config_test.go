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

func TestDockerSecretFileEnvOverride(t *testing.T) {
	dir := t.TempDir()
	secretFile := filepath.Join(dir, "agent_token.txt")
	if err := os.WriteFile(secretFile, []byte("docker-secret-token-999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VMLG_AGENT_TOKEN_FILE", secretFile)
	t.Setenv("VMLG_NODE_TAGS", "CN2 GIA, 10Gbps")

	cfg := Default(ModeHub)
	cfg.ApplyEnv()
	if err := cfg.NormalizeAndValidate(); err != nil {
		t.Fatalf("expected valid hub config from *_FILE secret, got: %v", err)
	}
	if cfg.Hub.AgentToken != "docker-secret-token-999" {
		t.Fatalf("unexpected token: %q", cfg.Hub.AgentToken)
	}
	if len(cfg.Node.Tags) != 2 || cfg.Node.Tags[0] != "CN2 GIA" {
		t.Fatalf("unexpected tags: %+v", cfg.Node.Tags)
	}
}
