package strategy

import (
	"aurumflow/pkg/models"
	"testing"
)

func TestComposerIgnoresRadarContext(t *testing.T) {
	in := ComposerInput{
		Candles:        []models.Candle{{Close: 100}},
		LiquidityEvent: &models.LiquidityEvent{Type: "BUY_SIDE", Strength: 0.25},
		Structure:      models.MarketStructure{Trend: "BULLISH"},
		ATR:            1.5,
		RSI:            40,
	}
	a, oka := SignalComposer(in, 0.5, 5, 1)
	in.Context = &DecisionContext{RadarMode: ModeShadowLike(), Pressure: 90, Confidence: 99, RadarState: "EXPANSION"}
	b, okb := SignalComposer(in, 0.5, 5, 1)
	if oka != okb {
		t.Fatal("radar context must not change signal presence")
	}
	if oka && (a.Direction != b.Direction || a.Score != b.Score) {
		t.Fatalf("radar changed signal a=%+v b=%+v", a, b)
	}
}

func ModeShadowLike() string { return "SHADOW" }
