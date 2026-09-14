package opportunity

import (
	"testing"

	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

func TestWeightsFrozenV1(t *testing.T) {
	if Weights["MacroAlignment"] != 15 || Weights["CapitalFlowAlignment"] != 15 || Weights["PositioningAsymmetry"] != 15 {
		t.Fatal(Weights)
	}
	if Weights["CrossAssetConfirmation"] != 10 || Weights["Momentum"] != 15 || Weights["VolatilitySuitability"] != 10 {
		t.Fatal(Weights)
	}
	if Weights["MicrostructureReadiness"] != 10 || Weights["DataQuality"] != 10 {
		t.Fatal(Weights)
	}
}

func TestScoreExplainableAndNoFabrication(t *testing.T) {
	empty := worldstate.WorldState{Markets: map[string]worldstate.MarketState{
		"Z": {Market: "Z", PriceTrend: "UNKNOWN", DataQuality: worlddomain.HealthUnknown, CapitalFlowContext: "UNKNOWN", Positioning: "UNKNOWN", MacroAlignment: "UNKNOWN"},
	}}
	if len(Rank(empty)) != 0 {
		t.Fatal("fabricated rank")
	}
	st := worldstate.MarketState{
		Market: "GOLD", PriceTrend: "UP", RelativeStrength: "UP", Volatility: "observed",
		CapitalFlowContext: "INFLOW", Positioning: "CFTC GOLD managed-money observed",
		MacroAlignment: "EXPANDING", MicroAvailable: false, DataQuality: worlddomain.HealthHealthy,
		Evidence: []string{"Gold ETF flows observed"}, Eligibility: worlddomain.EligNotCalibrated,
	}
	ws := worldstate.WorldState{Markets: map[string]worldstate.MarketState{"GOLD": st}}
	rs := Rank(ws)
	if len(rs) != 1 || rs[0].Score <= 0 || len(rs[0].Why) == 0 || len(rs[0].Components) == 0 {
		t.Fatalf("%+v", rs)
	}
	if !ConceptsSeparated(rs[0].Score, worlddomain.SetupNone, worlddomain.EligNotCalibrated) {
		t.Fatal("concepts collapsed")
	}
	if rs[0].State.Eligibility == worlddomain.EligDemo {
		t.Fatal("auto eligible")
	}
	if rs[0].Coverage <= 0 || rs[0].Coverage > 100 {
		t.Fatal(rs[0].Coverage)
	}
	if rs[0].EvidenceConfidence != rs[0].Coverage/100 {
		t.Fatal(rs[0].EvidenceConfidence)
	}
	if rs[0].Coverage < 50 && rs[0].EvidenceConfidence > 0.5 {
		t.Fatal("coverage must not present as high confidence")
	}
}

func TestMomentumIgnoresMagnitude(t *testing.T) {
	base := worldstate.MarketState{DataQuality: worlddomain.HealthHealthy, Volatility: "NORMAL", RelativeStrength: "FLAT"}
	up := Rank(worldstate.WorldState{Markets: map[string]worldstate.MarketState{"A": {Market: "A", PriceTrend: "STRONG_DOWN", DataQuality: base.DataQuality, Volatility: "NORMAL", RelativeStrength: "FLAT"}}})
	flat := Rank(worldstate.WorldState{Markets: map[string]worldstate.MarketState{"B": {Market: "B", PriceTrend: "FLAT", DataQuality: base.DataQuality, Volatility: "NORMAL", RelativeStrength: "FLAT"}}})
	if len(up) != 1 || len(flat) != 1 || up[0].Components["Momentum"] != flat[0].Components["Momentum"] {
		t.Fatalf("V1 momentum is known-state not magnitude: %+v %+v", up, flat)
	}
	ex := Explain(up[0].State, up[0])
	if len(ex) != 8 || ex[0].Name == "" {
		t.Fatal(ex)
	}
}

func TestCouplingHint(t *testing.T) {
	if CouplingHint(45, 50) != "ATTENTION_COVERAGE_COUPLING" {
		t.Fatal("expected coupling")
	}
	if CouplingHint(90, 40) != "" {
		t.Fatal("distinct axes")
	}
}

func TestMissingNotZeroAndCoverageSeparate(t *testing.T) {
	st := worldstate.MarketState{Market: "OIL", PriceTrend: "UP", DataQuality: worlddomain.HealthUnknown, CapitalFlowContext: "UNKNOWN", Positioning: "UNKNOWN", MacroAlignment: "UNKNOWN"}
	rs := Rank(worldstate.WorldState{Markets: map[string]worldstate.MarketState{"OIL": st}})
	if len(rs) != 1 || rs[0].Score == 0 {
		t.Fatal("missing treated as invisible/zero")
	}
	if rs[0].Coverage >= rs[0].Score && rs[0].Coverage >= 80 {
		t.Fatal("coverage collapsed into attention")
	}
	if rs[0].Coverage >= 50 {
		t.Fatalf("sparse market high coverage %v", rs[0].Coverage)
	}
}
