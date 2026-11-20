package cluster

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudapp3/vmlg/internal/version"
)

// NodeStatusSnapshot is the coarse per-node health payload carried in every
// heartbeat: load averages, memory usage percent, and agent uptime. Deliberately
// excludes disk, per-interface traffic, and host identity (privacy by design).
type NodeStatusSnapshot struct {
	LoadAvg1       float64 `json:"loadavg_1"`
	LoadAvg5       float64 `json:"loadavg_5"`
	LoadAvg15      float64 `json:"loadavg_15"`
	MemUsedPercent float64 `json:"mem_used_percent"`
	UptimeSeconds  int64   `json:"uptime_seconds"`
	Version        string  `json:"version"`
}

// CollectNodeStatus reads a snapshot from /proc (Linux). Reads are cheap
// (microseconds); missing files yield zero values instead of errors so the
// heartbeat keeps flowing on non-Linux dev machines.
func CollectNodeStatus(startedAt time.Time) NodeStatusSnapshot {
	uptime := int64(time.Since(startedAt).Seconds())
	if uptime == 0 {
		uptime = 1 // a live process has at least one second of uptime (keeps omitempty honest)
	}
	sn := NodeStatusSnapshot{
		UptimeSeconds: uptime,
		Version:       version.Version,
	}
	if raw, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(raw))
		if len(fields) >= 3 {
			sn.LoadAvg1, _ = strconv.ParseFloat(fields[0], 64)
			sn.LoadAvg5, _ = strconv.ParseFloat(fields[1], 64)
			sn.LoadAvg15, _ = strconv.ParseFloat(fields[2], 64)
		}
	}
	if raw, err := os.ReadFile("/proc/meminfo"); err == nil {
		var total, available float64
		for _, line := range strings.Split(string(raw), "\n") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			val, _ := strconv.ParseFloat(parts[1], 64)
			switch parts[0] {
			case "MemTotal:":
				total = val
			case "MemAvailable:":
				available = val
			}
		}
		if total > 0 {
			used := total - available
			if used < 0 {
				used = 0
			}
			sn.MemUsedPercent = round1(used / total * 100)
		}
	}
	return sn
}

func round1(v float64) float64 {
	return float64(int64(v*10+0.5)) / 10
}
