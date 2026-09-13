package bookfeatures

import (
	"testing"
	"time"

	"aurumflow/internal/book"
)

func TestUnsyncedGate(t *testing.T) {
	b := book.New()
	d := Observe(b, time.Now().UTC())
	if d.Available {
		t.Fatal("unsynced features must be rejected")
	}
}

func TestDepletionReplenishPersist(t *testing.T) {
	e := New()
	b := book.New()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	b.ApplySnapshot(1, []book.Level{{Price: 100, Qty: 10}, {Price: 99, Qty: 5}}, []book.Level{{Price: 101, Qty: 8}, {Price: 102, Qty: 4}})
	e.OnBook(b, now)
	b.ApplySnapshot(2, []book.Level{{Price: 100, Qty: 2}, {Price: 99, Qty: 5}}, []book.Level{{Price: 101, Qty: 8}, {Price: 102, Qty: 4}})
	_, liq := e.OnBook(b, now.Add(time.Second))
	if liq.BidDepletion <= 0 {
		t.Fatalf("expected bid depletion %+v", liq)
	}
	b.ApplySnapshot(3, []book.Level{{Price: 100, Qty: 9}, {Price: 99, Qty: 5}}, []book.Level{{Price: 101, Qty: 1}, {Price: 102, Qty: 4}})
	_, liq2 := e.OnBook(b, now.Add(2*time.Second))
	if liq2.BidReplenishment <= 0 {
		t.Fatalf("expected replenishment %+v", liq2)
	}
	if liq2.AskDepletion <= 0 {
		t.Fatalf("expected ask depletion %+v", liq2)
	}
	if liq2.BidPersistence <= 0 {
		t.Fatalf("persistence %+v", liq2)
	}
}

func TestDirectionNormalization(t *testing.T) {
	d := Depth{Available: true, Mid: 100, Microprice: 100.1, Imb5: 0.4}
	liq := Liquidity{BidReplenishment: 3, AskReplenishment: 1, BidDepletion: 0.5, AskDepletion: 2, BidPersistence: 4, AskPersistence: 1}
	long := Response(1, d, liq)
	short := Response(-1, d, liq)
	if long.SupportingReplenishment != 3 || short.SupportingReplenishment != 1 {
		t.Fatalf("%+v %+v", long, short)
	}
	if long.OpposingDepletion != 2 || short.OpposingDepletion != 0.5 {
		t.Fatal("opposing")
	}
	if long.BookImbalanceResponse != 0.4 || short.BookImbalanceResponse != -0.4 {
		t.Fatal("imb")
	}
}

func TestPotentialSweepName(t *testing.T) {
	if PotentialSweep(10, 3, 1) != "POTENTIAL_SWEEP" {
		t.Fatal("sweep")
	}
	if PotentialSweep(1, 1, 1) != "" {
		t.Fatal("must not claim institutional")
	}
}
