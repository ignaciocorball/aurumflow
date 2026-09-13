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
}

func TestGroups(t *testing.T) {
	if GroupOf("US100") != GroupUSEquity || GroupOf("GOLD") != GroupPrecious || GroupOf("OIL") != GroupEnergy {
		t.Fatal("groups")
	}
}
