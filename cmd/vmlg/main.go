package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cloudapp3/vmlg/internal/bgp"
	"github.com/cloudapp3/vmlg/internal/cluster"
	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/enrich"
	"github.com/cloudapp3/vmlg/internal/guard"
	"github.com/cloudapp3/vmlg/internal/mcp"
	"github.com/cloudapp3/vmlg/internal/server"
	"github.com/cloudapp3/vmlg/internal/speedtest"
	"github.com/cloudapp3/vmlg/internal/updater"
	"github.com/cloudapp3/vmlg/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	cmd := strings.ToLower(strings.TrimSpace(os.Getenv("VMLG_MODE")))
	if cmd == "" {
		cmd = "serve"
	}
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	} else if len(args) > 0 && (args[0] == "-version" || args[0] == "--version") {
		cmd = "version"
		args = args[1:]
	} else if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		printUsage()
		return 0
	}

	switch cmd {
	case "serve", "standalone":
		return runHTTPMode(config.ModeStandalone, args)
	case "hub":
		return runHTTPMode(config.ModeHub, args)
	case "agent":
		return runAgentMode(args)
	case "mcp":
		return runMCPMode(args)
	case "healthcheck":
		return runHealthcheck(args)
	case "update":
		return runUpdate(args)
	case "check":
		return runCheck(args)
	case "version":
		return runVersion(args)
	case "help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		printUsage()
		return 2
	}
}

func runHTTPMode(defaultMode string, args []string) int {
	fs := flag.NewFlagSet(defaultMode, flag.ContinueOnError)
	cfgPath := fs.String("config", "", "Path to TOML config file")
	listen := fs.String("listen", "", "HTTP listen address (e.g. 0.0.0.0:8080)")
	title := fs.String("title", "", "Portal header title")
	nodeID := fs.String("node-id", "", "Node identifier")
	nodeName := fs.String("node-name", "", "Node display name")
	location := fs.String("location", "", "Node city/country")
	region := fs.String("region", "", "Node region badge (e.g. APAC, NA, EU)")
	ipv4 := fs.String("ipv4", "", "Node public IPv4")
	ipv6 := fs.String("ipv6", "", "Node public IPv6")
	pubURL := fs.String("public-url", "", "Node public HTTP(S) base URL")
	agentToken := fs.String("agent-token", "", "Hub shared token for Agent connections (hub mode)")
	enableLocal := fs.Bool("enable-local-node", false, "Also expose the Hub host itself as a testable node")
	bgpDriver := fs.String("bgp-driver", "", "Enable BGP module with driver: bird, frr, gobgp, or mock")
	allowPrivate := fs.Bool("allow-private", false, "Allow probing RFC1918/loopback addresses (for lab/testing)")
	jsonLog := fs.Bool("json-log", false, "Emit structured JSON logs")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	logger := newLogger(*jsonLog)
	cfgFile := strings.TrimSpace(*cfgPath)
	if cfgFile == "" {
		cfgFile = strings.TrimSpace(os.Getenv("VMLG_CONFIG"))
	}
	cfg := config.Default(defaultMode)
	if cfgFile != "" {
		loaded, err := config.Load(cfgFile, defaultMode)
		if err != nil {
			logger.Error("failed to load config", "err", err)
			return 1
		}
		cfg = loaded
	} else {
		cfg.ApplyEnv()
	}

	if defaultMode == config.ModeHub && cfgFile == "" {
		cfg.Mode = config.ModeHub
	}
	if *listen != "" {
		cfg.Server.Listen = *listen
	}
	if *title != "" {
		cfg.Server.Title = *title
	}
	if *nodeID != "" {
		cfg.Node.ID = *nodeID
	}
	if *nodeName != "" {
		cfg.Node.Name = *nodeName
	}
	if *location != "" {
		cfg.Node.Location = *location
	}
	if *region != "" {
		cfg.Node.Region = *region
	}
	if *ipv4 != "" {
		cfg.Node.IPv4 = *ipv4
	}
	if *ipv6 != "" {
		cfg.Node.IPv6 = *ipv6
	}
	if *pubURL != "" {
		cfg.Node.PublicURL = *pubURL
	}
	if *agentToken != "" {
		cfg.Hub.AgentToken = *agentToken
	}
	if *enableLocal {
		cfg.Hub.EnableLocalNode = true
	}
	if *bgpDriver != "" {
		cfg.BGP.Enabled = true
		cfg.BGP.Driver = *bgpDriver
	}
	if *allowPrivate {
		cfg.Guard.AllowPrivateTargets = true
	}

	if err := cfg.NormalizeAndValidate(); err != nil {
		logger.Error("invalid configuration", "err", err)
		return 1
	}

	srv, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("failed to initialize server", "err", err)
		return 1
	}
	defer srv.Close()

	httpSrv := &http.Server{
		Addr:              cfg.Server.Listen,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("vmlg server listening",
			"mode", cfg.Mode,
			"listen", cfg.Server.Listen,
			"node_id", cfg.Node.ID,
			"version", version.Version,
		)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("http server terminated", "err", err)
			return 1
		}
	case <-ctx.Done():
		logger.Info("shutting down vmlg server...")
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(sctx)
	}
	return 0
}

func runAgentMode(args []string) int {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "Path to TOML config file")
	hubURL := fs.String("hub-url", "", "Hub WebSocket or HTTP base URL (e.g. wss://lg.example.com)")
	token := fs.String("token", "", "Cluster shared authentication token")
	nodeID := fs.String("node-id", "", "Unique node identifier (e.g. hkg-01)")
	nodeName := fs.String("node-name", "", "Display name (e.g. Hong Kong BGP)")
	location := fs.String("location", "", "Location label")
	region := fs.String("region", "", "Region group (e.g. APAC, NA, EU)")
	provider := fs.String("provider", "", "Provider name")
	asn := fs.String("asn", "", "Node ASN (e.g. AS64512)")
	ipv4 := fs.String("ipv4", "", "Node public IPv4")
	ipv6 := fs.String("ipv6", "", "Node public IPv6")
	pubURL := fs.String("public-url", "", "Direct speedtest public URL (e.g. https://hk-speed.example.com)")
	pubHost := fs.String("public-host", "", "Public hostname/IP for iPerf3 commands")
	speedListen := fs.String("speedtest-listen", "", "Optional direct speedtest HTTP listen addr (e.g. 0.0.0.0:8081)")
	tags := fs.String("tags", "", "Comma-separated node tags (e.g. CN2 GIA,10Gbps)")
	bgpDriver := fs.String("bgp-driver", "", "Enable BGP module with driver: bird, frr, gobgp, or mock")
	allowPrivate := fs.Bool("allow-private", false, "Allow probing RFC1918/loopback addresses")
	jsonLog := fs.Bool("json-log", false, "Emit structured JSON logs")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	logger := newLogger(*jsonLog)
	cfgFile := strings.TrimSpace(*cfgPath)
	if cfgFile == "" {
		cfgFile = strings.TrimSpace(os.Getenv("VMLG_CONFIG"))
	}
	cfg := config.Default(config.ModeAgent)
	if cfgFile != "" {
		loaded, err := config.Load(cfgFile, config.ModeAgent)
		if err != nil {
			logger.Error("failed to load config", "err", err)
			return 1
		}
		cfg = loaded
		cfg.Mode = config.ModeAgent
	} else {
		cfg.ApplyEnv()
	}

	if *hubURL != "" {
		cfg.Agent.HubURL = *hubURL
	}
	if *token != "" {
		cfg.Agent.Token = *token
	}
	if *nodeID != "" {
		cfg.Node.ID = *nodeID
	}
	if *nodeName != "" {
		cfg.Node.Name = *nodeName
	}
	if *location != "" {
		cfg.Node.Location = *location
	}
	if *region != "" {
		cfg.Node.Region = *region
	}
	if *provider != "" {
		cfg.Node.Provider = *provider
	}
	if *asn != "" {
		cfg.Node.ASN = *asn
	}
	if *ipv4 != "" {
		cfg.Node.IPv4 = *ipv4
	}
	if *ipv6 != "" {
		cfg.Node.IPv6 = *ipv6
	}
	if *pubURL != "" {
		cfg.Node.PublicURL = *pubURL
	}
	if *pubHost != "" {
		cfg.Node.PublicHost = *pubHost
	}
	if *speedListen != "" {
		cfg.Agent.SpeedtestListen = *speedListen
	}
	if *tags != "" {
		parts := strings.Split(*tags, ",")
		cfg.Node.Tags = nil
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				cfg.Node.Tags = append(cfg.Node.Tags, t)
			}
		}
	}
	if *allowPrivate {
		cfg.Guard.AllowPrivateTargets = true
	}
	if *bgpDriver != "" {
		cfg.BGP.Enabled = true
		cfg.BGP.Driver = *bgpDriver
	}

	if err := cfg.NormalizeAndValidate(); err != nil {
		logger.Error("invalid agent configuration", "err", err)
		return 1
	}

	g, err := guard.New(cfg.Guard, cfg.Server.TrustedProxies)
	if err != nil {
		logger.Error("failed to initialize guard", "err", err)
		return 1
	}
	enricher := enrich.New(cfg.Probe.EnableDNSASN, cfg.Probe.EnablePTR)
	ipm := speedtest.NewIPerfManager(cfg.IPerf3)
	defer ipm.Close()

	localNode := cluster.NewLocalNode(cfg, g, true, enricher, ipm, bgp.NewDriver(cfg.BGP))
	client := cluster.NewAgentClient(cfg, localNode, g, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := client.Run(ctx); err != nil {
		logger.Error("agent exited with error", "err", err)
		return 1
	}
	return 0
}

func runMCPMode(args []string) int {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "Optional TOML config file for local mode")
	hubHTTP := fs.String("hub-http", "", "Optional remote vmlg Hub/Standalone HTTP(S) URL to query")
	allowPrivate := fs.Bool("allow-private", false, "Allow probing RFC1918/loopback addresses in local mode")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		mcpSrv *mcp.Server
		err    error
	)
	if strings.TrimSpace(*hubHTTP) != "" {
		mcpSrv, err = mcp.NewRemote(*hubHTTP, os.Stdout, os.Stderr)
	} else {
		cfg := config.Default(config.ModeStandalone)
		if *cfgPath != "" {
			cfg, err = config.Load(*cfgPath, config.ModeStandalone)
			if err != nil {
				fmt.Fprintf(os.Stderr, "mcp config error: %v\n", err)
				return 1
			}
		} else {
			cfg.ApplyEnv()
		}
		if *allowPrivate {
			cfg.Guard.AllowPrivateTargets = true
		}
		if err := cfg.NormalizeAndValidate(); err != nil {
			fmt.Fprintf(os.Stderr, "mcp config validation error: %v\n", err)
			return 1
		}
		mcpSrv, err = mcp.NewLocal(cfg, os.Stdout, os.Stderr)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start mcp server: %v\n", err)
		return 1
	}
	defer mcpSrv.Close()

	if err := mcpSrv.Serve(ctx, os.Stdin); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "mcp server exited: %v\n", err)
		return 1
	}
	return 0
}

func runHealthcheck(args []string) int {
	fs := flag.NewFlagSet("healthcheck", flag.ContinueOnError)
	targetURL := fs.String("url", "", "Explicit healthz URL to check")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	u := strings.TrimSpace(*targetURL)
	if u == "" {
		mode := strings.ToLower(strings.TrimSpace(os.Getenv("VMLG_MODE")))
		if mode == config.ModeAgent {
			sp := strings.TrimSpace(os.Getenv("VMLG_SPEEDTEST_LISTEN"))
			if sp == "" {
				st, err := os.Stat(cluster.AgentReadyFile)
				if err != nil || time.Since(st.ModTime()) > 90*time.Second {
					fmt.Fprintln(os.Stderr, "agent tunnel not ready")
					return 1
				}
				return 0
			}
			_, port, err := net.SplitHostPort(sp)
			if err != nil {
				port = "8081"
			}
			u = "http://127.0.0.1:" + port + "/healthz"
		} else {
			listen := strings.TrimSpace(os.Getenv("VMLG_LISTEN"))
			if listen == "" {
				listen = "127.0.0.1:8080"
			}
			_, port, err := net.SplitHostPort(listen)
			if err != nil {
				port = "8080"
			}
			u = "http://127.0.0.1:" + port + "/healthz"
		}
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck failed: %v\n", err)
		return 1
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck returned status %d\n", resp.StatusCode)
		return 1
	}
	return 0
}

func runUpdate(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	targetVer := fs.String("version", "", "Target release tag (e.g. v0.1.0); defaults to latest")
	asJSON := fs.Bool("json", false, "Output update result as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	res, err := updater.Update(ctx, version.Version, updater.Options{
		TargetVersion: *targetVer,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "update failed: %v\n", err)
		return 1
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}
	if !res.Updated {
		fmt.Printf("vmlg is already up to date (%s)\n", res.Version)
		return 0
	}
	fmt.Printf("Updated vmlg from %s to %s (%s)\n", res.PreviousVersion, res.Version, res.SHA256[:12])
	return 0
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "Path to TOML config file")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *cfgPath == "" {
		fmt.Fprintln(os.Stderr, "check requires -config <path>")
		return 2
	}
	cfg, err := config.Load(*cfgPath, config.ModeStandalone)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config invalid: %v\n", err)
		return 1
	}
	fmt.Printf("OK: mode=%s node_id=%s listen=%s\n", cfg.Mode, cfg.Node.ID, cfg.Server.Listen)
	return 0
}

func runVersion(args []string) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "Output version info as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	info := version.Get()
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(info)
		return 0
	}
	fmt.Printf("%s %s (%s/%s, %s, commit %s)\n", info.Name, info.Version, info.OS, info.Arch, info.GoVersion, info.Commit)
	return 0
}

func newLogger(asJSON bool) *slog.Logger {
	if asJSON {
		return slog.New(slog.NewJSONHandler(os.Stderr, nil))
	}
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func printUsage() {
	fmt.Print(`vmlg - Modern Single-Binary Multi-Node Network Looking Glass

Usage:
  vmlg [serve] [flags]    Start standalone single-node Looking Glass (default)
  vmlg hub [flags]        Start central multi-node Looking Glass Hub portal
  vmlg agent [flags]      Start lightweight regional Looking Glass Agent
  vmlg mcp [flags]        Start stdio JSON-RPC MCP server (local or remote Hub)
  vmlg healthcheck        Probe container/local service readiness
  vmlg update [flags]     Self-update binary from GitHub Releases (SHA-256 verified)
  vmlg check -config FILE Validate a TOML configuration file
  vmlg version [-json]    Print build and runtime version info

Run "vmlg <command> -h" for command-specific flags.
`)
}
