package flow

import (
	"testing"
	"time"
)

func TestAggressorAndCVD(t *testing.T) {
	if ClassifyAggressor(true) != "aggressive_sell" {
		t.Fatal("maker buy => taker sell")
	}
	var e Engine
	e.OnTrade(Trade{Price: 1, Qty: 2, BuyerMaker: false})
	e.OnTrade(Trade{Price: 1, Qty: 1, BuyerMaker: true})
	if e.CVD() != 1 {
		t.Fatalf("cvd=%v", e.CVD())
	}
	w := Roll([]Trade{{1, 2, false}, {1, 1, true}})
	if w.Buy != 2 || w.Sell != 1 || w.Signed != 1 {
		t.Fatalf("%+v", w)
	}
}

func TestWindows(t *testing.T) {
	now := time.Unix(100, 0).UTC()
	trs := []TimedTrade{
		{T: now.Add(-2 * time.Second), Trade: Trade{Price: 1, Qty: 5, BuyerMaker: false}},
		{T: now.Add(-30 * time.Second), Trade: Trade{Price: 1, Qty: 9, BuyerMaker: true}},
	}
	w1 := Window(trs, now, time.Second)
	if w1.Count != 0 {
		t.Fatal("1s")
	}
	w5 := Window(trs, now, 5*time.Second)
	if w5.Buy != 5 || w5.Count != 1 {
		t.Fatalf("%+v", w5)
	}
}

func TestMedianZ(t *testing.T) {
	if Median([]float64{3, 1, 2}) != 2 {
		t.Fatal("median")
	}
	if ZScore(3, 1, 1) != 2 {
		t.Fatal("z")
	}
}
