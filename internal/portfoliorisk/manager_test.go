package portfoliorisk

import "testing"

func TestAggregateAndCorrelation(t *testing.T) {
	m := New()
	open := []Position{{Canonical: "US100", Region: "UNITED_STATES", Asset: "EQUITIES", RiskMoney: 50, RiskKnown: true}}
	r := m.Evaluate(open, Position{Canonical: "US500", Region: "UNITED_STATES", Asset: "EQUITIES", RiskMoney: 50, RiskKnown: true})
	if r.Pass {
		t.Fatal("US100+US500 must share US_EQUITY_INDEX cap")
	}
	r = m.Evaluate(open, Position{Canonical: "GOLD", Region: "GLOBAL", Asset: "PRECIOUS_METALS", RiskMoney: 40, RiskKnown: true})
	if !r.Pass {
		t.Fatal(r.Blocks)
	}
}

func TestUnknownRiskBlocks(t *testing.T) {
	m := New()
	r := m.Evaluate(nil, Position{Canonical: "OIL", RiskKnown: false})
	if r.Pass {
		t.Fatal("unknown risk allowed")
	}
	r = m.Evaluate([]Position{{Canonical: "GOLD", RiskKnown: false}, {Canonical: "SILVER", RiskKnown: false}}, Position{Canonical: "US100", RiskKnown: false})
	if len(r.Blocks) != 1 {
		t.Fatalf("duplicate blockers: %v", r.Blocks)
	}
}

func TestGroups(t *testing.T) {
	if GroupOf("US100") != GroupUSEquity || GroupOf("GOLD") != GroupPrecious || GroupOf("OIL") != GroupEnergy || GroupOf("OIL_CRUDE") != GroupEnergy {
		t.Fatal("groups")
	}
}

func TestRiskUnitsAndRemaining(t *testing.T) {
	m := New()
	if m.Caps.Unit != RiskUnit {
		t.Fatal(m.Caps.Unit)
	}
	r := m.Evaluate([]Position{{Canonical: "GOLD", Region: "GLOBAL", Asset: "PRECIOUS_METALS", RiskMoney: 40, RiskKnown: true}}, Position{})
	if r.RemainingAggregate != 260 {
		t.Fatal(r.RemainingAggregate)
	}
	if r.Unit != RiskUnit {
		t.Fatal(r.Unit)
	}
}

func TestRegionAndAssetCaps(t *testing.T) {
	m := New()
	open := []Position{{Canonical: "GOLD", Region: "GLOBAL", Asset: "PRECIOUS_METALS", RiskMoney: 140, RiskKnown: true}}
	r := m.Evaluate(open, Position{Canonical: "SILVER", Region: "GLOBAL", Asset: "PRECIOUS_METALS", RiskMoney: 20, RiskKnown: true})
	if r.Pass {
		t.Fatal("asset/region cap")
	}
}
