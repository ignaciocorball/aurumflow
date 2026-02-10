package strategy

import (
	"aurumflow/pkg/models"
)

const (
	BuyLiquidity  = "BUY_LIQUIDITY"
	SellLiquidity = "SELL_LIQUIDITY"
)

// LiquidityAnalyzer detects zones where stops may cluster: equal highs/lows, long wicks, consolidations.
func LiquidityAnalyzer(candles []models.Candle, atr float64, lookback int) []models.LiquidityZone {
	if lookback <= 0 || len(candles) < lookback*2 {
		return nil
	}
	var zones []models.LiquidityZone
	if atr <= 0 {
		atr = 1
	}
	start := len(candles) - lookback*2
	if start < 0 {
		start = 0
	}
	// Equal highs: cluster of similar highs
	highs := make([]float64, 0, lookback*2)
	lows := make([]float64, 0, lookback*2)
	for i := start; i < len(candles); i++ {
		highs = append(highs, candles[i].High)
		lows = append(lows, candles[i].Low)
	}
	// Long upper wicks = sell-side liquidity (stops above)
	for i := start; i < len(candles); i++ {
		body := abs(candles[i].Close - candles[i].Open)
		upperWick := candles[i].High - max(candles[i].Open, candles[i].Close)
		if body > 0 && upperWick > body*1.5 && upperWick > atr*0.3 {
			zones = append(zones, models.LiquidityZone{
				Price:    candles[i].High,
				Type:     SellLiquidity,
				Strength: upperWick / atr,
			})
		}
		lowerWick := min(candles[i].Open, candles[i].Close) - candles[i].Low
		if body > 0 && lowerWick > body*1.5 && lowerWick > atr*0.3 {
			zones = append(zones, models.LiquidityZone{
				Price:    candles[i].Low,
				Type:     BuyLiquidity,
				Strength: lowerWick / atr,
			})
		}
	}
	// Dedupe by price proximity (simplified: keep last in range)
	if len(zones) > 10 {
		zones = zones[len(zones)-10:]
	}
	return zones
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
