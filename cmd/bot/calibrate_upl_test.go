package main

import (
	"testing"

	"aurumflow/internal/money"
)

func TestDemoWeekEligibleRequiresRuntimeAndFlat(t *testing.T) {
	spec := money.MonetaryInstrumentSpec{ValidationStatus: money.BrokerMetadataOnly, MoneyPerPriceUnit: 1}
	if demoWeekEligible(spec, 0) {
		t.Fatal("metadata only")
	}
	spec.ValidationStatus = money.RuntimeValidated
	if demoWeekEligible(spec, 1) {
		t.Fatal("open position")
	}
	if !demoWeekEligible(spec, 0) {
		t.Fatal("expected eligible")
	}
}
