package strategy

import (
	"aurumflow/pkg/models"
	"testing"
)

func TestHasSweep(t *testing.T) {
	// When LiquidityEvent is set, HasSweep is true
	in := ComposerInput{LiquidityEvent: &models.LiquidityEvent{Type: "BUY_SIDE", Strength: 0.25}}
	if !HasSweep(in) {
		t.Error("HasSweep: expected true when LiquidityEvent != nil")
	}

	// When SweepBuy is true and LiquidityEvent is nil (fallback path), HasSweep is true
	in = ComposerInput{SweepBuy: true, LiquidityEvent: nil}
	if !HasSweep(in) {
		t.Error("HasSweep: expected true when SweepBuy=true and LiquidityEvent=nil")
	}

	// When SweepSell is true and LiquidityEvent is nil, HasSweep is true
	in = ComposerInput{SweepSell: true, LiquidityEvent: nil}
	if !HasSweep(in) {
		t.Error("HasSweep: expected true when SweepSell=true and LiquidityEvent=nil")
	}

	// When none of the three are set, HasSweep is false
	in = ComposerInput{}
	if HasSweep(in) {
		t.Error("HasSweep: expected false when no sweep")
	}
}

// TestSignalSLTPAlwaysValid ensures the composer never returns invalid SL/TP geometry:
// BUY => SL < entry and TP > entry; SELL => SL > entry and TP < entry (P0-A sanity gate).
func TestSignalSLTPAlwaysValid(t *testing.T) {
	// BUY with entry above Fib.High so Fib-based TP would be below entry → sanity gate recalc with ATR.
	in := ComposerInput{
		Candles:          []models.Candle{{Close: 100}},
		LiquidityEvent:   &models.LiquidityEvent{Type: "BUY_SIDE", Strength: 0.25},
		Structure:       models.MarketStructure{Trend: "BULLISH"},
		EquilibriumTouch: true,
		FibValid:        true,
		Fib:             models.Fibonacci{Low: 90, High: 95, ImpulseDirection: "UP"},
		ATR:             2,
	}
	sig, ok := SignalComposerWithZones(in, 1, 500, 6, 38, 42, 58, 62, false, "BOOST")
	if !ok {
		t.Fatal("expected signal")
	}
	if sig.Direction != "BUY" {
		t.Errorf("expected BUY; got %s", sig.Direction)
	}
	if sig.StopLoss >= sig.Entry {
		t.Errorf("BUY: SL must be < entry; got SL=%v entry=%v", sig.StopLoss, sig.Entry)
	}
	if sig.TakeProfit <= sig.Entry {
		t.Errorf("BUY: TP must be > entry; got TP=%v entry=%v", sig.TakeProfit, sig.Entry)
	}

	// SELL with entry below Fib.Low so Fib-based TP would be above entry → sanity gate recalc.
	in2 := ComposerInput{
		Candles:          []models.Candle{{Close: 100}},
		LiquidityEvent:   &models.LiquidityEvent{Type: "SELL_SIDE", Strength: 0.25},
		Structure:       models.MarketStructure{Trend: "BEARISH"},
		EquilibriumTouch: true,
		FibValid:        true,
		Fib:             models.Fibonacci{Low: 95, High: 105, ImpulseDirection: "DOWN"},
		ATR:             2,
	}
	sig2, ok2 := SignalComposerWithZones(in2, 1, 500, 6, 38, 42, 58, 62, false, "BOOST")
	if !ok2 {
		t.Fatal("expected SELL signal")
	}
	if sig2.Direction != "SELL" {
		t.Errorf("expected SELL; got %s", sig2.Direction)
	}
	if sig2.StopLoss <= sig2.Entry {
		t.Errorf("SELL: SL must be > entry; got SL=%v entry=%v", sig2.StopLoss, sig2.Entry)
	}
	if sig2.TakeProfit >= sig2.Entry {
		t.Errorf("SELL: TP must be < entry; got TP=%v entry=%v", sig2.TakeProfit, sig2.Entry)
	}
}

// TestATRBucket1218Score asserts that ATR in [12, 18] adds +1 to the score (optimal bucket per backtest).
func TestATRBucket1218Score(t *testing.T) {
	baseInput := ComposerInput{
		Candles:          []models.Candle{{Close: 100}},
		LiquidityEvent:   &models.LiquidityEvent{Type: "BUY_SIDE", Strength: 0.25},
		Structure:       models.MarketStructure{Trend: "BULLISH"},
		EquilibriumTouch: true,
		ATR:              0, // set per case
	}
	minATR, maxATR := 5.0, 80.0
	threshold := 6

	// ATR=7: valid, outside 12-18 and outside 8-12 → no bucket bonus/penalty
	in7 := baseInput
	in7.ATR = 7
	score7, _, _ := SignalDiagnostics(in7, minATR, maxATR, threshold, 38, 42, 58, 62, false, "BOOST")

	// ATR=15: valid and in 12-18 → +1 bucket bonus
	in15 := baseInput
	in15.ATR = 15
	score15, _, _ := SignalDiagnostics(in15, minATR, maxATR, threshold, 38, 42, 58, 62, false, "BOOST")

	if score15 != score7+1 {
		t.Errorf("ATR 12-18 bucket should add +1: score(ATR=7)=%d score(ATR=15)=%d", score7, score15)
	}

	// ATR=18: edge of bucket still gets +1
	in18 := baseInput
	in18.ATR = 18
	score18, _, _ := SignalDiagnostics(in18, minATR, maxATR, threshold, 38, 42, 58, 62, false, "BOOST")
	if score18 != score15 {
		t.Errorf("ATR=18 should get same bonus as ATR=15: score15=%d score18=%d", score15, score18)
	}

	// ATR=12: lower edge
	in12 := baseInput
	in12.ATR = 12
	score12, _, _ := SignalDiagnostics(in12, minATR, maxATR, threshold, 38, 42, 58, 62, false, "BOOST")
	if score12 != score15 {
		t.Errorf("ATR=12 should get same bonus as ATR=15: score12=%d score15=%d", score12, score15)
	}
}

// TestATRBucket812Penalty asserts that ATR in [8, 12) adds -1 to the score (underperforming bucket per backtest).
func TestATRBucket812Penalty(t *testing.T) {
	baseInput := ComposerInput{
		Candles:          []models.Candle{{Close: 100}},
		LiquidityEvent:   &models.LiquidityEvent{Type: "BUY_SIDE", Strength: 0.25},
		Structure:       models.MarketStructure{Trend: "BULLISH"},
		EquilibriumTouch: true,
		ATR:              0,
	}
	minATR, maxATR := 5.0, 80.0
	threshold := 6

	// ATR=7: valid, bucket 0-8 → no penalty
	in7 := baseInput
	in7.ATR = 7
	score7, _, _ := SignalDiagnostics(in7, minATR, maxATR, threshold, 38, 42, 58, 62, false, "BOOST")

	// ATR=10: valid, bucket 8-12 → -1 penalty
	in10 := baseInput
	in10.ATR = 10
	score10, _, _ := SignalDiagnostics(in10, minATR, maxATR, threshold, 38, 42, 58, 62, false, "BOOST")

	if score10 != score7-1 {
		t.Errorf("ATR 8-12 bucket should add -1: score(ATR=7)=%d score(ATR=10)=%d", score7, score10)
	}

	// ATR=8: lower edge of penalty bucket
	in8 := baseInput
	in8.ATR = 8
	score8, _, _ := SignalDiagnostics(in8, minATR, maxATR, threshold, 38, 42, 58, 62, false, "BOOST")
	if score8 != score7-1 {
		t.Errorf("ATR=8 should get same penalty as ATR=10: score7=%d score8=%d", score7, score8)
	}
}
