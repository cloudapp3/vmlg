package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const DefaultRepo = "cloudapp3/vmlg"

type Options struct {
	Repo          string
	APIBaseURL    string
	DownloadBase  string
	TargetVersion string
	Executable    string
	OS            string
	Arch          string
	HTTPClient    *http.Client
}

type Result struct {
	PreviousVersion string `json:"previous_version"`
	Version         string `json:"version"`
	Asset           string `json:"asset"`
	SHA256          string `json:"sha256"`
	Updated         bool   `json:"updated"`
}

func Update(ctx context.Context, currentVersion string, opts Options) (Result, error) {
	if opts.Repo == "" {
		opts.Repo = DefaultRepo
	}
	if opts.APIBaseURL == "" {
		opts.APIBaseURL = "https://api.github.com"
	}
	if opts.DownloadBase == "" {
		opts.DownloadBase = "https://github.com"
	}
	if opts.OS == "" {
		opts.OS = runtime.GOOS
	}
	if opts.Arch == "" {
		opts.Arch = runtime.GOARCH
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}
	if opts.Executable == "" {
		exe, err := os.Executable()
		if err != nil {
			return Result{}, fmt.Errorf("locate current executable: %w", err)
		}
		resolved, err := filepath.EvalSymlinks(exe)
		if err == nil {
			exe = resolved
		}
		opts.Executable = exe
	}

	tag, err := resolveTag(ctx, opts)
	if err != nil {
		return Result{}, err
	}

	normCur := strings.TrimPrefix(strings.TrimSpace(currentVersion), "v")
	normTag := strings.TrimPrefix(tag, "v")
	if opts.TargetVersion == "" && normCur != "" && normCur == normTag {
		return Result{
			PreviousVersion: currentVersion,
			Version:         tag,
			Updated:         false,
		}, nil
	}

	ext := "tar.gz"
	if opts.OS == "windows" {
		ext = "zip"
	}
	assetName := fmt.Sprintf("vmlg_%s_%s_%s.%s", normTag, opts.OS, opts.Arch, ext)
	baseDownload := fmt.Sprintf("%s/%s/releases/download/%s", strings.TrimRight(opts.DownloadBase, "/"), opts.Repo, tag)

	checksumBytes, err := fetchBytes(ctx, opts.HTTPClient, baseDownload+"/checksums.txt", 1<<20)
	if err != nil {
		return Result{}, fmt.Errorf("download checksums.txt: %w", err)
	}
	expectedHex, err := findAssetSHA256(string(checksumBytes), assetName)
	if err != nil {
		return Result{}, err
	}

	archiveBytes, err := fetchBytes(ctx, opts.HTTPClient, baseDownload+"/"+assetName, 128<<20)
	if err != nil {
		return Result{}, fmt.Errorf("download %s: %w", assetName, err)
	}

	sum := sha256.Sum256(archiveBytes)
	gotHex := hex.EncodeToString(sum[:])
	if !strings.EqualFold(gotHex, expectedHex) {
		return Result{}, fmt.Errorf("SHA-256 mismatch for %s: expected %s, got %s", assetName, expectedHex, gotHex)
	}

	binName := "vmlg"
	if opts.OS == "windows" {
		binName = "vmlg.exe"
	}

	var binBytes []byte
	if ext == "zip" {
		binBytes, err = extractFromZip(archiveBytes, binName)
	} else {
		binBytes, err = extractFromTarGz(archiveBytes, binName)
	}
	if err != nil {
		return Result{}, err
	}

	if err := atomicReplaceExecutable(opts.Executable, binBytes); err != nil {
		return Result{}, err
	}

	return Result{
		PreviousVersion: currentVersion,
		Version:         tag,
		Asset:           assetName,
		SHA256:          gotHex,
		Updated:         true,
	}, nil
}

func resolveTag(ctx context.Context, opts Options) (string, error) {
	if t := strings.TrimSpace(opts.TargetVersion); t != "" {
		if !strings.HasPrefix(t, "v") {
			t = "v" + t
		}
		return t, nil
	}

	apiURL := fmt.Sprintf("%s/repos/%s/releases/latest", strings.TrimRight(opts.APIBaseURL, "/"), opts.Repo)
	body, err := fetchBytes(ctx, opts.HTTPClient, apiURL, 2<<20)
	if err != nil {
		return "", fmt.Errorf("query latest release: %w", err)
	}
	var doc struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &doc); err != nil || doc.TagName == "" {
		return "", errors.New("failed to parse latest release tag_name")
	}
	return doc.TagName, nil
}

func fetchBytes(ctx context.Context, client *http.Client, rawURL string, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "vmlg-updater")
	if tok := firstNonEmpty(os.Getenv("GITHUB_TOKEN"), os.Getenv("GH_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, rawURL)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxBytes))
}

func findAssetSHA256(checksumsText, assetName string) (string, error) {
	for _, line := range strings.Split(checksumsText, "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == assetName {
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", fmt.Errorf("asset %s not found in checksums.txt", assetName)
}

func extractFromTarGz(archive []byte, binName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == binName {
			return io.ReadAll(io.LimitReader(tr, 128<<20))
		}
	}
	return nil, fmt.Errorf("binary %s not found inside archive", binName)
}

func extractFromZip(archive []byte, binName string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	for _, f := range zr.File {
		if filepath.Base(f.Name) == binName {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(io.LimitReader(rc, 128<<20))
		}
	}
	return nil, fmt.Errorf("binary %s not found inside zip archive", binName)
}

func atomicReplaceExecutable(dstPath string, binBytes []byte) error {
	dir := filepath.Dir(dstPath)
	mode := os.FileMode(0o755)
	if st, err := os.Stat(dstPath); err == nil {
		mode = st.Mode().Perm()
		if mode&0o111 == 0 {
			mode |= 0o755
		}
	}

	tmp, err := os.CreateTemp(dir, ".vmlg-update-*")
	if err != nil {
		return fmt.Errorf("create temp binary in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmp.Write(binBytes); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, dstPath)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
