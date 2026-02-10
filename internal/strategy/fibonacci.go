package strategy

import (
	"aurumflow/internal/structure"
	"aurumflow/pkg/models"
)

// FibLevels are the retracement ratios (0.382, 0.5, 0.618, 0.705).
const (
	Fib382 = 0.382
	Fib50  = 0.5
	Fib618 = 0.618
	Fib705 = 0.705
)

// CalculateFib computes Fibonacci retracement levels from the last swing high and low.
// Low/High are always the minimum and maximum price (direction-aware: no inverted Fib).
// ImpulseDirection is "UP" when the last swing was a high (bullish move), "DOWN" when the last was a low (bearish move).
func CalculateFib(swings []models.Swing) (models.Fibonacci, bool) {
	var fib models.Fibonacci
	high, okHigh := structure.LastSwingHigh(swings)
	low, okLow := structure.LastSwingLow(swings)
	if !okHigh || !okLow {
		return fib, false
	}
	priceLow := low.Price
	priceHigh := high.Price
	if priceLow > priceHigh {
		priceLow, priceHigh = priceHigh, priceLow
	}
	fib.Low = priceLow
	fib.High = priceHigh
	diff := fib.High - fib.Low
	// Retracement levels from high toward low (standard 0.382, 0.5, 0.618, 0.705)
	fib.Level382 = fib.High - diff*Fib382
	fib.Level50 = fib.High - diff*Fib50
	fib.Level618 = fib.High - diff*Fib618
	fib.Level705 = fib.High - diff*Fib705
	if high.Index > low.Index {
		fib.ImpulseDirection = "UP"
	} else {
		fib.ImpulseDirection = "DOWN"
	}
	return fib, true
}

// InFibZone returns true if price is within the 0.5–0.618 zone (entry zone for pullback).
func InFibZone(price float64, fib models.Fibonacci) bool {
	if fib.High == fib.Low {
		return false
	}
	// Zone between 50% and 61.8%
	low := fib.Level618
	high := fib.Level50
	if low > high {
		low, high = high, low
	}
	return price >= low && price <= high
}

// InFibZoneExtended returns true if price is within 0.382–0.705 (wider zone).
func InFibZoneExtended(price float64, fib models.Fibonacci) bool {
	if fib.High == fib.Low {
		return false
	}
	low := fib.Level705
	high := fib.Level382
	if low > high {
		low, high = high, low
	}
	return price >= low && price <= high
}
