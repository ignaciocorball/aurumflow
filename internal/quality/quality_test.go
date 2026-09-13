package quality

import "testing"

func TestQualityGates(t *testing.T) {
	if (Report{}).Usable() {
		t.Fatal("empty not usable")
	}
	ok := Report{Rows: 100, InvalidPrice: 0, OutOfOrder: 0}
	if !ok.Usable() {
		t.Fatal("ok")
	}
	bad := Report{Rows: 100, InvalidPrice: 5}
	if bad.Usable() {
		t.Fatal("too many invalid prices")
	}
}
