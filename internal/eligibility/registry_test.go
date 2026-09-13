package eligibility

import (
	"testing"

	"aurumflow/internal/money"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

func TestEligibilityLifecycle(t *testing.T) {
	r := New()
	e := Advance(Entry{Canonical: "US100"}, money.MonetaryInstrumentSpec{}, false, false)
	if e.Status != worlddomain.EligAnalysis {
		t.Fatal(e)
	}
	e = Advance(Entry{Canonical: "US100", Epic: "US100"}, money.MonetaryInstrumentSpec{ValidationStatus: money.Unverified}, false, false)
	if e.Status != worlddomain.EligDiscovered {
		t.Fatal(e.Status)
	}
	spec := money.MonetaryInstrumentSpec{
		ValidationStatus: money.BrokerMetadataOnly,
		LotSize: 1, MinDealSize: 0.1, MaxDealSize: 10, SizeIncrement: 0.1, MoneyPerPriceUnit: 1,
	}
	e = Advance(Entry{Canonical: "US100", Epic: "US100"}, spec, false, false)
	if e.Status != worlddomain.EligSpecValid {
		t.Fatal(e.Status)
	}
	spec.ValidationStatus = money.RuntimeValidated
	e = Advance(Entry{Canonical: "GOLD", Epic: "GOLD", Tradeable: true, DemoHost: true}, spec, true, false)
	if e.Status != worlddomain.EligDemo {
		t.Fatal(e.Status)
	}
	e = Advance(e, spec, true, true)
	if e.Status != worlddomain.EligBlocked || e.Reason != "LIVE_PROHIBITED" {
		t.Fatal(e)
	}
	r.Set(e)
	if r.Get("GOLD").Status != worlddomain.EligBlocked {
		t.Fatal("store")
	}
}

func TestNoJumpsAndSingleAuthority(t *testing.T) {
	e := Entry{Canonical: "GOLD", Status: worlddomain.EligAnalysis}
	e = ApplyTransition(e, worlddomain.EligCalibrated, "jump")
	if e.Status != worlddomain.EligAnalysis {
		t.Fatal("jump allowed")
	}
	e = ApplyTransition(e, worlddomain.EligDiscovered, "DEMO_DISCOVERED")
	if e.Status != worlddomain.EligDiscovered {
		t.Fatal(e)
	}
	if !SameAuthority(worlddomain.EligDiscovered, worlddomain.EligNotCalibrated) {
		t.Fatal("GOLD must not diverge DEMO_NOT_CALIBRATED vs DEMO_DISCOVERED")
	}
}

func TestApplyToWorldSingleAuthority(t *testing.T) {
	r := New()
	r.Set(Entry{Canonical: "GOLD", Epic: "GOLD", Status: worlddomain.EligDiscovered, Reason: "DEMO_DISCOVERED"})
	ws := worldstate.WorldState{Markets: map[string]worldstate.MarketState{
		"GOLD": {Market: "GOLD", Eligibility: worlddomain.EligNotCalibrated, EligReason: "DEMO_NOT_CALIBRATED"},
	}}
	ws = r.ApplyToWorld(ws)
	if ws.Markets["GOLD"].Eligibility != worlddomain.EligDiscovered {
		t.Fatal(ws.Markets["GOLD"])
	}
	if !SameAuthority(ws.Markets["GOLD"].Eligibility, worlddomain.EligDiscovered) {
		t.Fatal("diverged")
	}
}

func TestMarketSpecificSpecNotShared(t *testing.T) {
	gold := money.MonetaryInstrumentSpec{Epic: "GOLD", MoneyPerPriceUnit: 1.0}
	us100 := money.MonetaryInstrumentSpec{Epic: "US100", MoneyPerPriceUnit: 0}
	if gold.MoneyPerPriceUnit == us100.MoneyPerPriceUnit {
		t.Fatal("shared MPU")
	}
}
