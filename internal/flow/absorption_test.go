package flow

import "testing"

func TestPotentialAbsorption(t *testing.T) {
	ok, tag := PotentialAbsorption(10, 0.01, 8, 2)
	if !ok || tag != "POTENTIAL_ABSORPTION" {
		t.Fatal(ok, tag)
	}
	if ok, _ := PotentialAbsorption(1, 0.01, 8, 2); ok {
		t.Fatal("weak flow")
	}
}
