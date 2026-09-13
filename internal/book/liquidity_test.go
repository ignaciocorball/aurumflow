package book

import "testing"

func TestLiquidityNeedsPersistence(t *testing.T) {
	h := LiquidityHeuristics{}
	f := h.Observe(100, 101, 50, 10, 5)
	if f.PersistentBid || f.DepletionBid {
		t.Fatalf("first observation must not signal: %+v", f)
	}
	f = h.Observe(100, 101, 50, 10, 5)
	f = h.Observe(100, 101, 50, 10, 5)
	if !f.PersistentBid {
		t.Fatal("expected persistent bid after 3 observations")
	}
	f = h.Observe(100, 101, 10, 10, 5)
	if !f.DepletionBid {
		t.Fatal("expected bid depletion")
	}
	f = h.Observe(100, 101, 40, 10, 5)
	if !f.ReplenishBid {
		t.Fatal("expected replenish")
	}
}

func TestBookDepth(t *testing.T) {
	b := New()
	b.ApplySnapshot(1, []Level{{10, 2}, {9, 3}}, []Level{{11, 4}, {12, 1}})
	bd, ad := b.Depth(5)
	if bd != 5 || ad != 5 {
		t.Fatalf("depth %v %v", bd, ad)
	}
}
