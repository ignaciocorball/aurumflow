package demoprep

import (
	"testing"

	"aurumflow/internal/instrument"
	"aurumflow/internal/worlddomain"
)

func TestQueueDoesNotCalibrate(t *testing.T) {
	q := Queue([]Row{
		{Market: "US100", Epic: "US100", Discovered: true, Tradeable: true, Eligibility: worlddomain.EligDiscovered},
		{Market: "GOLD", Epic: "GOLD", Discovered: true, Tradeable: false, Eligibility: worlddomain.EligDiscovered},
	}, nil)
	if q[0].Market != "US100" {
		t.Fatal(q)
	}
	if len(ReadyNames(q)) != 1 {
		t.Fatal(ReadyNames(q))
	}
}

func TestCalStateAndRecommend(t *testing.T) {
	if CalState(Row{}) != CalUnresolved {
		t.Fatal("unresolved")
	}
	if CalState(Row{Market: "SILVER", Epic: "SILVER", Discovered: true, SpecValid: true, MoneyConfidence: "UNKNOWN", MonetaryMeta: "UNKNOWN", Tradeable: true}) != CalRiskUnknown {
		t.Fatal("risk")
	}
	next, reason := RecommendAfterGold([]Row{{Market: "SILVER", Epic: "SILVER", Discovered: true, Tradeable: false, Legacy: "LEGACY_NEEDS_CONFIG"}})
	if next != "SILVER" || reason == "" {
		t.Fatal(next, reason)
	}
}

func TestUniverseUnresolvedUS500(t *testing.T) {
	rows := UniverseRows([]instrument.Mapping{{Canonical: "GOLD", Epic: "GOLD"}, {Canonical: "US500", Unresolved: true}})
	var us Row
	for _, r := range rows {
		if r.Market == "US500" {
			us = r
		}
	}
	if us.Discovered || us.Eligibility != worlddomain.EligAnalysis {
		t.Fatal(us)
	}
}
