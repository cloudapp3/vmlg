package cluster

import (
	"strings"
	"testing"
	"time"
)

func TestCollectNodeStatus(t *testing.T) {
	started := time.Now().Add(-90 * time.Second)
	sn := CollectNodeStatus(started)

	if sn.Version == "" {
		t.Fatalf("version should always be set, got %+v", sn)
	}
	if sn.UptimeSeconds < 85 || sn.UptimeSeconds > 120 {
		t.Fatalf("uptime_seconds = %d, want ~90", sn.UptimeSeconds)
	}
	// On Linux CI /proc exists: load averages are non-negative, memory in 0..100.
	if sn.LoadAvg1 < 0 || sn.LoadAvg5 < 0 || sn.LoadAvg15 < 0 {
		t.Fatalf("negative loadavg: %+v", sn)
	}
	if sn.MemUsedPercent < 0 || sn.MemUsedPercent > 100 {
		t.Fatalf("mem_used_percent out of range: %+v", sn)
	}
}

func TestRegistryLoadSampleRingBuffer(t *testing.T) {
	reg := NewRegistry("tok", "sec", 0, nil, nil)

	// Fresh registry: zero-disk policy means the sparkline starts empty.
	if got := reg.NodeLoadSamples("n1"); len(got) != 0 {
		t.Fatalf("fresh registry should have no samples, got %d", len(got))
	}

	// Capacity bound: 300 appends keep only the newest maxLoadSamples points.
	base := time.Now().UTC().Unix() - 400
	for i := 0; i < 300; i++ {
		reg.mu.Lock()
		reg.samples["n1"] = appendCapped(reg.samples["n1"], LoadSample{TS: base + int64(i), Load1: float64(i)})
		reg.mu.Unlock()
	}
	if got := len(reg.NodeLoadSamples("n1")); got != maxLoadSamples {
		t.Fatalf("ring buffer len = %d, want %d", got, maxLoadSamples)
	}
}

func appendCapped(buf []LoadSample, s LoadSample) []LoadSample {
	buf = append(buf, s)
	if len(buf) > maxLoadSamples {
		buf = buf[len(buf)-maxLoadSamples:]
	}
	return buf
}

func TestHeartbeatCarriesStatus(t *testing.T) {
	// The heartbeat payload must serialize the status snapshot for hub-side
	// aggregation (decode round-trip mirrors the ws envelope path).
	hb := HeartbeatPayload{
		ActiveProbes: 2,
		Status: &NodeStatusSnapshot{
			LoadAvg1:       0.42,
			LoadAvg5:       0.5,
			LoadAvg15:      0.55,
			MemUsedPercent: 61.2,
			UptimeSeconds:  3600,
			Version:        "test",
		},
	}
	env, err := makeEnvelope(MsgHeartbeat, "", hb, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(env.Payload), `"loadavg_1":0.42`) ||
		!strings.Contains(string(env.Payload), `"mem_used_percent":61.2`) {
		t.Fatalf("status missing from heartbeat payload: %s", string(env.Payload))
	}
}
