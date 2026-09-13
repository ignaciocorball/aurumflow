package absorption

import (
	"testing"

	"aurumflow/internal/bookfeatures"
	"aurumflow/internal/exhaustion"
)

func TestStatusesAndNoBroker(t *testing.T) {
	if MayMutateBroker() {
		t.Fatal("broker")
	}
	v1 := exhaustion.Snapshot{Classification: exhaustion.ClassExhaustion, DirectionalPressure: -20, Features: exhaustion.FeatureSet{ImpactFailure: 2}}
	if Observe(v1, false, 0, bookfeatures.PassiveLiquidityResponse{}).Status != Unavailable {
		t.Fatal("unavailable")
	}
	if Observe(v1, true, 9, bookfeatures.PassiveLiquidityResponse{}).Status != InsufficientData {
		t.Fatal("stale")
	}
	plr := bookfeatures.PassiveLiquidityResponse{
		SupportingReplenishment: 1, OpposingDepletion: 1, SupportingPersistence: 1,
		BookImbalanceResponse: 0.2, MicropriceResponse: 0.01,
	}
	s := Observe(v1, true, 1, plr)
	if s.Status != StronglySupportive {
		t.Fatalf("%s", s.Status)
	}
	weak := Observe(v1, true, 1, bookfeatures.PassiveLiquidityResponse{})
	if weak.Status != Mixed && weak.Status != NotSupportive {
		t.Fatalf("%s", weak.Status)
	}
}
