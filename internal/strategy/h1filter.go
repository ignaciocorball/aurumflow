package strategy

// H1Allowed computes whether a signal is allowed by the H1 bias filter.
// Caller must only call when UseH1Filter is true and trendH1 is non-empty.
// blockH1Range: if true, RANGE always rejects (no crypto override).
// crypto: if true, RANGE may be allowed when finalScore >= scoreThreshold + extra.
// h1RangeExtraScore: extra points required for RANGE when crypto=true; if <= 0 treated as 1.
func H1Allowed(blockH1Range, crypto bool, h1RangeExtraScore int, trendH1, direction string, finalScore, scoreThreshold int) bool {
	if blockH1Range && trendH1 == "RANGE" {
		return false
	}
	allowed := (direction == "BUY" && trendH1 == "BULLISH") || (direction == "SELL" && trendH1 == "BEARISH")
	if allowed {
		return true
	}
	if trendH1 == "RANGE" && crypto {
		extra := h1RangeExtraScore
		if extra <= 0 {
			extra = 1
		}
		minScore := scoreThreshold + extra
		return finalScore >= minScore
	}
	return false
}
