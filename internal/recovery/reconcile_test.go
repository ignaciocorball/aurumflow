package recovery

import "testing"

func TestReconcileClasses(t *testing.T) {
	out, unk := Reconcile(
		[]BrokerPos{{DealID: "A", Epic: "GOLD"}, {DealID: "B", Epic: "BTCUSD"}},
		[]JournalPos{{DealID: "A", Epic: "GOLD"}},
	)
	if unk != 0 || out[0].Class != Managed || out[1].Class != Recovered {
		t.Fatalf("%+v unk=%d", out, unk)
	}
	if BlockNewOrders(1) != true || BlockNewOrders(0) {
		t.Fatal("block")
	}
}
