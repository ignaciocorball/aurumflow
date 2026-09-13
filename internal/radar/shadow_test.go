package radar

import "testing"

func TestShadowCannotMutate(t *testing.T) {
	if MayMutateBroker(ModeShadow) || MayMutateBroker(ModeOff) || MayMutateBroker("AUTO_EXECUTE") {
		t.Fatal("radar must never mutate")
	}
}
