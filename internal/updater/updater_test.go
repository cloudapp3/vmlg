package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelfUpdateVerifiesSHA256AndReplacesBinary(t *testing.T) {
	newBinContent := []byte("#!/bin/sh\necho vmlg-v0.2.0\n")
	archiveBytes := buildTestTarGz(t, "vmlg", newBinContent)
	sum := sha256.Sum256(archiveBytes)
	sumHex := hex.EncodeToString(sum[:])
	assetName := "vmlg_0.2.0_linux_amd64.tar.gz"

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/cloudapp3/vmlg/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v0.2.0"}`))
	})
	mux.HandleFunc("/cloudapp3/vmlg/releases/download/v0.2.0/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "%s  %s\n", sumHex, assetName)
	})
	mux.HandleFunc("/cloudapp3/vmlg/releases/download/v0.2.0/"+assetName, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archiveBytes)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	targetBin := filepath.Join(t.TempDir(), "vmlg")
	if err := os.WriteFile(targetBin, []byte("old-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := Update(context.Background(), "0.1.0", Options{
		Repo:         "cloudapp3/vmlg",
		APIBaseURL:   srv.URL,
		DownloadBase: srv.URL,
		Executable:   targetBin,
		OS:           "linux",
		Arch:         "amd64",
		HTTPClient:   srv.Client(),
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !res.Updated || res.Version != "v0.2.0" {
		t.Fatalf("unexpected update result: %+v", res)
	}

	gotBytes, err := os.ReadFile(targetBin)
	if err != nil || !bytes.Equal(gotBytes, newBinContent) {
		t.Fatalf("binary was not replaced properly")
	}

	// Verify tampered checksum is rejected
	muxBad := http.NewServeMux()
	muxBad.HandleFunc("/cloudapp3/vmlg/releases/download/v0.3.0/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "%s  vmlg_0.3.0_linux_amd64.tar.gz\n", strings.Repeat("0", 64))
	})
	muxBad.HandleFunc("/cloudapp3/vmlg/releases/download/v0.3.0/vmlg_0.3.0_linux_amd64.tar.gz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archiveBytes)
	})
	badSrv := httptest.NewServer(muxBad)
	defer badSrv.Close()

	_, err = Update(context.Background(), "0.2.0", Options{
		Repo:          "cloudapp3/vmlg",
		DownloadBase:  badSrv.URL,
		TargetVersion: "v0.3.0",
		Executable:    targetBin,
		OS:            "linux",
		Arch:          "amd64",
		HTTPClient:    badSrv.Client(),
	})
	if err == nil || !strings.Contains(err.Error(), "SHA-256 mismatch") {
		t.Fatalf("expected SHA-256 mismatch error, got: %v", err)
	}
}

func buildTestTarGz(t *testing.T, name string, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(data)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(data); err != nil {
		t.Fatal(err)
	}
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}
