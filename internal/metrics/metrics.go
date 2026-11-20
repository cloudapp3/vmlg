package metrics

import (
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cloudapp3/vmlg/internal/version"
)

type SnapshotGauge struct {
	OnlineNodes       int
	OfflineNodes      int
	ActiveProbes      int
	ActiveSpeedtests  int
	ActiveIPerfLeases int
	NodeUps           []NodeUp
}

// NodeUp is one per-node liveness gauge sample (1 = online, 0 = offline).
type NodeUp struct {
	NodeID string
	Up     int
}

type Collector struct {
	mode        string
	bearerToken string
	gaugeFn     func() SnapshotGauge

	mu            sync.RWMutex
	requests      map[string]*atomic.Uint64
	guardRejects  map[string]*atomic.Uint64
	downloadBytes atomic.Uint64
	uploadBytes   atomic.Uint64
}

func New(mode, bearerToken string, gaugeFn func() SnapshotGauge) *Collector {
	return &Collector{
		mode:         mode,
		bearerToken:  strings.TrimSpace(bearerToken),
		gaugeFn:      gaugeFn,
		requests:     make(map[string]*atomic.Uint64),
		guardRejects: make(map[string]*atomic.Uint64),
	}
}

func (c *Collector) IncRequest(kind string) {
	if c == nil {
		return
	}
	counterFor(&c.mu, c.requests, kind).Add(1)
}

func (c *Collector) IncGuardReject(reason string) {
	if c == nil {
		return
	}
	counterFor(&c.mu, c.guardRejects, reason).Add(1)
}

func (c *Collector) AddDownloadBytes(n int64) {
	if c != nil && n > 0 {
		c.downloadBytes.Add(uint64(n))
	}
}

func (c *Collector) AddUploadBytes(n int64) {
	if c != nil && n > 0 {
		c.uploadBytes.Add(uint64(n))
	}
}

func counterFor(mu *sync.RWMutex, m map[string]*atomic.Uint64, key string) *atomic.Uint64 {
	mu.RLock()
	v, ok := m[key]
	mu.RUnlock()
	if ok {
		return v
	}
	mu.Lock()
	defer mu.Unlock()
	if v, ok = m[key]; ok {
		return v
	}
	v = &atomic.Uint64{}
	m[key] = v
	return v
}

func (c *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if c.bearerToken != "" {
		got := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer"))
		if subtle.ConstantTimeCompare([]byte(got), []byte(c.bearerToken)) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="vmlg-metrics"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	c.WritePrometheus(w)
}

func (c *Collector) WritePrometheus(w io.Writer) {
	info := version.Get()
	_, _ = fmt.Fprintf(w, "# HELP vmlg_build_info Build and runtime metadata for vmlg.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_build_info gauge\n")
	_, _ = fmt.Fprintf(w, "vmlg_build_info{version=%q,go_version=%q,mode=%q} 1\n", info.Version, info.GoVersion, c.mode)

	var g SnapshotGauge
	if c.gaugeFn != nil {
		g = c.gaugeFn()
	}

	_, _ = fmt.Fprintf(w, "# HELP vmlg_nodes_total Number of registered nodes by status.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_nodes_total gauge\n")
	_, _ = fmt.Fprintf(w, "vmlg_nodes_total{status=\"online\"} %d\n", g.OnlineNodes)
	_, _ = fmt.Fprintf(w, "vmlg_nodes_total{status=\"offline\"} %d\n", g.OfflineNodes)

	if len(g.NodeUps) > 0 {
		_, _ = fmt.Fprintf(w, "# HELP vmlg_node_up Node liveness by node_id (1 = online, 0 = offline).\n")
		_, _ = fmt.Fprintf(w, "# TYPE vmlg_node_up gauge\n")
		ups := make([]NodeUp, len(g.NodeUps))
		copy(ups, g.NodeUps)
		sort.Slice(ups, func(i, j int) bool { return ups[i].NodeID < ups[j].NodeID })
		for _, n := range ups {
			_, _ = fmt.Fprintf(w, "vmlg_node_up{node_id=%q} %d\n", n.NodeID, n.Up)
		}
	}

	_, _ = fmt.Fprintf(w, "# HELP vmlg_active_probes Currently executing diagnostic probes.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_active_probes gauge\n")
	_, _ = fmt.Fprintf(w, "vmlg_active_probes %d\n", g.ActiveProbes)

	_, _ = fmt.Fprintf(w, "# HELP vmlg_active_speedtests Currently executing speedtest streams.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_active_speedtests gauge\n")
	_, _ = fmt.Fprintf(w, "vmlg_active_speedtests %d\n", g.ActiveSpeedtests)

	_, _ = fmt.Fprintf(w, "# HELP vmlg_active_iperf_leases Currently active ephemeral iPerf3 server leases.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_active_iperf_leases gauge\n")
	_, _ = fmt.Fprintf(w, "vmlg_active_iperf_leases %d\n", g.ActiveIPerfLeases)

	_, _ = fmt.Fprintf(w, "# HELP vmlg_speedtest_bytes_total Total speedtest bytes transferred.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_speedtest_bytes_total counter\n")
	_, _ = fmt.Fprintf(w, "vmlg_speedtest_bytes_total{direction=\"download\"} %d\n", c.downloadBytes.Load())
	_, _ = fmt.Fprintf(w, "vmlg_speedtest_bytes_total{direction=\"upload\"} %d\n", c.uploadBytes.Load())

	_, _ = fmt.Fprintf(w, "# HELP vmlg_requests_total Total Looking Glass API requests by kind.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_requests_total counter\n")
	writeMapCounters(w, "vmlg_requests_total", "kind", &c.mu, c.requests)

	_, _ = fmt.Fprintf(w, "# HELP vmlg_guard_rejections_total Total requests rejected by anti-abuse guard.\n")
	_, _ = fmt.Fprintf(w, "# TYPE vmlg_guard_rejections_total counter\n")
	writeMapCounters(w, "vmlg_guard_rejections_total", "reason", &c.mu, c.guardRejects)
}

func writeMapCounters(w io.Writer, metricName, labelKey string, mu *sync.RWMutex, m map[string]*atomic.Uint64) {
	mu.RLock()
	keys := make([]string, 0, len(m))
	vals := make(map[string]uint64, len(m))
	for k, v := range m {
		keys = append(keys, k)
		vals[k] = v.Load()
	}
	mu.RUnlock()

	sort.Strings(keys)
	for _, k := range keys {
		_, _ = fmt.Fprintf(w, "%s{%s=%q} %d\n", metricName, labelKey, k, vals[k])
	}
}
