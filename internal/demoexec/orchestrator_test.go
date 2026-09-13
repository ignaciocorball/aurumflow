package demoexec

import (
	"testing"

	"aurumflow/internal/orchestrator"
	"aurumflow/internal/worlddomain"
)

func TestMultiDemoOffRefuses(t *testing.T) {
	t.Setenv(EnvFlag, "OFF")
	o := New()
	r := o.Accept(orchestrator.DecisionProposal{
		Decision: "DEMO_CANDIDATE", Eligibility: worlddomain.EligDemo,
	}, false)
	if r.Accepted || !stringsHas(r.Reason, "OFF") {
		t.Fatal(r)
	}
	if o.CanMutateBroker() {
		t.Fatal("broker")
	}
}

func TestLiveImpossible(t *testing.T) {
	o := &Orchestrator{Enabled: true}
	r := o.Accept(orchestrator.DecisionProposal{Decision: "DEMO_CANDIDATE", Eligibility: worlddomain.EligDemo}, true)
	if r.Accepted {
		t.Fatal(r)
	}
}

func stringsHas(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || (len(s) > 0 && contains(s, sub)))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
