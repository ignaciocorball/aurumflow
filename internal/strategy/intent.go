package strategy

import (
	"aurumflow/pkg/models"
)

// SweepOptions configures sweep detection (institutional: broke zone + wick + close inside).
// If nil, defaults are used: CloseTolerance=0.2, MinPenetration=0.12, MinWickRatio=1.4.
type SweepOptions struct {
	CloseTolerance  float64 // ATR multiplier for close-inside, e.g. 0.2
	MinPenetration  float64 // min wick in ATR units, e.g. 0.12
	MinWickRatio    float64 // min wick/body ratio, e.g. 1.4
}

// SweepDiagnostics holds the evaluation step-by-step for debug logging.
type SweepDiagnostics struct {
	WickSizeUp      float64
	WickSizeDown    float64
	BodySize        float64
	WickRatioUp     float64
	WickRatioDown   float64
	BrokeBuyZone    bool
	BrokeSellZone   bool
	CloseInsideBuy  bool
	CloseInsideSell bool
	SweepBuy        bool
	SweepSell       bool
	ZonePriceBuy    float64 // zone used for buy-side check (0 if none)
	ZonePriceSell   float64 // zone used for sell-side check (0 if none)
}

func defaultSweepOpts(opts *SweepOptions) (closeTol, minPen, minRatio float64) {
	closeTol, minPen, minRatio = 0.2, 0.12, 1.4
	if opts != nil {
		if opts.CloseTolerance > 0 {
			closeTol = opts.CloseTolerance
		}
		if opts.MinPenetration > 0 {
			minPen = opts.MinPenetration
		}
		if opts.MinWickRatio > 0 {
			minRatio = opts.MinWickRatio
		}
	}
	return closeTol, minPen, minRatio
}

// bestBuyZone returns the BUY_LIQUIDITY zone price that was swept (c.Low < zone.Price), preferring highest such Price.
func bestBuyZone(zones []models.LiquidityZone, candleLow float64) (price float64, ok bool) {
	for _, z := range zones {
		if z.Type != BuyLiquidity {
			continue
		}
		if candleLow < z.Price && (!ok || z.Price > price) {
			price, ok = z.Price, true
		}
	}
	return price, ok
}

// bestSellZone returns the SELL_LIQUIDITY zone price that was swept (c.High > zone.Price), preferring lowest such Price.
func bestSellZone(zones []models.LiquidityZone, candleHigh float64) (price float64, ok bool) {
	for _, z := range zones {
		if z.Type != SellLiquidity {
			continue
		}
		if candleHigh > z.Price && (!ok || z.Price < price) {
			price, ok = z.Price, true
		}
	}
	return price, ok
}

// IntentDetector evaluates whether the last candle was a liquidity sweep.
// With zones: institutional logic (broke zone + wick >= MinPenetration*ATR + wickRatio >= MinWickRatio + close inside zone ± ATR*CloseTolerance).
// Without zones: fallback using only wick size and ratio (no close-inside check).
// Returns sweepBuy, sweepSell, strength, and diagnostics for logging.
func IntentDetector(candles []models.Candle, zones []models.LiquidityZone, atr float64, lookback int, opts *SweepOptions) (sweepBuy, sweepSell bool, strength float64, diag *SweepDiagnostics) {
	diag = &SweepDiagnostics{}
	if lookback <= 0 || len(candles) < lookback+1 || atr <= 0 {
		return false, false, 0, diag
	}
	closeTol, minPen, minRatio := defaultSweepOpts(opts)
	i := len(candles) - 1
	c := candles[i]
	body := c.Close - c.Open
	if body < 0 {
		body = -body
	}
	if body < 1e-9 {
		body = 1e-9
	}
	upperWick := c.High - max(c.Open, c.Close)
	lowerWick := min(c.Open, c.Close) - c.Low
	rangeSize := c.High - c.Low
	if rangeSize == 0 {
		return false, false, 0, diag
	}
	wickRatioUp := upperWick / body
	wickRatioDown := lowerWick / body
	diag.WickSizeUp = upperWick
	diag.WickSizeDown = lowerWick
	diag.BodySize = body
	diag.WickRatioUp = wickRatioUp
	diag.WickRatioDown = wickRatioDown

	zoneBuy, hasBuyZone := bestBuyZone(zones, c.Low)
	zoneSell, hasSellZone := bestSellZone(zones, c.High)
	diag.BrokeBuyZone = hasBuyZone && c.Low < zoneBuy
	diag.BrokeSellZone = hasSellZone && c.High > zoneSell
	diag.ZonePriceBuy = zoneBuy
	diag.ZonePriceSell = zoneSell

	tol := atr * closeTol
	closeInsideBuy := !hasBuyZone || (c.Close >= zoneBuy-tol)
	closeInsideSell := !hasSellZone || (c.Close <= zoneSell+tol)
	diag.CloseInsideBuy = closeInsideBuy
	diag.CloseInsideSell = closeInsideSell

	// Buy-side sweep: broke buy liquidity (swept lows) + long lower wick + close back inside
	if len(zones) > 0 && hasBuyZone {
		if c.Low < zoneBuy && lowerWick >= atr*minPen && wickRatioDown >= minRatio && c.Close >= zoneBuy-tol {
			sweepBuy = true
			strength = lowerWick / atr
		}
	} else {
		// Fallback: no zones — use wick + penetration + body ratio only
		if lowerWick >= atr*minPen && wickRatioDown >= minRatio {
			bodyRatio := body / rangeSize
			if bodyRatio < 0.5 {
				sweepBuy = true
				strength = lowerWick / atr
			}
		}
	}
	diag.SweepBuy = sweepBuy

	// Sell-side sweep
	if len(zones) > 0 && hasSellZone {
		if c.High > zoneSell && upperWick >= atr*minPen && wickRatioUp >= minRatio && c.Close <= zoneSell+tol {
			sweepSell = true
			if s := upperWick / atr; s > strength {
				strength = s
			}
		}
	} else {
		if upperWick >= atr*minPen && wickRatioUp >= minRatio {
			bodyRatio := body / rangeSize
			if bodyRatio < 0.5 {
				sweepSell = true
				if s := upperWick / atr; s > strength {
					strength = s
				}
			}
		}
	}
	diag.SweepSell = sweepSell
	return sweepBuy, sweepSell, strength, diag
}

// LiquidityEventFromIntent builds a LiquidityEvent from intent detection.
func LiquidityEventFromIntent(sweepBuy, sweepSell bool, strength float64) *models.LiquidityEvent {
	if !sweepBuy && !sweepSell {
		return nil
	}
	e := &models.LiquidityEvent{Strength: strength}
	if sweepBuy {
		e.Type = "BUY_SIDE"
	} else {
		e.Type = "SELL_SIDE"
	}
	return e
}
