package speedtest

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
	"github.com/cloudapp3/vmlg/internal/guard"
)

const chunkBufSize = 1 << 20 // 1 MiB incompressible PRNG block

var (
	prngOnce  sync.Once
	prngBlock []byte

	uploadBufPool = sync.Pool{
		New: func() any {
			b := make([]byte, 64<<10)
			return &b
		},
	}
)

func getPRNGBlock() []byte {
	prngOnce.Do(func() {
		prngBlock = make([]byte, chunkBufSize)
		var seed [32]byte
		copy(seed[:], "vmlg-incompressible-speedtest-v1")
		rng := rand.NewChaCha8(seed)
		_, _ = rng.Read(prngBlock)
	})
	return prngBlock
}

type Handler struct {
	cfg           config.SpeedtestConfig
	guard         *guard.Guard
	nodeID        string
	signingSecret string
	requireSigned bool
	corsOrigins   []string
	allowedFiles  map[string]int64
	onReq         func(kind string)
	onBytes       func(direction string, n int64)
}

func (h *Handler) SetTelemetry(onReq func(kind string), onBytes func(direction string, n int64)) {
	h.onReq = onReq
	h.onBytes = onBytes
}

func NewHandler(cfg config.SpeedtestConfig, g *guard.Guard, nodeID, signingSecret string, requireSigned bool, corsOrigins []string) *Handler {
	files := make(map[string]int64, len(cfg.TestFiles))
	for _, name := range cfg.TestFiles {
		if sz, ok := ParseFileSize(name); ok {
			files[name] = sz
		}
	}
	return &Handler{
		cfg:           cfg,
		guard:         g,
		nodeID:        nodeID,
		signingSecret: signingSecret,
		requireSigned: requireSigned,
		corsOrigins:   corsOrigins,
		allowedFiles:  files,
	}
}

func ParseFileSize(name string) (int64, bool) {
	clean := strings.TrimSpace(name)
	upper := strings.ToUpper(strings.TrimSuffix(clean, ".bin"))
	switch {
	case strings.HasSuffix(upper, "GB"):
		n, err := strconv.ParseInt(strings.TrimSuffix(upper, "GB"), 10, 64)
		if err == nil && n > 0 && n <= 100 {
			return n << 30, true
		}
	case strings.HasSuffix(upper, "MB"):
		n, err := strconv.ParseInt(strings.TrimSuffix(upper, "MB"), 10, 64)
		if err == nil && n > 0 && n <= 102400 {
			return n << 20, true
		}
	}
	return 0, false
}

func (h *Handler) SetCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	allowed := "*"
	if len(h.corsOrigins) > 0 && h.corsOrigins[0] != "*" {
		allowed = ""
		for _, o := range h.corsOrigins {
			if o == origin {
				allowed = origin
				break
			}
		}
	}
	if allowed != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-VMLG-Ticket")
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	return false
}

func (h *Handler) checkAuth(r *http.Request) bool {
	if !h.requireSigned {
		return true
	}
	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		ticket = r.Header.Get("X-VMLG-Ticket")
	}
	return VerifyTicket(h.signingSecret, h.nodeID, ticket, time.Now())
}

func (h *Handler) ServePing(w http.ResponseWriter, r *http.Request) {
	if h.SetCORS(w, r) {
		return
	}
	if !h.cfg.Enabled {
		http.Error(w, "speedtest disabled", http.StatusForbidden)
		return
	}
	if !h.checkAuth(r) {
		http.Error(w, "valid speedtest ticket required", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Server-Timing", "app;dur=0")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ServeDownload(w http.ResponseWriter, r *http.Request) {
	if h.SetCORS(w, r) {
		return
	}
	if !h.cfg.Enabled {
		http.Error(w, "speedtest disabled", http.StatusForbidden)
		return
	}
	if !h.checkAuth(r) {
		http.Error(w, "valid speedtest ticket required", http.StatusUnauthorized)
		return
	}

	maxBytes := h.cfg.MaxDownloadBytes
	if maxBytes <= 0 {
		maxBytes = 256 << 20
	}
	size := int64(25 << 20)
	if raw := r.URL.Query().Get("bytes"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			size = parsed
		}
	}
	size = min(max(size, 64<<10), maxBytes)

	release, err := h.acquireSlot(r, false)
	if err != nil {
		writeGuardErr(w, err)
		return
	}
	defer release()

	if h.onReq != nil {
		h.onReq("speedtest_download")
	}
	h.streamRandomBytes(w, r, size, "speedtest.bin")
}

func (h *Handler) ServeUpload(w http.ResponseWriter, r *http.Request) {
	if h.SetCORS(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.cfg.Enabled {
		http.Error(w, "speedtest disabled", http.StatusForbidden)
		return
	}
	if !h.checkAuth(r) {
		http.Error(w, "valid speedtest ticket required", http.StatusUnauthorized)
		return
	}

	release, err := h.acquireSlot(r, false)
	if err != nil {
		writeGuardErr(w, err)
		return
	}
	defer release()

	maxUpload := h.cfg.MaxUploadBytes
	if maxUpload <= 0 {
		maxUpload = 128 << 20
	}
	reader := http.MaxBytesReader(w, r.Body, maxUpload)
	defer reader.Close()

	bufPtr := uploadBufPool.Get().(*[]byte)
	defer uploadBufPool.Put(bufPtr)
	buf := *bufPtr

	start := time.Now()
	var total int64
	bytesPerSec := h.cfg.MaxRateMbps * 125000.0 // 1 Mbps = 125,000 B/s

	for {
		if r.Context().Err() != nil {
			return
		}
		n, rerr := reader.Read(buf)
		if n > 0 {
			total += int64(n)
			if bytesPerSec > 0 {
				paceStream(r, start, total, bytesPerSec)
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				break
			}
			http.Error(w, "upload read interrupted or exceeded limit", http.StatusRequestEntityTooLarge)
			return
		}
	}

	if h.onReq != nil {
		h.onReq("speedtest_upload")
	}
	if h.onBytes != nil && total > 0 {
		h.onBytes("upload", total)
	}

	durMs := float64(time.Since(start).Microseconds()) / 1000.0
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"received_bytes": total,
		"duration_ms":    durMs,
	})
}

func (h *Handler) ServeTestFile(w http.ResponseWriter, r *http.Request) {
	if h.SetCORS(w, r) {
		return
	}
	if !h.cfg.Enabled {
		http.Error(w, "speedtest disabled", http.StatusForbidden)
		return
	}
	if !h.checkAuth(r) {
		http.Error(w, "valid speedtest ticket required", http.StatusUnauthorized)
		return
	}

	name := r.PathValue("name")
	if name == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			name = parts[len(parts)-1]
		}
	}
	size, ok := h.allowedFiles[name]
	if !ok {
		http.Error(w, "requested test file is not enabled", http.StatusNotFound)
		return
	}

	release, err := h.acquireSlot(r, true)
	if err != nil {
		writeGuardErr(w, err)
		return
	}
	defer release()

	if h.onReq != nil {
		h.onReq("testfile")
	}
	h.streamRandomBytes(w, r, size, name)
}

func (h *Handler) streamRandomBytes(w http.ResponseWriter, r *http.Request, totalBytes int64, filename string) {
	block := getPRNGBlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Encoding", "identity")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, no-transform")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Length", strconv.FormatInt(totalBytes, 10))
	w.Header().Set("X-Accel-Buffering", "no")

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusOK)

	flusher, canFlush := w.(http.Flusher)
	start := time.Now()
	var written int64
	bytesPerSec := h.cfg.MaxRateMbps * 125000.0
	writeSlice := 64 << 10 // 64 KiB write granularity for responsive cancel & pacing
	offset := 0

	for written < totalBytes {
		if r.Context().Err() != nil {
			return
		}
		rem := totalBytes - written
		chunk := int64(writeSlice)
		if rem < chunk {
			chunk = rem
		}
		if offset+int(chunk) > len(block) {
			offset = 0
		}

		n, err := w.Write(block[offset : offset+int(chunk)])
		if n > 0 {
			written += int64(n)
			offset += n
			if h.onBytes != nil {
				h.onBytes("download", int64(n))
			}
		}
		if err != nil {
			return
		}
		if canFlush && written%(256<<10) == 0 {
			flusher.Flush()
		}
		if bytesPerSec > 0 {
			paceStream(r, start, written, bytesPerSec)
		}
	}
	if canFlush {
		flusher.Flush()
	}
}

func paceStream(r *http.Request, start time.Time, transferred int64, bytesPerSec float64) {
	expectedSec := float64(transferred) / bytesPerSec
	actualSec := time.Since(start).Seconds()
	if sleepSec := expectedSec - actualSec; sleepSec > 0.005 {
		d := time.Duration(sleepSec * float64(time.Second))
		t := time.NewTimer(min(d, 200*time.Millisecond))
		defer t.Stop()
		select {
		case <-r.Context().Done():
		case <-t.C:
		}
	}
}

func (h *Handler) acquireSlot(r *http.Request, checkRate bool) (func(), error) {
	if h.guard == nil {
		return func() {}, nil
	}
	return h.guard.AcquireSpeedtest(h.guard.ClientIP(r), checkRate)
}

func writeGuardErr(w http.ResponseWriter, err error) {
	status := http.StatusTooManyRequests
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
