package dataint

import "testing"

func TestClassifyDegradedOnDrops(t *testing.T) {
	in := Classify(966, 0, 2, true)
	if in.Status != Degraded || in.Reason != "BUS_BACKPRESSURE" {
		t.Fatalf("%+v", in)
	}
	ok := Classify(0, 0, 1, true)
	if ok.Status != Valid {
		t.Fatalf("%+v", ok)
	}
	bad := Classify(1, 1, 1, false)
	if bad.Status != Invalid {
		t.Fatalf("%+v", bad)
	}
}
