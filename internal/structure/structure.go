package structure

import (
	"aurumflow/pkg/models"
)

const (
	TrendBullish = "BULLISH"
	TrendBearish = "BEARISH"
	TrendRange   = "RANGE"
	PhaseImpulse = "IMPULSE"
	PhasePullback = "PULLBACK"
	PhaseExpansion = "EXPANSION"
)

// BuildStructure classifies market structure from swings: HH+HL = bullish, LH+LL = bearish.
func BuildStructure(candles []models.Candle, swings []models.Swing) models.MarketStructure {
	ms := models.MarketStructure{Trend: TrendRange}
	if len(swings) < 4 {
		return ms
	}
	// Order swings by index and take last 4 (2 highs, 2 lows) to classify.
	highs := make([]models.Swing, 0, 4)
	lows := make([]models.Swing, 0, 4)
	for _, s := range swings {
		if s.Type == SwingHigh {
			highs = append(highs, s)
		} else {
			lows = append(lows, s)
		}
	}
	if len(highs) >= 2 && len(lows) >= 2 {
		h1, h2 := highs[len(highs)-2], highs[len(highs)-1]
		l1, l2 := lows[len(lows)-2], lows[len(lows)-1]
		hh := h2.Price > h1.Price
		hl := l2.Price > l1.Price
		lh := h2.Price < h1.Price
		ll := l2.Price < l1.Price
		ms.HH = hh
		ms.HL = hl
		ms.LH = lh
		ms.LL = ll
		if hh && hl {
			ms.Trend = TrendBullish
		} else if lh && ll {
			ms.Trend = TrendBearish
		}
	}
	return ms
}

// StructureStateFromSwings derives phase and direction from structure and recent price.
func StructureStateFromSwings(ms models.MarketStructure, candles []models.Candle, swings []models.Swing) models.StructureState {
	ss := models.StructureState{Phase: PhasePullback}
	if ms.Trend == TrendBullish {
		ss.Direction = TrendBullish
	} else if ms.Trend == TrendBearish {
		ss.Direction = TrendBearish
	}
	if len(candles) == 0 || len(swings) < 2 {
		return ss
	}
	lastClose := candles[len(candles)-1].Close
	lastHigh, hasHigh := LastSwingHigh(swings)
	lastLow, hasLow := LastSwingLow(swings)
	if !hasHigh || !hasLow {
		return ss
	}
	// Simple heuristic: if price is near last swing high after a low, impulse up; near last swing low after a high, impulse down.
	if lastHigh.Index > lastLow.Index {
		if lastClose > (lastHigh.Price+lastLow.Price)/2 {
			ss.Phase = PhaseImpulse
		} else {
			ss.Phase = PhasePullback
		}
	} else {
		if lastClose < (lastHigh.Price+lastLow.Price)/2 {
			ss.Phase = PhaseImpulse
		} else {
			ss.Phase = PhasePullback
		}
	}
	return ss
}
