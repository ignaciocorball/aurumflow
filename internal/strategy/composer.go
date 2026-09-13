package strategy

import (
	"fmt"
	"aurumflow/internal/indicators"
	"aurumflow/internal/structure"
	"aurumflow/pkg/models"
)

// ComposerInput holds all inputs for the signal composer.
type ComposerInput struct {
	Candles       []models.Candle
	Swings        []models.Swing
	Structure     models.MarketStructure
	StructureState models.StructureState
	Fib           models.Fibonacci
	FibValid      bool
	LiquidityZones []models.LiquidityZone
	LiquidityEvent *models.LiquidityEvent
	RSI           float64
	ATR           float64
	EquilibriumTouch bool
	SweepBuy      bool
	SweepSell     bool
	IntentStrength float64
	EntryDelayCandles int
	ImpulseCandleIndex int // index of last strong impulse candle; 0 = none
	SweepDiagnostics *SweepDiagnostics // optional; filled for SweepDebugLog
	// Context is optional radar/intelligence context. Composer scoring ignores it.
	Context *DecisionContext
}

// DecisionContext is intelligence-plane context. It must never carry broker mutation rights.
type DecisionContext struct {
	RadarMode  string
	RadarState string
	Pressure   float64
	Confidence float64
	BookSynced bool
}

// HasSweep returns true when a liquidity sweep is detected: either LiquidityEvent is set (canonical)
// or SweepBuy/SweepSell flags are true. Use this as the single source of truth for the state machine.
func HasSweep(in ComposerInput) bool {
	return in.LiquidityEvent != nil || in.SweepBuy || in.SweepSell
}

// SignalComposer scores inputs and produces a TradeSignal if score >= threshold (uses default RSI zones 38-42, 58-62).
func SignalComposer(in ComposerInput, minATR, maxATR float64, scoreThreshold int) (*models.TradeSignal, bool) {
	return SignalComposerWithZones(in, minATR, maxATR, scoreThreshold, 38, 42, 58, 62, true, "BOOST")
}

// SignalComposerWithZones is like SignalComposer but with configurable RSI zones and mode.
// rsiMode: GATE = reject if not in zone; BOOST = +1 when in zone; PENALTY = +2 in zone, 0 neutral, -1 when against (BUY+RSI high / SELL+RSI low).
func SignalComposerWithZones(in ComposerInput, minATR, maxATR float64, scoreThreshold int, rsiBuyLow, rsiBuyHigh, rsiSellLow, rsiSellHigh float64, rsiRequired bool, rsiMode string) (*models.TradeSignal, bool) {
	score := 0
	var direction string
	var entry, stopLoss, takeProfit float64

	if len(in.Candles) == 0 {
		return nil, false
	}
	lastClose := in.Candles[len(in.Candles)-1].Close

	if in.EntryDelayCandles > 0 && in.ImpulseCandleIndex > 0 {
		candlesSinceImpulse := len(in.Candles) - 1 - in.ImpulseCandleIndex
		if candlesSinceImpulse < in.EntryDelayCandles {
			return nil, false
		}
	}

	if in.LiquidityEvent != nil && in.LiquidityEvent.Strength > 0 {
		score += 3
		if in.Structure.Trend == "BULLISH" && in.LiquidityEvent.Type == "BUY_SIDE" {
			direction = "BUY"
		} else if in.Structure.Trend == "BEARISH" && in.LiquidityEvent.Type == "SELL_SIDE" {
			direction = "SELL"
		}
	}
	if direction == "" {
		if in.Structure.Trend == "BULLISH" && in.SweepBuy {
			direction = "BUY"
		} else if in.Structure.Trend == "BEARISH" && in.SweepSell {
			direction = "SELL"
		}
	}

	if (direction == "BUY" && in.Structure.Trend == "BULLISH") || (direction == "SELL" && in.Structure.Trend == "BEARISH") {
		score += 2
	}
	if in.EquilibriumTouch {
		score += 2
	}
	if indicators.ATRValid(in.ATR, minATR, maxATR) {
		score += 1
	}
	// ATR bucket 12-18 is optimal per backtest; add +1 when in range
	if in.ATR >= 12 && in.ATR <= 18 {
		score += 1
	}
	// ATR bucket 8-12 underperforms per backtest; penalize -1
	if in.ATR >= 8 && in.ATR < 12 {
		score -= 1
	}
	// RSI: GATE / BOOST / PENALTY
	if rsiBuyLow <= 0 {
		rsiBuyLow, rsiBuyHigh = 38, 42
	}
	if rsiSellLow <= 0 {
		rsiSellLow, rsiSellHigh = 58, 62
	}
	mode := rsiMode
	if mode == "" {
		if rsiRequired {
			mode = "GATE"
		} else {
			mode = "BOOST"
		}
	}
	switch mode {
	case "GATE":
		if direction == "BUY" && !indicators.InBuyZone(in.RSI, rsiBuyLow, rsiBuyHigh) {
			return nil, false
		}
		if direction == "SELL" && !indicators.InSellZone(in.RSI, rsiSellLow, rsiSellHigh) {
			return nil, false
		}
	case "BOOST":
		if direction == "BUY" && indicators.InBuyZone(in.RSI, rsiBuyLow, rsiBuyHigh) {
			score += 1
		}
		if direction == "SELL" && indicators.InSellZone(in.RSI, rsiSellLow, rsiSellHigh) {
			score += 1
		}
	case "PENALTY":
		if direction == "BUY" {
			if indicators.InBuyZone(in.RSI, rsiBuyLow, rsiBuyHigh) {
				score += 2
			} else if in.RSI >= rsiSellLow {
				score -= 1 // overbought, against buy
			}
		}
		if direction == "SELL" {
			if indicators.InSellZone(in.RSI, rsiSellLow, rsiSellHigh) {
				score += 2
			} else if in.RSI <= rsiBuyHigh {
				score -= 1 // oversold, against sell
			}
		}
	}
	if score < 0 {
		score = 0
	}
	if score < scoreThreshold || direction == "" {
		return nil, false
	}

	// Build signal: entry = last close, SL/TP from Fib (direction-aware) or ATR fallback
	entry = lastClose
	useFib := in.FibValid && (direction == "BUY" && in.Fib.ImpulseDirection == "UP" || direction == "SELL" && in.Fib.ImpulseDirection == "DOWN")
	if useFib {
		if direction == "BUY" {
			stopLoss = in.Fib.Low - in.ATR*0.2
			takeProfit = in.Fib.High + in.ATR*0.5
		} else {
			stopLoss = in.Fib.High + in.ATR*0.2
			takeProfit = in.Fib.Low - in.ATR*0.5
		}
	}
	if !useFib {
		// ATR fallback: 1R stop, 2R target
		if direction == "BUY" {
			stopLoss = lastClose - in.ATR
			takeProfit = lastClose + in.ATR*2
		} else {
			stopLoss = lastClose + in.ATR
			takeProfit = lastClose - in.ATR*2
		}
	}
	// P0-A SL/TP Sanity Gate: BUY => SL < entry and TP > entry; SELL => SL > entry and TP < entry. Else recalc with ATR.
	const minRR = 1.5
	if direction == "BUY" {
		if stopLoss >= entry || takeProfit <= entry {
			stopLoss = entry - in.ATR
			takeProfit = entry + in.ATR*minRR
		}
	} else {
		if stopLoss <= entry || takeProfit >= entry {
			stopLoss = entry + in.ATR
			takeProfit = entry - in.ATR*minRR
		}
	}

	return &models.TradeSignal{
		Direction:  direction,
		Entry:       entry,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
		Confidence: float64(score) / 10.0,
		Score:      score,
	}, true
}

// SignalDiagnostics returns partial score and reasons why signal might not be generated (for logging).
// rsiMode: GATE/BOOST/PENALTY same as SignalComposerWithZones.
func SignalDiagnostics(in ComposerInput, minATR, maxATR float64, scoreThreshold int, rsiBuyLow, rsiBuyHigh, rsiSellLow, rsiSellHigh float64, rsiRequired bool, rsiMode string) (score int, direction string, reasons []string) {
	score = 0
	if in.LiquidityEvent != nil && in.LiquidityEvent.Strength > 0 {
		score += 3
		if in.Structure.Trend == "BULLISH" && in.LiquidityEvent.Type == "BUY_SIDE" {
			direction = "BUY"
		} else if in.Structure.Trend == "BEARISH" && in.LiquidityEvent.Type == "SELL_SIDE" {
			direction = "SELL"
		}
	}
	if direction == "" {
		if in.Structure.Trend == "BULLISH" && in.SweepBuy {
			direction = "BUY"
		} else if in.Structure.Trend == "BEARISH" && in.SweepSell {
			direction = "SELL"
		}
	}
	if (direction == "BUY" && in.Structure.Trend == "BULLISH") || (direction == "SELL" && in.Structure.Trend == "BEARISH") {
		score += 2
	} else if direction != "" {
		reasons = append(reasons, "structure not aligned")
	}
	if in.EquilibriumTouch {
		score += 2
	} else {
		reasons = append(reasons, "not at equilibrium")
	}
	if indicators.ATRValid(in.ATR, minATR, maxATR) {
		score += 1
	} else {
		reasons = append(reasons, "ATR invalid")
	}
	if in.ATR >= 12 && in.ATR <= 18 {
		score += 1
	}
	if in.ATR >= 8 && in.ATR < 12 {
		score -= 1
	}
	// RSI: GATE / BOOST / PENALTY (same as composer)
	if rsiBuyLow <= 0 {
		rsiBuyLow, rsiBuyHigh = 38, 42
	}
	if rsiSellLow <= 0 {
		rsiSellLow, rsiSellHigh = 58, 62
	}
	mode := rsiMode
	if mode == "" {
		if rsiRequired {
			mode = "GATE"
		} else {
			mode = "BOOST"
		}
	}
	switch mode {
	case "GATE":
		if direction == "BUY" && !indicators.InBuyZone(in.RSI, rsiBuyLow, rsiBuyHigh) {
			reasons = append(reasons, fmt.Sprintf("RSI %.2f not in buy zone [%.0f-%.0f]", in.RSI, rsiBuyLow, rsiBuyHigh))
		}
		if direction == "SELL" && !indicators.InSellZone(in.RSI, rsiSellLow, rsiSellHigh) {
			reasons = append(reasons, fmt.Sprintf("RSI %.2f not in sell zone [%.0f-%.0f]", in.RSI, rsiSellLow, rsiSellHigh))
		}
	case "BOOST":
		if direction == "BUY" && indicators.InBuyZone(in.RSI, rsiBuyLow, rsiBuyHigh) {
			score += 1
		}
		if direction == "SELL" && indicators.InSellZone(in.RSI, rsiSellLow, rsiSellHigh) {
			score += 1
		}
	case "PENALTY":
		if direction == "BUY" {
			if indicators.InBuyZone(in.RSI, rsiBuyLow, rsiBuyHigh) {
				score += 2
			} else if in.RSI >= rsiSellLow {
				score -= 1
			}
		}
		if direction == "SELL" {
			if indicators.InSellZone(in.RSI, rsiSellLow, rsiSellHigh) {
				score += 2
			} else if in.RSI <= rsiBuyHigh {
				score -= 1
			}
		}
	}
	if score < 0 {
		score = 0
	}
	if direction == "" {
		hasSweepOrLiquidity := (in.LiquidityEvent != nil && in.LiquidityEvent.Strength > 0) || in.SweepBuy || in.SweepSell
		if hasSweepOrLiquidity {
			reasons = append(reasons, "direction not set: structure not aligned with sweep/liquidity event")
		} else {
			reasons = append(reasons, "no direction (no sweep/liquidity event)")
		}
	}
	if score < scoreThreshold {
		reasons = append(reasons, fmt.Sprintf("score %d < threshold %d", score, scoreThreshold))
	}
	return score, direction, reasons
}

// BuildComposerInput builds ComposerInput from candles, config, and structure components.
// sweepOpts configures sweep detection; if nil, defaults are used.
// useRSIWilder: if true, RSI uses Wilder smoothing (matches TradingView).
func BuildComposerInput(
	candles []models.Candle,
	lookback int,
	entryDelayCandles int,
	rsiPeriod, atrPeriod int,
	sweepOpts *SweepOptions,
	useRSIWilder bool,
) ComposerInput {
	swings := structure.DetectSwings(candles, lookback)
	ms := structure.BuildStructure(candles, swings)
	ss := structure.StructureStateFromSwings(ms, candles, swings)
	fib, fibValid := CalculateFib(swings)
	atr := indicators.ATR(candles, atrPeriod)
	var rsi float64
	if useRSIWilder {
		rsi = indicators.RSIWilder(candles, rsiPeriod)
	} else {
		rsi = indicators.RSI(candles, rsiPeriod)
	}
	zones := LiquidityAnalyzer(candles, atr, lookback)
	sweepBuy, sweepSell, intentStrength, sweepDiag := IntentDetector(candles, zones, atr, lookback, sweepOpts)
	// LiquidityEvent is the canonical representation of a sweep when detection occurs; BuildComposerInput
	// always sets it when sweepBuy || sweepSell, so state machine and composer stay aligned.
	var liqEvent *models.LiquidityEvent
	if sweepBuy || sweepSell {
		liqEvent = LiquidityEventFromIntent(sweepBuy, sweepSell, intentStrength)
	}
	eqTouch := EquilibriumTouch(candles, swings, 0.005)
	impulseIdx := 0
	for i := len(candles) - 1; i >= 1 && i >= len(candles)-20; i-- {
		body := candles[i].Close - candles[i].Open
		if body < 0 {
			body = -body
		}
		if atr > 0 && body > atr*0.8 {
			impulseIdx = i
			break
		}
	}
	return ComposerInput{
		Candles:          candles,
		Swings:           swings,
		Structure:        ms,
		StructureState:   ss,
		Fib:              fib,
		FibValid:         fibValid,
		LiquidityZones:   zones,
		LiquidityEvent:   liqEvent,
		RSI:              rsi,
		ATR:              atr,
		EquilibriumTouch: eqTouch,
		SweepBuy:         sweepBuy,
		SweepSell:        sweepSell,
		IntentStrength:   intentStrength,
		EntryDelayCandles: entryDelayCandles,
		ImpulseCandleIndex: impulseIdx,
		SweepDiagnostics:  sweepDiag,
	}
}
