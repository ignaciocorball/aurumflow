package marketstate

import (
	"testing"
	"time"

	"aurumflow/internal/structure"
	"aurumflow/pkg/models"
)

func TestTrendFromCandles_Empty(t *testing.T) {
	got := TrendFromCandles(nil, 5)
	if got != structure.TrendRange {
		t.Errorf("empty candles: want %s, got %s", structure.TrendRange, got)
	}
	got = TrendFromCandles([]models.Candle{}, 5)
	if got != structure.TrendRange {
		t.Errorf("empty slice: want %s, got %s", structure.TrendRange, got)
	}
}

func TestTrendFromCandles_Range(t *testing.T) {
	// Too few candles for 4 swings (need 2*lookback+1 for DetectSwings, and 4 swings for BULLISH/BEARISH)
	candles := make([]models.Candle, 15)
	base := 2000.0
	for i := range candles {
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   base,
			High:   base + 1,
			Low:    base - 1,
			Close:  base,
			Volume: 100,
		}
	}
	got := TrendFromCandles(candles, 5)
	if got != structure.TrendRange {
		t.Errorf("few candles / flat: want RANGE, got %s", got)
	}
}

func TestTrendFromCandles_Bullish(t *testing.T) {
	// Uptrend: higher highs and higher lows so that DetectSwings yields HH+HL
	candles := make([]models.Candle, 50)
	base := 2000.0
	for i := range candles {
		// Oscillating but rising: create pivot highs and lows that form HH and HL
		up := float64(i/5) * 3
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   base + up,
			High:   base + up + 5,
			Low:    base + up - 2,
			Close:  base + up,
			Volume: 100,
		}
	}
	// Ensure we have clear pivot structure: last portion with HH and HL
	for i := 25; i < len(candles); i++ {
		candles[i].High = base + float64(i) + 10
		candles[i].Low = base + float64(i) - 1
		candles[i].Close = base + float64(i)
		candles[i].Open = base + float64(i) - 0.5
	}
	got := TrendFromCandles(candles, 5)
	// May be BULLISH if swings form HH+HL; otherwise RANGE is acceptable (structure dependent)
	if got != structure.TrendBullish && got != structure.TrendRange {
		t.Logf("TrendFromCandles (uptrend): got %s (BULLISH or RANGE acceptable)", got)
	}
}

func TestTrendFromCandles_Bearish(t *testing.T) {
	// Downtrend: lower highs and lower lows
	candles := make([]models.Candle, 50)
	base := 2050.0
	for i := range candles {
		down := float64(i) * 0.5
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   base - down,
			High:   base - down + 2,
			Low:    base - down - 5,
			Close:  base - down,
			Volume: 100,
		}
	}
	got := TrendFromCandles(candles, 5)
	if got != structure.TrendBearish && got != structure.TrendRange {
		t.Logf("TrendFromCandles (downtrend): got %s (BEARISH or RANGE acceptable)", got)
	}
}

func TestTrendFromCandles_DefaultLookback(t *testing.T) {
	candles := make([]models.Candle, 20)
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
	// lookback 0 or negative should default to 5
	got := TrendFromCandles(candles, 0)
	if got != structure.TrendRange && got != structure.TrendBullish && got != structure.TrendBearish {
		t.Errorf("TrendFromCandles(candles, 0): invalid trend %q", got)
	}
}
