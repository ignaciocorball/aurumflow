package indicators

import (
	"math"
	"testing"
	"time"

	"aurumflow/pkg/models"
)

func TestATR(t *testing.T) {
	candles := make([]models.Candle, 20)
	for i := range candles {
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   100,
			High:   102 + float64(i%3),
			Low:    98 - float64(i%2),
			Close:  100,
			Volume: 100,
		}
	}
	atr := ATR(candles, 14)
	if math.IsNaN(atr) {
		t.Fatal("ATR returned NaN")
	}
	if atr <= 0 {
		t.Errorf("ATR should be positive, got %f", atr)
	}
}

func TestATR_NotEnoughData(t *testing.T) {
	candles := make([]models.Candle, 5)
	atr := ATR(candles, 14)
	if !math.IsNaN(atr) {
		t.Errorf("expected NaN for insufficient data, got %f", atr)
	}
}

func TestATRValid(t *testing.T) {
	if !ATRValid(100, 80, 350) {
		t.Error("ATR 100 should be valid")
	}
	if ATRValid(50, 80, 350) {
		t.Error("ATR 50 below min should be invalid")
	}
	if ATRValid(400, 80, 350) {
		t.Error("ATR 400 above max should be invalid")
	}
	if ATRValid(math.NaN(), 80, 350) {
		t.Error("NaN should be invalid")
	}
}
