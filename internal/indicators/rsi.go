package indicators

import (
	"math"

	"aurumflow/pkg/models"
)

// RSI computes the Relative Strength Index for the last candle (simple average of gains/losses).
// Period is typically 9. Returns 0-100; NaN if not enough data.
func RSI(candles []models.Candle, period int) float64 {
	if period <= 0 || len(candles) < period+1 {
		return math.NaN()
	}
	gains := 0.0
	losses := 0.0
	for i := len(candles) - period; i < len(candles); i++ {
		chg := candles[i].Close - candles[i-1].Close
		if chg > 0 {
			gains += chg
		} else {
			losses -= chg
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// RSIWilder computes RSI using Wilder's smoothing (matches TradingView RSI).
// First average = SMA of gains/losses over period; then avgGain = (prevAvgGain*(period-1) + gain)/period.
func RSIWilder(candles []models.Candle, period int) float64 {
	if period <= 0 || len(candles) < period+1 {
		return math.NaN()
	}
	start := len(candles) - period - 1
	// First average: SMA of first period gains/losses
	gains := 0.0
	losses := 0.0
	for i := start + 1; i <= start+period; i++ {
		chg := candles[i].Close - candles[i-1].Close
		if chg > 0 {
			gains += chg
		} else {
			losses -= chg
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	// Wilder smoothing for remaining candles
	for i := start + period + 1; i < len(candles); i++ {
		chg := candles[i].Close - candles[i-1].Close
		gain, loss := 0.0, 0.0
		if chg > 0 {
			gain = chg
		} else {
			loss = -chg
		}
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
	}
	if avgLoss == 0 {
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// InBuyZone returns true if RSI is in the pullback buy zone (38-42).
func InBuyZone(rsi float64, low, high float64) bool {
	if math.IsNaN(rsi) {
		return false
	}
	if low == 0 && high == 0 {
		low, high = 38, 42
	}
	return rsi >= low && rsi <= high
}

// InSellZone returns true if RSI is in the pullback sell zone (58-62).
func InSellZone(rsi float64, low, high float64) bool {
	if math.IsNaN(rsi) {
		return false
	}
	if low == 0 && high == 0 {
		low, high = 58, 62
	}
	return rsi >= low && rsi <= high
}
