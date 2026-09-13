package orchestrator

import (
	"reflect"
	"testing"

	"aurumflow/internal/opportunity"
	"aurumflow/internal/shadow"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

func TestOrchestratorNoBrokerMutation(t *testing.T) {
	d := New()
	if d.CanMutateBroker() || d.Mode != Mode {
		t.Fatal(d)
	}
	if shadow.HasExecutionProvider(d) {
		t.Fatal("execution provider")
	}
	typ := reflect.TypeOf(*d)
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Type.String() == "execution.ExecutionProvider" {
			t.Fatal("field")
		}
	}
	p := d.Propose(worldstate.WorldState{Confidence: 0.7}, opportunity.Ranked{
		Market: "GOLD", Score: 80,
		State: worldstate.MarketState{Market: "GOLD", PriceTrend: "UP", Eligibility: worlddomain.EligNotCalibrated, DataQuality: worlddomain.HealthHealthy},
	}, "", "")
	if p.Eligibility == worlddomain.EligDemo || p.Direction == "" {
		t.Fatal(p)
	}
	if p.Setup == worlddomain.SetupPotential && p.Eligibility == worlddomain.EligDemo {
		t.Fatal("collapsed")
	}
}

func TestAttentionSetupCandidateNotOrder(t *testing.T) {
	d := New()
	ws := worldstate.WorldState{Confidence: 0.8}
	r := opportunity.Ranked{Market: "GOLD", Score: 90, Coverage: 82, State: worldstate.MarketState{Market: "GOLD", PriceTrend: "UP", DataQuality: worlddomain.HealthHealthy, Eligibility: worlddomain.EligAnalysis}}
	watch := d.Propose(ws, r, "", "")
	if watch.Decision != "WATCH" || watch.Setup == worlddomain.SetupPotential {
		t.Fatalf("attention became setup: %+v", watch)
	}
	setup := d.ProposeFull(ws, r, "LONG", "", worlddomain.EligNotCalibrated, nil)
	if setup.Decision != "BLOCKED" || setup.Setup != worlddomain.SetupPotential {
		t.Fatalf("setup collapsed: %+v", setup)
	}
	cand := d.ProposeFull(ws, r, "LONG", "", worlddomain.EligCalibrated, nil)
	if cand.Decision != "DEMO_CANDIDATE" {
		t.Fatalf("%+v", cand)
	}
	if d.CanMutateBroker() {
		t.Fatal("order")
	}
}
