package structure

import (
	"testing"
	"time"

	"aurumflow/pkg/models"
)

func TestBuildStructure(t *testing.T) {
	candles := make([]models.Candle, 20)
	for i := range candles {
		candles[i] = models.Candle{
			Time:   time.Unix(int64(i)*3600, 0),
			Open:   100,
			High:   100 + float64(i),
			Low:    100 - float64(i),
			Close:  100,
			Volume: 100,
		}
	}
	// Create swings: HH + HL = bullish
	swings := []models.Swing{
		{Index: 2, Price: 98, Type: SwingLow},
		{Index: 5, Price: 102, Type: SwingHigh},
		{Index: 8, Price: 100, Type: SwingLow},  // HL
		{Index: 11, Price: 105, Type: SwingHigh}, // HH
	}
	ms := BuildStructure(candles, swings)
	if ms.Trend != TrendBullish {
		t.Errorf("expected BULLISH, got %s", ms.Trend)
	}
	if !ms.HH || !ms.HL {
		t.Errorf("expected HH and HL true, got HH=%v HL=%v", ms.HH, ms.HL)
	}
}

func TestBuildStructure_Bearish(t *testing.T) {
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
	swings := []models.Swing{
		{Index: 2, Price: 102, Type: SwingHigh},
		{Index: 5, Price: 98, Type: SwingLow},
		{Index: 8, Price: 100, Type: SwingHigh},  // LH
		{Index: 11, Price: 96, Type: SwingLow},    // LL
	}
	ms := BuildStructure(candles, swings)
	if ms.Trend != TrendBearish {
		t.Errorf("expected BEARISH, got %s", ms.Trend)
	}
	if !ms.LH || !ms.LL {
		t.Errorf("expected LH and LL true, got LH=%v LL=%v", ms.LH, ms.LL)
	}
}

func TestStructureStateFromSwings(t *testing.T) {
	candles := make([]models.Candle, 15)
	for i := range candles {
		candles[i] = models.Candle{Time: time.Unix(int64(i), 0), Open: 100, High: 101, Low: 99, Close: 100, Volume: 100}
	}
	candles[14].Close = 103
	ms := models.MarketStructure{Trend: TrendBullish}
	swings := []models.Swing{
		{Index: 5, Price: 98, Type: SwingLow},
		{Index: 10, Price: 104, Type: SwingHigh},
	}
	ss := StructureStateFromSwings(ms, candles, swings)
	if ss.Direction != TrendBullish {
		t.Errorf("expected direction BULLISH, got %s", ss.Direction)
	}
}
