package microflow

import (
	"testing"
	"time"
)

func TestWindowsExpireAndPastOnly(t *testing.T) {
	e := New()
	t0 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	e.OnTrade(t0.Add(-2*time.Second), 100, 1, false)
	e.OnTrade(t0.Add(-500*time.Millisecond), 100, 2, true)
	e.OnTrade(t0, 100, 99, false)
	e.OnTrade(t0.Add(time.Second), 100, 99, false)
	s := e.Snapshot(t0)
	w1 := s.Windows["1s"]
	if w1.TradeCount != 1 || w1.SellQty != 2 {
		t.Fatalf("1s %+v", w1)
	}
	if w1.TradesPerSec != 1 {
		t.Fatalf("tps=%v", w1.TradesPerSec)
	}
	w5 := s.Windows["5s"]
	if w5.TradeCount != 2 {
		t.Fatalf("5s count=%d", w5.TradeCount)
	}
	if s.Windows["1s"].BuyQty != 0 {
		t.Fatal("t0 and later leaked into 1s")
	}
}

func TestNotionalAndQtyVelocity(t *testing.T) {
	e := New()
	t0 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	e.OnTrade(t0.Add(-time.Second), 200, 3, false)
	w := e.Snapshot(t0).Windows["1s"]
	if w.TotalNotional != 600 || w.NotionalVelocity != 600 || w.QtyVelocity != 3 {
		t.Fatalf("%+v", w)
	}
}

func TestReplayLiveParity(t *testing.T) {
	t0 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	mk := func() Snapshot {
		e := New()
		e.OnTrade(t0.Add(-3*time.Second), 100, 1, false)
		e.OnTrade(t0.Add(-time.Second), 101, 2, true)
		return e.Snapshot(t0)
	}
	a, b := mk(), mk()
	if a.Windows["5s"] != b.Windows["5s"] {
		t.Fatalf("%+v vs %+v", a.Windows["5s"], b.Windows["5s"])
	}
}
