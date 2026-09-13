package mktcfg

import "testing"

func TestMechanicsNotReturnFit(t *testing.T) {
	g := Of("GOLD")
	if g.Session != "LONDON+NY" || g.PricePrecision != 2 {
		t.Fatal(g)
	}
	if Of("US100").Session != "US_CASH" {
		t.Fatal(Of("US100"))
	}
}
