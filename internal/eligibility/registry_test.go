package eligibility

import (
	"testing"

	"aurumflow/internal/money"
	"aurumflow/internal/worlddomain"
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
	if e.Status != worlddomain.EligLiveProhibited {
		t.Fatal(e)
	}
	r.Set(e)
	if r.Get("GOLD").Status != worlddomain.EligLiveProhibited {
		t.Fatal("store")
	}
}

func TestMarketSpecificSpecNotShared(t *testing.T) {
	gold := money.MonetaryInstrumentSpec{Epic: "GOLD", MoneyPerPriceUnit: 1.0}
	us100 := money.MonetaryInstrumentSpec{Epic: "US100", MoneyPerPriceUnit: 0}
	if gold.MoneyPerPriceUnit == us100.MoneyPerPriceUnit {
		t.Fatal("shared MPU")
	}
}
