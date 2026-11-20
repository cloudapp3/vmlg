package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	ModeStandalone = "standalone"
	ModeHub        = "hub"
	ModeAgent      = "agent"
)

type Config struct {
	Mode      string          `toml:"mode"`
	Server    ServerConfig    `toml:"server"`
	Node      NodeConfig      `toml:"node"`
	Hub       HubConfig       `toml:"hub"`
	Agent     AgentConfig     `toml:"agent"`
	Guard     GuardConfig     `toml:"guard"`
	Probe     ProbeConfig     `toml:"probe"`
	Speedtest SpeedtestConfig `toml:"speedtest"`
	IPerf3    IPerf3Config    `toml:"iperf3"`
	BGP       BGPConfig       `toml:"bgp"`
}

type ServerConfig struct {
	Listen         string   `toml:"listen"`
	Title          string   `toml:"title"`
	Subtitle       string   `toml:"subtitle"`
	DefaultLocale  string   `toml:"default_locale"`
	TrustedProxies []string `toml:"trusted_proxies"`
	CORSOrigins    []string `toml:"cors_origins"`
	EnableMetrics  bool     `toml:"enable_metrics"`
	MetricsToken   string   `toml:"metrics_token"`
	RepoURL        string   `toml:"repo_url"`
	CommunityURL   string   `toml:"community_url"`
	CommunityLabel string   `toml:"community_label"`
	BrandURL       string   `toml:"brand_url"`
	BrandLabel     string   `toml:"brand_label"`
	AllowEmbed     bool     `toml:"allow_embed"`
}

type NodeConfig struct {
	ID         string   `toml:"id"`
	Name       string   `toml:"name"`
	Location   string   `toml:"location"`
	Region     string   `toml:"region"`
	Provider   string   `toml:"provider"`
	ASN        string   `toml:"asn"`
	IPv4       string   `toml:"ipv4"`
	IPv6       string   `toml:"ipv6"`
	PublicURL  string   `toml:"public_url"`
	PublicHost string   `toml:"public_host"`
	Tags       []string `toml:"tags"`
}

type HubConfig struct {
	AgentToken      string `toml:"agent_token"`
	SigningSecret   string `toml:"signing_secret"`
	EnableLocalNode bool   `toml:"enable_local_node"`
	TokenTTLSeconds int    `toml:"token_ttl_seconds"`
}

type AgentConfig struct {
	HubURL              string `toml:"hub_url"`
	Token               string `toml:"token"`
	SpeedtestListen     string `toml:"speedtest_listen"`
	SigningSecret       string `toml:"signing_secret"`
	RequireSignedSpeed  bool   `toml:"require_signed_speed"`
	HeartbeatSeconds    int    `toml:"heartbeat_seconds"`
	ReconnectMaxSeconds int    `toml:"reconnect_max_seconds"`
}

type GuardConfig struct {
	AllowPrivateTargets    bool     `toml:"allow_private_targets"`
	AllowedReservedCIDRs   []string `toml:"allowed_reserved_cidrs"`
	BlockedCIDRs           []string `toml:"blocked_cidrs"`
	ClientRPM              int      `toml:"client_rpm"`
	ClientBurst            int      `toml:"client_burst"`
	TargetRPM              int      `toml:"target_rpm"`
	TargetBurst            int      `toml:"target_burst"`
	MaxConcurrentProbes    int      `toml:"max_concurrent_probes"`
	MaxConcurrentSpeedtest int      `toml:"max_concurrent_speedtest"`
}

type ProbeConfig struct {
	PingDefaultCount int  `toml:"ping_default_count"`
	PingMaxCount     int  `toml:"ping_max_count"`
	PingTimeoutMs    int  `toml:"ping_timeout_ms"`
	PingIntervalMs   int  `toml:"ping_interval_ms"`
	MTRDefaultCycles int  `toml:"mtr_default_cycles"`
	MTRMaxCycles     int  `toml:"mtr_max_cycles"`
	MTRMaxHops       int  `toml:"mtr_max_hops"`
	DigTimeoutMs     int  `toml:"dig_timeout_ms"`
	HTTPTimeoutMs    int  `toml:"http_timeout_ms"`
	EnableDNSASN     bool `toml:"enable_dns_asn"`
	EnablePTR        bool `toml:"enable_ptr"`
}

type SpeedtestConfig struct {
	Enabled          bool     `toml:"enabled"`
	MaxRateMbps      float64  `toml:"max_rate_mbps"`
	MaxDownloadBytes int64    `toml:"max_download_bytes"`
	MaxUploadBytes   int64    `toml:"max_upload_bytes"`
	TestFiles        []string `toml:"test_files"`
}

type IPerf3Config struct {
	Enabled      bool   `toml:"enabled"`
	Binary       string `toml:"binary"`
	PortStart    int    `toml:"port_start"`
	PortEnd      int    `toml:"port_end"`
	LeaseSeconds int    `toml:"lease_seconds"`
	MaxActive    int    `toml:"max_active"`
}

type BGPConfig struct {
	Enabled     bool              `toml:"enabled"`
	Driver      string            `toml:"driver"` // "bird" | "frr" | "gobgp" | "mock"
	RouterASN   string            `toml:"router_asn"`
	BirdSocket  string            `toml:"bird_socket"`
	VtyshBinary string            `toml:"vtysh_binary"`
	GoBGPBinary string            `toml:"gobgp_binary"`
	Communities map[string]string `toml:"communities"`
}

func Default(mode string) Config {
	if mode == "" {
		mode = ModeStandalone
	}
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "local-node"
	}
	nodeID := sanitizeID(hostname)

	return Config{
		Mode: mode,
		Server: ServerConfig{
			Listen:         "0.0.0.0:8080",
			Title:          "vmlg Looking Glass",
			Subtitle:       "Multi-Node Network Diagnostics & Speedtest Portal",
			DefaultLocale:  "en-US",
			CORSOrigins:    []string{"*"},
			EnableMetrics:  true,
			RepoURL:        "https://github.com/cloudapp3/vmlg",
			CommunityURL:   "https://t.me/VMPulse",
			CommunityLabel: "Telegram",
			AllowEmbed:     true,
		},
		Node: NodeConfig{
			ID:       nodeID,
			Name:     hostname,
			Location: "Local",
			Region:   "Default",
			Tags:     []string{},
		},
		Hub: HubConfig{
			EnableLocalNode: false,
			TokenTTLSeconds: 300,
		},
		Agent: AgentConfig{
			HeartbeatSeconds:    15,
			ReconnectMaxSeconds: 30,
		},
		Guard: GuardConfig{
			AllowPrivateTargets:    false,
			ClientRPM:              30,
			ClientBurst:            10,
			TargetRPM:              20,
			TargetBurst:            8,
			MaxConcurrentProbes:    8,
			MaxConcurrentSpeedtest: 4,
		},
		Probe: ProbeConfig{
			PingDefaultCount: 4,
			PingMaxCount:     10,
			PingTimeoutMs:    1500,
			PingIntervalMs:   350,
			MTRDefaultCycles: 5,
			MTRMaxCycles:     10,
			MTRMaxHops:       30,
			DigTimeoutMs:     5000,
			HTTPTimeoutMs:    10000,
			EnableDNSASN:     true,
			EnablePTR:        true,
		},
		Speedtest: SpeedtestConfig{
			Enabled:          true,
			MaxRateMbps:      0, // 0 = unlimited
			MaxDownloadBytes: 256 << 20,
			MaxUploadBytes:   128 << 20,
			TestFiles:        []string{"10MB.bin", "100MB.bin", "1GB.bin"},
		},
		IPerf3: IPerf3Config{
			Enabled:      true,
			Binary:       "iperf3",
			PortStart:    5201,
			PortEnd:      5210,
			LeaseSeconds: 120,
			MaxActive:    4,
		},
		BGP: BGPConfig{
			Enabled:     false,
			Driver:      "bird",
			BirdSocket:  "/run/bird/bird.ctl",
			VtyshBinary: "vtysh",
			GoBGPBinary: "gobgp",
			Communities: map[string]string{},
		},
	}
}

func Load(path string, defaultMode string) (Config, error) {
	cfg := Default(defaultMode)
	if path != "" {
		md, err := toml.DecodeFile(path, &cfg)
		if err != nil {
			return Config{}, fmt.Errorf("parse config %s: %w", path, err)
		}
		if undecoded := md.Undecoded(); len(undecoded) > 0 {
			keys := make([]string, 0, len(undecoded))
			for _, k := range undecoded {
				keys = append(keys, k.String())
			}
			return Config{}, fmt.Errorf("unknown config keys in %s: %s", path, strings.Join(keys, ", "))
		}
	}
	cfg.ApplyEnv()
	if err := cfg.NormalizeAndValidate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) ApplyEnv() {
	if v := readEnvOrFile("VMLG_MODE"); v != "" {
		c.Mode = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_LISTEN")); v != "" {
		c.Server.Listen = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_TITLE")); v != "" {
		c.Server.Title = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_SUBTITLE")); v != "" {
		c.Server.Subtitle = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_LOCALE")); v != "" {
		c.Server.DefaultLocale = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_REPO_URL")); v != "" {
		c.Server.RepoURL = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_COMMUNITY_URL")); v != "" {
		c.Server.CommunityURL = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_COMMUNITY_LABEL")); v != "" {
		c.Server.CommunityLabel = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_BRAND_URL")); v != "" {
		c.Server.BrandURL = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_BRAND_LABEL")); v != "" {
		c.Server.BrandLabel = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_ALLOW_EMBED")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Server.AllowEmbed = b
		}
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_TRUSTED_PROXIES")); v != "" {
		c.Server.TrustedProxies = splitCSV(v)
	}

	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_ID")); v != "" {
		c.Node.ID = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_NAME")); v != "" {
		c.Node.Name = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_LOCATION")); v != "" {
		c.Node.Location = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_REGION")); v != "" {
		c.Node.Region = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_PROVIDER")); v != "" {
		c.Node.Provider = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_ASN")); v != "" {
		c.Node.ASN = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_IPV4")); v != "" {
		c.Node.IPv4 = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_IPV6")); v != "" {
		c.Node.IPv6 = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_PUBLIC_URL")); v != "" {
		c.Node.PublicURL = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_PUBLIC_HOST")); v != "" {
		c.Node.PublicHost = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_NODE_TAGS")); v != "" {
		c.Node.Tags = splitCSV(v)
	}

	if v := readEnvOrFile("VMLG_AGENT_TOKEN"); v != "" {
		if c.Hub.AgentToken == "" {
			c.Hub.AgentToken = v
		}
		if c.Agent.Token == "" {
			c.Agent.Token = v
		}
	}
	if v := readEnvOrFile("VMLG_SIGNING_SECRET"); v != "" {
		if c.Hub.SigningSecret == "" {
			c.Hub.SigningSecret = v
		}
		if c.Agent.SigningSecret == "" {
			c.Agent.SigningSecret = v
		}
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_HUB_ENABLE_LOCAL_NODE")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Hub.EnableLocalNode = b
		}
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_HUB_URL")); v != "" && c.Agent.HubURL == "" {
		c.Agent.HubURL = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_SPEEDTEST_LISTEN")); v != "" {
		c.Agent.SpeedtestListen = v
	}
	if v := readEnvOrFile("VMLG_METRICS_TOKEN"); v != "" && c.Server.MetricsToken == "" {
		c.Server.MetricsToken = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_ALLOW_PRIVATE")); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.Guard.AllowPrivateTargets = b
		}
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_BGP_DRIVER")); v != "" {
		c.BGP.Enabled = true
		c.BGP.Driver = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_BIRD_SOCKET")); v != "" {
		c.BGP.BirdSocket = v
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_IPERF3_PORT_START")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.IPerf3.PortStart = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("VMLG_IPERF3_PORT_END")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.IPerf3.PortEnd = n
		}
	}
}

func readEnvOrFile(key string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	if filePath := strings.TrimSpace(os.Getenv(key + "_FILE")); filePath != "" {
		if b, err := os.ReadFile(filePath); err == nil {
			return strings.TrimSpace(string(b))
		}
	}
	return ""
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func (c *Config) NormalizeAndValidate() error {
	c.Mode = strings.ToLower(strings.TrimSpace(c.Mode))
	switch c.Mode {
	case ModeStandalone, ModeHub, ModeAgent:
	default:
		return fmt.Errorf("invalid mode %q: expected standalone, hub, or agent", c.Mode)
	}

	c.Node.ID = sanitizeID(c.Node.ID)
	if c.Node.ID == "" {
		return errors.New("node.id must not be empty")
	}
	if strings.TrimSpace(c.Node.Name) == "" {
		c.Node.Name = c.Node.ID
	}

	if c.Mode == ModeStandalone || c.Mode == ModeHub {
		if _, _, err := net.SplitHostPort(c.Server.Listen); err != nil {
			return fmt.Errorf("invalid server.listen %q: %w", c.Server.Listen, err)
		}
	}

	for _, raw := range c.Server.TrustedProxies {
		if _, err := parseCIDROrAddr(raw); err != nil {
			return fmt.Errorf("invalid server.trusted_proxies entry %q: %w", raw, err)
		}
	}
	for _, raw := range c.Guard.AllowedReservedCIDRs {
		if _, err := parseCIDROrAddr(raw); err != nil {
			return fmt.Errorf("invalid guard.allowed_reserved_cidrs entry %q: %w", raw, err)
		}
	}
	for _, raw := range c.Guard.BlockedCIDRs {
		if _, err := parseCIDROrAddr(raw); err != nil {
			return fmt.Errorf("invalid guard.blocked_cidrs entry %q: %w", raw, err)
		}
	}

	if c.Node.IPv4 != "" {
		ip, err := netip.ParseAddr(c.Node.IPv4)
		if err != nil || !ip.Is4() {
			return fmt.Errorf("invalid node.ipv4 %q", c.Node.IPv4)
		}
	}
	if c.Node.IPv6 != "" {
		ip, err := netip.ParseAddr(c.Node.IPv6)
		if err != nil || !ip.Is6() {
			return fmt.Errorf("invalid node.ipv6 %q", c.Node.IPv6)
		}
	}
	if c.Node.PublicURL != "" {
		c.Node.PublicURL = strings.TrimRight(strings.TrimSpace(c.Node.PublicURL), "/")
		u, err := url.Parse(c.Node.PublicURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("invalid node.public_url %q: must be http(s)://host[:port]", c.Node.PublicURL)
		}
	}

	if c.Mode == ModeHub {
		if strings.TrimSpace(c.Hub.AgentToken) == "" {
			return errors.New("hub.agent_token (or VMLG_AGENT_TOKEN) is required in hub mode")
		}
		if c.Hub.SigningSecret == "" {
			var b [32]byte
			_, _ = rand.Read(b[:])
			c.Hub.SigningSecret = hex.EncodeToString(b[:])
		}
		if c.Hub.TokenTTLSeconds <= 0 {
			c.Hub.TokenTTLSeconds = 300
		}
	}

	if c.Mode == ModeAgent {
		if strings.TrimSpace(c.Agent.HubURL) == "" {
			return errors.New("agent.hub_url (or VMLG_HUB_URL) is required in agent mode")
		}
		u, err := url.Parse(c.Agent.HubURL)
		if err != nil || (u.Scheme != "ws" && u.Scheme != "wss" && u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("invalid agent.hub_url %q", c.Agent.HubURL)
		}
		if strings.TrimSpace(c.Agent.Token) == "" {
			return errors.New("agent.token (or VMLG_AGENT_TOKEN) is required in agent mode")
		}
		if c.Agent.SpeedtestListen != "" {
			if _, _, err := net.SplitHostPort(c.Agent.SpeedtestListen); err != nil {
				return fmt.Errorf("invalid agent.speedtest_listen %q: %w", c.Agent.SpeedtestListen, err)
			}
		}
		if c.Agent.RequireSignedSpeed && strings.TrimSpace(c.Agent.SigningSecret) == "" {
			return errors.New("agent.signing_secret is required when agent.require_signed_speed is true")
		}
		if c.Agent.HeartbeatSeconds <= 0 {
			c.Agent.HeartbeatSeconds = 15
		}
		if c.Agent.ReconnectMaxSeconds <= 0 {
			c.Agent.ReconnectMaxSeconds = 30
		}
	}

	if c.Guard.ClientRPM <= 0 {
		c.Guard.ClientRPM = 30
	}
	if c.Guard.ClientBurst <= 0 {
		c.Guard.ClientBurst = 10
	}
	if c.Guard.TargetRPM <= 0 {
		c.Guard.TargetRPM = 20
	}
	if c.Guard.TargetBurst <= 0 {
		c.Guard.TargetBurst = 8
	}
	if c.Guard.MaxConcurrentProbes <= 0 {
		c.Guard.MaxConcurrentProbes = 8
	}
	if c.Guard.MaxConcurrentSpeedtest <= 0 {
		c.Guard.MaxConcurrentSpeedtest = 4
	}

	if c.Probe.PingMaxCount <= 0 || c.Probe.PingMaxCount > 50 {
		return fmt.Errorf("probe.ping_max_count must be between 1 and 50 (got %d)", c.Probe.PingMaxCount)
	}
	if c.Probe.PingDefaultCount <= 0 || c.Probe.PingDefaultCount > c.Probe.PingMaxCount {
		c.Probe.PingDefaultCount = min(4, c.Probe.PingMaxCount)
	}
	if c.Probe.MTRMaxCycles <= 0 || c.Probe.MTRMaxCycles > 30 {
		return fmt.Errorf("probe.mtr_max_cycles must be between 1 and 30 (got %d)", c.Probe.MTRMaxCycles)
	}
	if c.Probe.MTRDefaultCycles <= 0 || c.Probe.MTRDefaultCycles > c.Probe.MTRMaxCycles {
		c.Probe.MTRDefaultCycles = min(5, c.Probe.MTRMaxCycles)
	}
	if c.Probe.MTRMaxHops <= 0 || c.Probe.MTRMaxHops > 64 {
		return fmt.Errorf("probe.mtr_max_hops must be between 1 and 64 (got %d)", c.Probe.MTRMaxHops)
	}
	if c.Probe.DigTimeoutMs <= 0 || c.Probe.DigTimeoutMs > 15000 {
		c.Probe.DigTimeoutMs = 5000
	}
	if c.Probe.HTTPTimeoutMs <= 0 || c.Probe.HTTPTimeoutMs > 30000 {
		c.Probe.HTTPTimeoutMs = 10000
	}

	if c.IPerf3.Enabled {
		if c.IPerf3.PortStart < 1024 || c.IPerf3.PortEnd > 65535 || c.IPerf3.PortStart > c.IPerf3.PortEnd {
			return fmt.Errorf("invalid iperf3 port range %d-%d", c.IPerf3.PortStart, c.IPerf3.PortEnd)
		}
		if c.IPerf3.LeaseSeconds < 10 || c.IPerf3.LeaseSeconds > 600 {
			return fmt.Errorf("iperf3.lease_seconds must be between 10 and 600 (got %d)", c.IPerf3.LeaseSeconds)
		}
		if c.IPerf3.MaxActive <= 0 {
			c.IPerf3.MaxActive = 4
		}
	}

	if c.BGP.Enabled {
		c.BGP.Driver = strings.ToLower(strings.TrimSpace(c.BGP.Driver))
		switch c.BGP.Driver {
		case "bird", "frr", "gobgp", "mock":
		default:
			return fmt.Errorf("invalid bgp.driver %q: expected bird, frr, gobgp, or mock", c.BGP.Driver)
		}
		if c.BGP.BirdSocket == "" {
			c.BGP.BirdSocket = "/run/bird/bird.ctl"
		}
		if c.BGP.VtyshBinary == "" {
			c.BGP.VtyshBinary = "vtysh"
		}
		if c.BGP.GoBGPBinary == "" {
			c.BGP.GoBGPBinary = "gobgp"
		}
		if c.BGP.RouterASN == "" && c.Node.ASN != "" {
			c.BGP.RouterASN = c.Node.ASN
		}
	}

	return nil
}

func (c *Config) PingTimeout() time.Duration {
	if c.Probe.PingTimeoutMs <= 0 {
		return 1500 * time.Millisecond
	}
	return time.Duration(c.Probe.PingTimeoutMs) * time.Millisecond
}

func (c *Config) PingInterval() time.Duration {
	if c.Probe.PingIntervalMs <= 0 {
		return 350 * time.Millisecond
	}
	return time.Duration(c.Probe.PingIntervalMs) * time.Millisecond
}

func parseCIDROrAddr(s string) (netip.Prefix, error) {
	s = strings.TrimSpace(s)
	if strings.Contains(s, "/") {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			return netip.Prefix{}, err
		}
		return p.Masked(), nil
	}
	ip, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(ip, ip.BitLen()), nil
}

func sanitizeID(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
			lastDash = false
		} else if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
