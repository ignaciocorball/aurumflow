package marketstate

import (
	"aurumflow/internal/structure"
	"aurumflow/pkg/models"
)

// TrendFromCandles returns the market trend from candles using swing detection and structure.
// Returns structure.TrendBullish ("BULLISH"), structure.TrendBearish ("BEARISH"), or structure.TrendRange ("RANGE").
// lookback is used for swing detection (e.g. 5); if <= 0, a default of 5 is used.
func TrendFromCandles(candles []models.Candle, lookback int) string {
	if len(candles) == 0 {
		return structure.TrendRange
	}
	if lookback <= 0 {
		lookback = 5
	}
	swings := structure.DetectSwings(candles, lookback)
	ms := structure.BuildStructure(candles, swings)
	return ms.Trend
}

// M5TimingResult is the result of M5 micro-structure vs signal direction (CONFIRM / REJECT / NEUTRAL).
const (
	M5TimingConfirm  = "CONFIRM"
	M5TimingReject   = "REJECT"
	M5TimingNeutral  = "NEUTRAL"
)

// M5RefinerResult holds timing and score from M5 refiner.
type M5RefinerResult struct {
	Timing string // M5TimingConfirm, M5TimingReject, M5TimingNeutral
	Score  int   // +1 favor, -1 against, 0 neutral
}

// M5TimingAndScore returns M5 micro-structure timing (CONFIRM/REJECT/NEUTRAL) and m5Score (+1/-1/0).
// direction is "BUY" or "SELL". Uses TrendFromCandles on M5 with lookback (e.g. 3–5).
// CONFIRM: direction aligned with M5 trend (BUY+BULLISH, SELL+BEARISH).
// REJECT: direction opposed (BUY+BEARISH, SELL+BULLISH).
// NEUTRAL: M5 RANGE or insufficient candles.
func M5TimingAndScore(candlesM5 []models.Candle, direction string, lookback int) M5RefinerResult {
	out := M5RefinerResult{Timing: M5TimingNeutral, Score: 0}
	if len(candlesM5) < 10 {
		return out
	}
	if lookback <= 0 {
		lookback = 5
	}
	trend := TrendFromCandles(candlesM5, lookback)
	switch {
	case (direction == "BUY" && trend == structure.TrendBullish) || (direction == "SELL" && trend == structure.TrendBearish):
		out.Timing = M5TimingConfirm
		out.Score = 1
	case (direction == "BUY" && trend == structure.TrendBearish) || (direction == "SELL" && trend == structure.TrendBullish):
		out.Timing = M5TimingReject
		out.Score = -1
	default:
		out.Timing = M5TimingNeutral
		out.Score = 0
	}
	return out
}
