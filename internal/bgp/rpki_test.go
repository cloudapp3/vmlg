package bgp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func routinatorMock(t *testing.T, state string, vrps string) *httptest.Server {
	t.Helper()
	body := `{"validated_route":{"route":{"asn":"AS15169","prefix":"8.8.8.0/24"},` +
		`"validity":{"state":"` + state + `","reason":null,` +
		`"description":"mock","VRPs":[` + vrps + `]}},` +
		`"generate_time":1700000000}`
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/validity/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

func TestRPKIValidatorStates(t *testing.T) {
	const vrp = `{"asn":"AS15169","prefix":"8.8.8.0/24","max_length":24}`

	t.Run("valid", func(t *testing.T) {
		srv := routinatorMock(t, "valid", vrp)
		defer srv.Close()
		v := NewRPKIValidator(srv.URL, 2*time.Second)
		res := v.Validate(context.Background(), "15169", "8.8.8.0/24")
		if res.State != RPKIStateValid || len(res.VRPs) != 1 || res.VRPs[0].MaxLength != 24 {
			t.Fatalf("valid case: %+v", res)
		}
		if res.VRPs[0].ASN != "AS15169" || res.Origin != "15169" || res.Prefix != "8.8.8.0/24" {
			t.Fatalf("echo fields wrong: %+v", res)
		}
	})

	t.Run("notFound normalized", func(t *testing.T) {
		srv := routinatorMock(t, "notFound", "")
		defer srv.Close()
		v := NewRPKIValidator(srv.URL, 2*time.Second)
		if res := v.Validate(context.Background(), "AS15169", "8.8.8.0/24"); res.State != RPKIStateNotFound {
			t.Fatalf("notFound case: %+v", res)
		}
	})

	t.Run("invalid carries conflicting VRPs", func(t *testing.T) {
		srv := routinatorMock(t, "invalid", `{"asn":"AS3356","prefix":"8.8.0.0/12","max_length":24}`)
		defer srv.Close()
		v := NewRPKIValidator(srv.URL, 2*time.Second)
		res := v.Validate(context.Background(), "15169", "8.8.8.0/24")
		if res.State != RPKIStateInvalid || len(res.VRPs) != 1 || res.VRPs[0].ASN != "AS3356" {
			t.Fatalf("invalid case: %+v", res)
		}
	})

	t.Run("unreachable degrades to unavailable", func(t *testing.T) {
		dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(5 * time.Second)
		}))
		deadURL := dead.URL
		dead.Close()
		v := NewRPKIValidator(deadURL, 500*time.Millisecond)
		start := time.Now()
		res := v.Validate(context.Background(), "15169", "8.8.8.0/24")
		if res.State != RPKIStateUnavailable || res.Reason == "" {
			t.Fatalf("unreachable case: %+v", res)
		}
		if time.Since(start) > 2*time.Second {
			t.Fatalf("degradation took %v, want fast", time.Since(start))
		}
	})

	t.Run("unconfigured validator is nil", func(t *testing.T) {
		if NewRPKIValidator("  ", 3*time.Second) != nil {
			t.Fatalf("empty URL must yield nil validator")
		}
	})
}

func TestBestPathOriginASN(t *testing.T) {
	res := RouteLookupResult{
		Prefix: "8.8.8.0/24",
		Paths: []RoutePath{
			{ASPath: []ASHop{{ASN: "6939"}}},
			{Best: true, ASPath: []ASHop{{ASN: "4809"}, {ASN: "15169"}}},
		},
	}
	asn, prefix := BestPathOriginASN(res)
	if asn != "15169" || prefix != "8.8.8.0/24" {
		t.Fatalf("got %s %s, want 15169 8.8.8.0/24", asn, prefix)
	}

	// No best flag: fall back to first path with an AS_PATH.
	res2 := RouteLookupResult{Paths: []RoutePath{{ASPath: []ASHop{{ASN: "174"}}}}}
	if asn, _ := BestPathOriginASN(res2); asn != "174" {
		t.Fatalf("fallback origin = %s, want 174", asn)
	}

	// No paths at all.
	res3 := RouteLookupResult{}
	if asn, _ := BestPathOriginASN(res3); asn != "" {
		t.Fatalf("empty result should yield empty origin, got %s", asn)
	}
}
