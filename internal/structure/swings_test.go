package structure

import (
	"testing"
	"time"

	"aurumflow/pkg/models"
)

func TestDetectSwings(t *testing.T) {
	// Zigzag: low at 2, high at 5, low at 8, high at 11
	candles := make([]models.Candle, 15)
	for i := range candles {
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   100,
			High:   100,
			Low:    100,
			Close:  100,
			Volume: 100,
		}
	}
	candles[2].Low, candles[2].Close = 95, 95
	candles[5].High, candles[5].Close = 108, 108
	candles[8].Low, candles[8].Close = 94, 94
	candles[11].High, candles[11].Close = 109, 109
	swings := DetectSwings(candles, 2)
	if len(swings) == 0 {
		t.Fatal("expected at least one swing")
	}
	hasHigh := false
	hasLow := false
	for _, s := range swings {
		if s.Type == SwingHigh {
			hasHigh = true
		}
		if s.Type == SwingLow {
			hasLow = true
		}
	}
	if !hasHigh || !hasLow {
		t.Errorf("expected both high and low swings, got %d swings", len(swings))
	}
}

func TestDetectSwings_NotEnoughData(t *testing.T) {
	candles := make([]models.Candle, 5)
	swings := DetectSwings(candles, 5)
	if swings != nil {
		t.Errorf("expected nil for insufficient data, got %d swings", len(swings))
	}
}

func TestLastSwingHighLow(t *testing.T) {
	swings := []models.Swing{
		{Index: 1, Price: 100, Type: SwingLow},
		{Index: 3, Price: 105, Type: SwingHigh},
		{Index: 5, Price: 102, Type: SwingLow},
		{Index: 7, Price: 108, Type: SwingHigh},
	}
	h, ok := LastSwingHigh(swings)
	if !ok || h.Price != 108 {
		t.Errorf("last high expected 108, got ok=%v price=%f", ok, h.Price)
	}
	l, ok := LastSwingLow(swings)
	if !ok || l.Price != 102 {
		t.Errorf("last low expected 102, got ok=%v price=%f", ok, l.Price)
	}
}
