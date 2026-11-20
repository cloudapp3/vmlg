package probe

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

//go:embed unlock.go
var unlockSrc string

// AC5: the unlock probe must stay pure Go — no shell, no exec, no third-party
// scripts. This embed-and-scan assertion runs in CI via `make test`.
func TestUnlockSourceHasNoShellExecution(t *testing.T) {
	for _, banned := range []string{"exec.Command", "os/exec", "syscall.Exec", "\"sh ", "\"bash "} {
		if strings.Contains(unlockSrc, banned) {
			t.Fatalf("unlock.go contains banned shell construct %q", banned)
		}
	}
}

func TestRunUnlockChecksVerdicts(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/yt":
			w.Header().Set("Location", "https://www.youtube.com/premium?gl=JP&hl=ja")
			w.WriteHeader(http.StatusFound)
		case "/nf":
			w.Header().Set("Location", "https://www.netflix.com/sg/")
			w.WriteHeader(http.StatusFound)
		case "/nf-blocked":
			w.WriteHeader(http.StatusForbidden)
		case "/disney":
			w.WriteHeader(http.StatusOK)
		case "/disney-blocked":
			w.WriteHeader(http.StatusForbidden)
		case "/chatgpt":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success":true}`))
		case "/chatgpt-blocked":
			w.WriteHeader(http.StatusForbidden)
		case "/tiktok":
			w.WriteHeader(http.StatusOK)
		case "/tiktok-down":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	res := RunUnlockChecks(context.Background(), 3000, &UnlockTargets{
		YouTube: srv.URL + "/yt",
		Netflix: srv.URL + "/nf",
		Disney:  srv.URL + "/disney",
		ChatGPT: srv.URL + "/chatgpt",
		TikTok:  srv.URL + "/tiktok",
	})
	byService := map[string]UnlockResult{}
	for _, r := range res.Results {
		byService[r.Service] = r
	}
	if len(res.Results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(res.Results))
	}
	if s := byService["youtube"]; s.Status != UnlockAvailable || s.Region != "JP" {
		t.Fatalf("youtube = %+v", s)
	}
	if s := byService["netflix"]; s.Status != UnlockUnlocked || s.Region != "sg" {
		t.Fatalf("netflix = %+v", s)
	}
	if s := byService["disney"]; s.Status != UnlockUnlocked {
		t.Fatalf("disney = %+v", s)
	}
	if s := byService["chatgpt"]; s.Status != UnlockAvailable {
		t.Fatalf("chatgpt = %+v", s)
	}
	if s := byService["tiktok"]; s.Status != UnlockAvailable {
		t.Fatalf("tiktok = %+v", s)
	}
	if res.CheckedAt == 0 {
		t.Fatalf("checked_at must be set")
	}

	// Locked / unavailable variants.
	res2 := RunUnlockChecks(context.Background(), 3000, &UnlockTargets{
		YouTube: srv.URL + "/yt",
		Netflix: srv.URL + "/nf-blocked",
		Disney:  srv.URL + "/disney-blocked",
		ChatGPT: srv.URL + "/chatgpt-blocked",
		TikTok:  srv.URL + "/tiktok-down",
	})
	by2 := map[string]UnlockResult{}
	for _, r := range res2.Results {
		by2[r.Service] = r
	}
	if by2["netflix"].Status != UnlockLocked {
		t.Fatalf("netflix blocked = %+v", by2["netflix"])
	}
	if by2["disney"].Status != UnlockLocked {
		t.Fatalf("disney blocked = %+v", by2["disney"])
	}
	if by2["chatgpt"].Status != UnlockUnavailable {
		t.Fatalf("chatgpt blocked = %+v", by2["chatgpt"])
	}
	if by2["tiktok"].Status != UnlockUnavailable {
		t.Fatalf("tiktok down = %+v", by2["tiktok"])
	}

	// Transport errors degrade to status=error, not a content verdict.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	deadURL := dead.URL
	dead.Close()
	res3 := RunUnlockChecks(context.Background(), 1000, &UnlockTargets{
		YouTube: deadURL + "/yt", Netflix: deadURL + "/nf", Disney: deadURL + "/d",
		ChatGPT: deadURL + "/c", TikTok: deadURL + "/t",
	})
	for _, r := range res3.Results {
		if r.Status != UnlockError || r.Error == "" {
			t.Fatalf("expected error status, got %+v", r)
		}
	}
}
