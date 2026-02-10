package indicators

import (
	"math"

	"aurumflow/pkg/models"
)

// ATR computes the Average True Range for the last candle.
// Period is typically 14.
func ATR(candles []models.Candle, period int) float64 {
	if period <= 0 || len(candles) < period+1 {
		return math.NaN()
	}
	sum := 0.0
	for i := len(candles) - period; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close
		tr := high - low
		if high-prevClose > tr {
			tr = high - prevClose
		}
		if prevClose-low > tr {
			tr = prevClose - low
		}
		sum += tr
	}
	return sum / float64(period)
}

// ATRValid returns true if ATR is between minATR and maxATR (avoids dead market and anomaly).
func ATRValid(atr, minATR, maxATR float64) bool {
	if math.IsNaN(atr) || minATR <= 0 || maxATR <= 0 {
		return false
	}
	return atr >= minATR && atr <= maxATR
}
