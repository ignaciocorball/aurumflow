package indicators

import (
	"math"
	"testing"
	"time"

	"aurumflow/pkg/models"
)

func TestRSI(t *testing.T) {
	// Synthetic: 15 candles, last 10 up then 5 down
	candles := make([]models.Candle, 15)
	base := 100.0
	for i := range candles {
		candles[i] = models.Candle{
			Time:  time.Unix(int64(i)*3600, 0),
			Open:  base,
			High:  base + 1,
			Low:   base - 1,
			Close: base,
			Volume: 100,
		}
		if i < 10 {
			base += 0.5
		} else {
			base -= 0.3
		}
		candles[i].Close = base
	}
	rsi := RSI(candles, 9)
	if math.IsNaN(rsi) {
		t.Fatal("RSI returned NaN")
	}
	if rsi < 0 || rsi > 100 {
		t.Errorf("RSI out of range: %f", rsi)
	}
}

func TestRSI_NotEnoughData(t *testing.T) {
	candles := make([]models.Candle, 5)
	rsi := RSI(candles, 9)
	if !math.IsNaN(rsi) {
		t.Errorf("expected NaN for insufficient data, got %f", rsi)
	}
}

func TestRSIWilder(t *testing.T) {
	// Same synthetic data as TestRSI: 15 candles, period 9
	candles := make([]models.Candle, 15)
	base := 100.0
	for i := range candles {
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   base,
			High:   base + 1,
			Low:    base - 1,
			Close:  base,
			Volume: 100,
		}
		if i < 10 {
			base += 0.5
		} else {
			base -= 0.3
		}
		candles[i].Close = base
	}
	rsi := RSIWilder(candles, 9)
	if math.IsNaN(rsi) {
		t.Fatal("RSIWilder returned NaN")
	}
	if rsi < 0 || rsi > 100 {
		t.Errorf("RSIWilder out of range: %f", rsi)
	}
	// Wilder smoothing typically gives different (often smoother) values than simple RSI
	rsiSimple := RSI(candles, 9)
	if math.IsNaN(rsiSimple) {
		t.Fatal("RSI returned NaN")
	}
	// Both should be in valid range; they may or may not be equal
	_ = rsiSimple
}

func TestRSIWilder_NotEnoughData(t *testing.T) {
	candles := make([]models.Candle, 5)
	rsi := RSIWilder(candles, 9)
	if !math.IsNaN(rsi) {
		t.Errorf("expected NaN for insufficient data, got %f", rsi)
	}
}

func TestInBuyZone(t *testing.T) {
	if !InBuyZone(40, 0, 0) {
		t.Error("RSI 40 should be in default buy zone")
	}
	if InBuyZone(50, 0, 0) {
		t.Error("RSI 50 should not be in default buy zone")
	}
	if InBuyZone(math.NaN(), 38, 42) {
		t.Error("NaN should not be in buy zone")
	}
}

func TestInSellZone(t *testing.T) {
	if !InSellZone(60, 0, 0) {
		t.Error("RSI 60 should be in default sell zone")
	}
	if InSellZone(50, 0, 0) {
		t.Error("RSI 50 should not be in default sell zone")
	}
}
