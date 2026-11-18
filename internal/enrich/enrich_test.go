package enrich

import (
	"context"
	"net/netip"
	"testing"
)

func TestCarrierAndIXPDetection(t *testing.T) {
	e := New(false, false)

	cn2 := e.Lookup(context.Background(), netip.MustParseAddr("59.43.187.10"))
	if cn2.LineTag != "CN2 GIA" || cn2.ASN != "AS4809" {
		t.Fatalf("unexpected CN2 lookup: %+v", cn2)
	}

	hkix := e.Lookup(context.Background(), netip.MustParseAddr("123.255.90.15"))
	if hkix.IXP != "HKIX" {
		t.Fatalf("expected HKIX detection, got %+v", hkix)
	}

	// CN2 GIA (59.43 without 202.97)
	vGIA := SummarizeRoute([]HopMeta{
		{ASN: "AS4809", LineTag: "CN2 GIA"},
	})
	if vGIA.Code != "CN2_GIA" || vGIA.Tier != "premium" {
		t.Fatalf("expected CN2_GIA verdict, got %+v", vGIA)
	}

	// CN2 GT (both 59.43 and 202.97)
	vGT := SummarizeRoute([]HopMeta{
		{ASN: "AS4809", LineTag: "CN2 GIA"},
		{ASN: "AS4134", LineTag: "CT 163"},
	})
	if vGT.Code != "CN2_GT" {
		t.Fatalf("expected CN2_GT verdict, got %+v", vGT)
	}

	// CU 10099 + 9929
	vCU := SummarizeRoute([]HopMeta{
		{ASN: "AS10099", LineTag: "CU 10099"},
		{ASN: "AS9929", LineTag: "CU 9929"},
	})
	if vCU.Code != "CU_9929" || vCU.Label != "CU 10099+9929" {
		t.Fatalf("expected CU 10099+9929 verdict, got %+v", vCU)
	}
}
