package speedtest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cloudapp3/vmlg/internal/config"
)

func TestSpeedtestTicketSignVerify(t *testing.T) {
	secret := "my-super-secret-hmac-key"
	ticket := SignTicket(secret, "hkg-01", 2*time.Minute)
	if !VerifyTicket(secret, "hkg-01", ticket, time.Now()) {
		t.Fatal("expected valid ticket")
	}
	if VerifyTicket(secret, "tyo-01", ticket, time.Now()) {
		t.Fatal("expected node mismatch rejection")
	}
	if VerifyTicket(secret, "hkg-01", ticket, time.Now().Add(5*time.Minute)) {
		t.Fatal("expected expired ticket rejection")
	}
}

func TestZeroDiskTestFileAndUpload(t *testing.T) {
	cfg := config.Default(config.ModeStandalone).Speedtest
	cfg.TestFiles = []string{"1MB.bin"}
	h := NewHandler(cfg, nil, "local", "sec", false, []string{"*"})

	// Test zero-disk file streaming
	req := httptest.NewRequest(http.MethodGet, "/speedtest/files/1MB.bin", nil)
	rec := httptest.NewRecorder()
	h.ServeTestFile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if got := int64(rec.Body.Len()); got != 1<<20 {
		t.Fatalf("expected 1 MiB body, got %d bytes", got)
	}

	// Test upload blackhole
	payload := bytes.Repeat([]byte("A"), 256<<10)
	upReq := httptest.NewRequest(http.MethodPost, "/api/v1/speedtest/upload", io.NopCloser(bytes.NewReader(payload)))
	upRec := httptest.NewRecorder()
	h.ServeUpload(upRec, upReq)
	if upRec.Code != http.StatusOK {
		t.Fatalf("expected upload 200 OK, got %d: %s", upRec.Code, upRec.Body.String())
	}
}

func TestIPerfLeaseLifecycle(t *testing.T) {
	cfg := config.IPerf3Config{
		Enabled:      true,
		PortStart:    45201,
		PortEnd:      45202,
		LeaseSeconds: 30,
		MaxActive:    2,
	}
	mgr := NewIPerfManager(cfg)
	mgr.binPath = "/fake/iperf3"
	mgr.starter = func(ctx context.Context, _ string, _ int) (func() error, error) {
		return func() error {
			<-ctx.Done()
			return nil
		}, nil
	}
	defer mgr.Close()

	l1, err := mgr.Acquire("1.1.1.1", "203.0.113.10", "hk-1")
	if err != nil {
		t.Fatalf("acquire 1 failed: %v", err)
	}
	// Same client gets same lease idempotently
	l1Again, err := mgr.Acquire("1.1.1.1", "203.0.113.10", "hk-1")
	if err != nil || l1Again.Port != l1.Port {
		t.Fatalf("expected idempotent lease reuse, got %+v err=%v", l1Again, err)
	}

	// Second distinct client gets second port
	l2, err := mgr.Acquire("2.2.2.2", "203.0.113.10", "hk-1")
	if err != nil || l2.Port == l1.Port {
		t.Fatalf("expected distinct second port, got %+v err=%v", l2, err)
	}

	// Third distinct client hits pool exhaustion
	if _, err := mgr.Acquire("3.3.3.3", "203.0.113.10", "hk-1"); err != ErrIPerfBusy {
		t.Fatalf("expected ErrIPerfBusy, got %v", err)
	}
}
