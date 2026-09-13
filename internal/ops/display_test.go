package ops

import "testing"

func TestDisplayMissingNotZero(t *testing.T) {
	if DisplayFloat(false, 0) != Missing {
		t.Fatal(DisplayFloat(false, 0))
	}
	if DisplayInt(false, 0) != Missing {
		t.Fatal(DisplayInt(false, 0))
	}
	if GoldMarketLabel("") != "WAITING" {
		t.Fatal(GoldMarketLabel(""))
	}
	if GoldQuotesOK("CLOSED", 0, 0) {
		t.Fatal("closed quotes")
	}
	if GoldQuotesOK("TRADEABLE", 0, 0) {
		t.Fatal("zero quotes")
	}
	if !GoldQuotesOK("TRADEABLE", 2300, 2301) {
		t.Fatal("tradeable quotes")
	}
}

func TestDecisionWhyUsesEvidenceOnly(t *testing.T) {
	st := Status{LastV1Class: "FLOW_EXHAUSTION_CONFIRM", DirectionalP: -18, LastLegacyDir: 1,
		AggSell: 10, AggBuy: 2, FlowEfficiency: -0.4, ImpactFailure: 0.6,
		BidRepl: 1.2, AskRepl: 0.1, L2QuotesOK: true, Microprice: 100.2, L2Mid: 100}
	w := DecisionWhy(st)
	if w == "" || !containsAll(w, "EXHAUSTION", "LONG", "deteriorating") {
		t.Fatal(w)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, p string) bool {
	return len(s) >= len(p) && (s == p || len(p) == 0 || (len(s) > 0 && (indexOf(s, p) >= 0)))
}

func indexOf(s, p string) int {
	for i := 0; i+len(p) <= len(s); i++ {
		if s[i:i+len(p)] == p {
			return i
		}
	}
	return -1
}
