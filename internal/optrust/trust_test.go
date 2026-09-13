package optrust

import "testing"

func TestGoldNotTrustedAfterIdentityAlone(t *testing.T) {
	label, _ := ForMarket("GOLD", true, true, true, true)
	if label == Trusted {
		t.Fatal("calibration alone must not grant operational trust")
	}
	s, pct, miss := Score(Gate{Identity: true, MarketData: true, History: true, Strategy: true, Monetary: true, Risk: true, Lifecycle: true, Reconcile: true})
	if s != Trusted || pct != 100 || len(miss) != 0 {
		t.Fatal(s, pct, miss)
	}
}

func TestIsolationDegrade(t *testing.T) {
	s, _, miss := Score(Gate{})
	if s == Trusted || len(miss) == 0 {
		t.Fatal("empty gates trusted")
	}
}
