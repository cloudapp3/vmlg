package probe

// Streaming unlock detection — fixed five services, public no-login endpoints,
// pure Go HTTP only (zero shell / zero third-party scripts, ADR-VMLG-002).

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Unlock status enumeration. Streaming services use unlocked/locked;
// availability-style services use available/unavailable. "error" marks a
// transport failure and is never a content verdict.
const (
	UnlockUnlocked    = "unlocked"
	UnlockLocked      = "locked"
	UnlockAvailable   = "available"
	UnlockUnavailable = "unavailable"
	UnlockError       = "error"
)

// UnlockTargets holds the probe endpoints; nil fields fall back to the
// production public endpoints. Injectable for deterministic tests.
type UnlockTargets struct {
	YouTube string
	Netflix string
	Disney  string
	ChatGPT string
	TikTok  string
}

func DefaultUnlockTargets() UnlockTargets {
	return UnlockTargets{
		YouTube: "https://www.youtube.com/premium",
		Netflix: "https://www.netflix.com/title/81280792",
		Disney:  "https://www.disneyplus.com/",
		ChatGPT: "https://api.openai.com/compliance/cookie_requirements",
		TikTok:  "https://www.tiktok.com/",
	}
}

// UnlockRequest is the wire form for a node-side unlock check. There is no
// user-controlled target — the fixed five services are always probed from the
// node's own egress. Targets is a test-injection seam (never serialized).
type UnlockRequest struct {
	TimeoutMs int            `json:"timeout_ms"`
	Targets   *UnlockTargets `json:"-"`
}

type UnlockResult struct {
	Service   string  `json:"service"`
	Status    string  `json:"status"`
	Region    string  `json:"region,omitempty"`
	ElapsedMs float64 `json:"elapsed_ms"`
	Error     string  `json:"error,omitempty"`
}

type UnlockCheckResult struct {
	Results   []UnlockResult `json:"results"`
	CheckedAt int64          `json:"checked_at"` // unix seconds
}

type unlockRequest struct {
	Service string
	URL     string
	// verdict maps the HTTP exchange to {status, region}.
	Verdict func(resp *http.Response) (status, region string)
}

// RunUnlockChecks probes all five services concurrently. Detection runs on the
// node itself (agent or standalone) so the egress IP under test is the node's.
func RunUnlockChecks(ctx context.Context, timeoutMs int, targets *UnlockTargets) UnlockCheckResult {
	if targets == nil {
		def := DefaultUnlockTargets()
		targets = &def
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // inspect each hop ourselves
		},
	}

	reqs := []unlockRequest{
		{
			Service: "youtube",
			URL:     targets.YouTube,
			Verdict: func(resp *http.Response) (string, string) {
				// /premium redirects to a locale-specific page; the Location
				// carries ?gl=CC when Premium is offered there.
				loc := resp.Header.Get("Location")
				if code := queryParam(loc, "gl"); code != "" {
					return UnlockAvailable, code
				}
				if resp.StatusCode == http.StatusOK {
					return UnlockUnavailable, ""
				}
				return UnlockUnavailable, ""
			},
		},
		{
			Service: "netflix",
			URL:     targets.Netflix,
			Verdict: func(resp *http.Response) (string, string) {
				// Non-member title probe: 2xx/3xx = library reachable (region
				// from the redirect locale path), 403/401 = geo blocked.
				region := regionFromLocation(resp.Header.Get("Location"))
				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					return UnlockUnlocked, region
				}
				if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
					return UnlockLocked, ""
				}
				return UnlockUnavailable, region
			},
		},
		{
			Service: "disney",
			URL:     targets.Disney,
			Verdict: func(resp *http.Response) (string, string) {
				if resp.StatusCode == http.StatusOK {
					return UnlockUnlocked, ""
				}
				if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnavailableForLegalReasons {
					return UnlockLocked, ""
				}
				return UnlockUnavailable, ""
			},
		},
		{
			Service: "chatgpt",
			URL:     targets.ChatGPT,
			Verdict: func(resp *http.Response) (string, string) {
				// compliance endpoint: 200 = service reachable from this egress;
				// 403 = Cloudflare geo-block on OpenAI properties.
				if resp.StatusCode == http.StatusOK {
					return UnlockAvailable, ""
				}
				if resp.StatusCode == http.StatusForbidden {
					return UnlockUnavailable, ""
				}
				return UnlockUnavailable, ""
			},
		},
		{
			Service: "tiktok",
			URL:     targets.TikTok,
			Verdict: func(resp *http.Response) (string, string) {
				if resp.StatusCode >= 200 && resp.StatusCode < 400 {
					return UnlockAvailable, ""
				}
				return UnlockUnavailable, ""
			},
		},
	}

	out := UnlockCheckResult{CheckedAt: time.Now().UTC().Unix(), Results: make([]UnlockResult, len(reqs))}
	var wg sync.WaitGroup
	for i := range reqs {
		wg.Add(1)
		go func(idx int, r unlockRequest) {
			defer wg.Done()
			start := time.Now()
			res := UnlockResult{Service: r.Service}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.URL, nil)
			if err != nil {
				res.Status, res.Error = UnlockError, err.Error()
				out.Results[idx] = res
				return
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) vmlg-unlock-probe/1.0")
			resp, err := client.Do(req)
			if err != nil {
				res.Status, res.Error = UnlockError, err.Error()
				res.ElapsedMs = msSince(start)
				out.Results[idx] = res
				return
			}
			// Drain a bounded slice so the connection can be reused/closed.
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
			resp.Body.Close()

			res.Status, res.Region = r.Verdict(resp)
			res.ElapsedMs = msSince(start)
			out.Results[idx] = res
		}(i, reqs[i])
	}
	wg.Wait()
	return out
}

// queryParam extracts a value from the query string of a raw URL.
func queryParam(rawURL, key string) string {
	if i := strings.Index(rawURL, "?"); i >= 0 {
		for _, kv := range strings.FieldsFunc(rawURL[i+1:], func(r rune) bool { return r == '&' || r == '#' }) {
			if k, v, ok := strings.Cut(kv, "="); ok && strings.EqualFold(k, key) {
				return v
			}
		}
	}
	return ""
}

// regionFromLocation pulls a two-letter locale from paths like /xx-YY/ or /xx/.
func regionFromLocation(loc string) string {
	if loc == "" {
		return ""
	}
	rest := strings.TrimPrefix(loc, "https://")
	rest = strings.TrimPrefix(rest, "http://")
	if i := strings.Index(rest, "/"); i >= 0 {
		rest = rest[i+1:]
	} else {
		return ""
	}
	seg := rest
	if i := strings.Index(seg, "/"); i >= 0 {
		seg = seg[:i]
	}
	seg = strings.TrimSuffix(seg, "-en")
	if len(seg) == 2 {
		return strings.ToLower(seg)
	}
	if parts := strings.Split(seg, "-"); len(parts) == 2 && len(parts[0]) == 2 {
		return strings.ToLower(parts[0])
	}
	return ""
}
