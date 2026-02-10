package structure

import (
	"aurumflow/pkg/models"
)

const (
	SwingHigh = "HIGH"
	SwingLow  = "LOW"
)

// DetectSwings finds pivot highs and lows with the given lookback (e.g. 5).
// A pivot high: high[i] is max of high[i-lookback..i+lookback].
// A pivot low: low[i] is min of low[i-lookback..i+lookback].
func DetectSwings(candles []models.Candle, lookback int) []models.Swing {
	if lookback <= 0 || len(candles) < 2*lookback+1 {
		return nil
	}
	var swings []models.Swing
	for i := lookback; i < len(candles)-lookback; i++ {
		// Pivot high
		isHigh := true
		for j := i - lookback; j <= i+lookback; j++ {
			if j == i {
				continue
			}
			if candles[j].High >= candles[i].High {
				isHigh = false
				break
			}
		}
		if isHigh {
			swings = append(swings, models.Swing{Index: i, Price: candles[i].High, Type: SwingHigh})
		}
		// Pivot low
		isLow := true
		for j := i - lookback; j <= i+lookback; j++ {
			if j == i {
				continue
			}
			if candles[j].Low <= candles[i].Low {
				isLow = false
				break
			}
		}
		if isLow {
			swings = append(swings, models.Swing{Index: i, Price: candles[i].Low, Type: SwingLow})
		}
	}
	return swings
}

// LastSwingHigh returns the most recent swing high from the swings slice (by index).
func LastSwingHigh(swings []models.Swing) (models.Swing, bool) {
	var last models.Swing
	found := false
	for i := len(swings) - 1; i >= 0; i-- {
		if swings[i].Type == SwingHigh {
			last = swings[i]
			found = true
			break
		}
	}
	return last, found
}

// LastSwingLow returns the most recent swing low from the swings slice (by index).
func LastSwingLow(swings []models.Swing) (models.Swing, bool) {
	var last models.Swing
	found := false
	for i := len(swings) - 1; i >= 0; i-- {
		if swings[i].Type == SwingLow {
			last = swings[i]
			found = true
			break
		}
	}
	return last, found
}
