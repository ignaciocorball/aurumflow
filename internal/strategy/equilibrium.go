package strategy

import (
	"aurumflow/internal/structure"
	"aurumflow/pkg/models"
)

// EquilibriumCalculator computes the midpoint of the last impulse (≈ 50% Fib).
// Price in this zone = "equilibrium touch" for entry.
func EquilibriumCalculator(swings []models.Swing, candles []models.Candle) (midpoint float64, valid bool) {
	high, okHigh := structure.LastSwingHigh(swings)
	low, okLow := structure.LastSwingLow(swings)
	if !okHigh || !okLow {
		return 0, false
	}
	midpoint = (high.Price + low.Price) / 2
	return midpoint, true
}

// PriceInEquilibriumZone returns true if price is within tolerance of the equilibrium (50%) level.
func PriceInEquilibriumZone(price, midpoint float64, tolerancePct float64) bool {
	if tolerancePct <= 0 {
		tolerancePct = 0.005
	}
	band := midpoint * tolerancePct
	if band < 0.5 {
		band = 0.5
	}
	return price >= midpoint-band && price <= midpoint+band
}

// EquilibriumTouch returns true if the last close is in the equilibrium zone.
func EquilibriumTouch(candles []models.Candle, swings []models.Swing, tolerancePct float64) bool {
	if len(candles) == 0 {
		return false
	}
	mid, ok := EquilibriumCalculator(swings, candles)
	if !ok {
		return false
	}
	return PriceInEquilibriumZone(candles[len(candles)-1].Close, mid, tolerancePct)
}
