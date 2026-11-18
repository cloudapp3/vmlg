package speedtest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
)

var (
	ErrIPerfDisabled    = errors.New("iperf3 service is disabled on this node")
	ErrIPerfUnavailable = errors.New("iperf3 binary is not installed on this node")
	ErrIPerfBusy        = errors.New("all iperf3 ephemeral ports are currently in use, please retry shortly")
)

type IPerfLease struct {
	NodeID         string    `json:"node_id,omitempty"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	LeaseSeconds   int       `json:"lease_seconds"`
	ExpiresAt      time.Time `json:"expires_at"`
	CommandTCP     string    `json:"command_tcp"`
	CommandReverse string    `json:"command_reverse"`
	CommandUDP     string    `json:"command_udp"`
}

type activeLease struct {
	lease     IPerfLease
	clientKey string
	cancel    context.CancelFunc
}

type processStarter func(ctx context.Context, bin string, port int) (waitFn func() error, err error)

type IPerfManager struct {
	cfg     config.IPerf3Config
	binPath string
	starter processStarter

	mu     sync.Mutex
	active map[int]*activeLease // keyed by port
}

func NewIPerfManager(cfg config.IPerf3Config) *IPerfManager {
	m := &IPerfManager{
		cfg:    cfg,
		active: make(map[int]*activeLease),
	}
	if cfg.Enabled {
		bin := cfg.Binary
		if bin == "" {
			bin = "iperf3"
		}
		if p, err := exec.LookPath(bin); err == nil {
			m.binPath = p
		}
	}
	m.starter = defaultProcessStarter
	return m
}

func defaultProcessStarter(ctx context.Context, bin string, port int) (func() error, error) {
	cmd := exec.CommandContext(ctx, bin, "-s", "-p", strconv.Itoa(port), "-1")
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd.Wait, nil
}

func (m *IPerfManager) Available() (bool, string) {
	if !m.cfg.Enabled {
		return false, ErrIPerfDisabled.Error()
	}
	if m.binPath == "" {
		return false, ErrIPerfUnavailable.Error()
	}
	return true, ""
}

func (m *IPerfManager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reapExpiredLocked(time.Now())
	return len(m.active)
}

func (m *IPerfManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for port, entry := range m.active {
		entry.cancel()
		delete(m.active, port)
	}
}

func (m *IPerfManager) Acquire(clientKey string, publicHost string, nodeID string) (IPerfLease, error) {
	if !m.cfg.Enabled {
		return IPerfLease{}, ErrIPerfDisabled
	}
	if m.binPath == "" {
		return IPerfLease{}, ErrIPerfUnavailable
	}
	if publicHost == "" {
		publicHost = "127.0.0.1"
	}

	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	m.reapExpiredLocked(now)

	// Return existing active lease if same client already holds one
	if clientKey != "" {
		for _, entry := range m.active {
			if entry.clientKey == clientKey && now.Before(entry.lease.ExpiresAt) {
				cp := entry.lease
				cp.LeaseSeconds = max(1, int(time.Until(cp.ExpiresAt).Seconds()))
				return cp, nil
			}
		}
	}

	maxActive := max(m.cfg.MaxActive, 1)
	if len(m.active) >= maxActive {
		return IPerfLease{}, ErrIPerfBusy
	}

	chosenPort := 0
	for p := m.cfg.PortStart; p <= m.cfg.PortEnd; p++ {
		if _, busy := m.active[p]; busy {
			continue
		}
		if isPortAvailable(p) {
			chosenPort = p
			break
		}
	}
	if chosenPort == 0 {
		return IPerfLease{}, ErrIPerfBusy
	}

	ttl := time.Duration(m.cfg.LeaseSeconds) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	waitFn, err := m.starter(ctx, m.binPath, chosenPort)
	if err != nil {
		cancel()
		return IPerfLease{}, fmt.Errorf("failed to start iperf3 server: %w", err)
	}

	hostFmt := publicHost
	if ip := net.ParseIP(publicHost); ip != nil && ip.To4() == nil {
		hostFmt = "[" + publicHost + "]"
	}

	lease := IPerfLease{
		NodeID:         nodeID,
		Host:           publicHost,
		Port:           chosenPort,
		LeaseSeconds:   m.cfg.LeaseSeconds,
		ExpiresAt:      now.Add(ttl),
		CommandTCP:     fmt.Sprintf("iperf3 -c %s -p %d -P 4", hostFmt, chosenPort),
		CommandReverse: fmt.Sprintf("iperf3 -c %s -p %d -P 4 -R", hostFmt, chosenPort),
		CommandUDP:     fmt.Sprintf("iperf3 -c %s -p %d -u -b 100M", hostFmt, chosenPort),
	}

	entry := &activeLease{
		lease:     lease,
		clientKey: clientKey,
		cancel:    cancel,
	}
	m.active[chosenPort] = entry

	go func(port int, target *activeLease) {
		_ = waitFn()
		target.cancel()
		m.mu.Lock()
		if cur, ok := m.active[port]; ok && cur == target {
			delete(m.active, port)
		}
		m.mu.Unlock()
	}(chosenPort, entry)

	return lease, nil
}

func (m *IPerfManager) reapExpiredLocked(now time.Time) {
	for port, entry := range m.active {
		if !now.Before(entry.lease.ExpiresAt) {
			entry.cancel()
			delete(m.active, port)
		}
	}
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}
