package backtest

import (
	"aurumflow/internal/core"
	"aurumflow/internal/journal"
	"aurumflow/internal/marketstate"
	"aurumflow/internal/strategy"
	"aurumflow/pkg/models"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds backtest parameters (same as live config subset).
type Config struct {
	RSIPeriod           int
	ATRPeriod           int
	MinATR              float64
	MaxATR              float64
	SwingLookback       int
	EntryDelayCandles   int
	ScoreThreshold      int
	RiskPerTrade        float64
	RSIRequired         bool
	RSIMode              string // GATE, BOOST, PENALTY
	RSIBuyLow           float64
	RSIBuyHigh          float64
	RSISellLow          float64
	RSISellHigh         float64
	SweepCloseTolerance float64
	SweepMinPenetration float64
	SweepMinWickRatio   float64
	UseH1Filter         bool
	BlockH1Range        bool   // if true and UseH1Filter, skip when H1 trend is RANGE
	Crypto              bool   // if true, allow H1 RANGE when finalScore >= ScoreThreshold + H1RangeExtraScore
	H1RangeExtraScore   int    // extra score required for H1 RANGE when Crypto=true; 0 treated as 1
	H1SwingLookback     int
	UseH4Filter         bool   // if true, BUY only when H4 BULLISH, SELL only when H4 BEARISH
	BlockH4Range        bool   // if true and UseH4Filter, skip when H4 trend is RANGE
	H4SwingLookback     int
	UseRSIWilder        bool
	ValuePerPoint       float64 // money per 1.0 price move per 1.0 size; 0 => 1.0
	// State machine and gating (same as live)
	MaxTrades                   int     // max open positions; 0 => 1
	StateTTLCandles             int     // TTL for state machine; 0 => 12
	InvalidateOnOppositeStructure bool  // invalidate setup when structure opposes
	BlockLondonNYOverlap        bool      // if true, skip bars during LONDON+NY overlap (default true when not set in main config)
	TradingSessions             []string  // allowed sessions: LONDON, NY, ASIA, or ALL; empty/nil => ALL
	SpreadPoints                float64   // cost per trade (half spread each side); applied to PnL
	SlippagePoints              float64 // slippage per trade; applied to fill
	// Exit management (simulated in backtest)
	UseBreakEven bool    // if true, effective SL becomes entry once price reaches BreakEvenR
	BreakEvenR   float64 // R level to trigger BE (e.g. 1.0)
	// Risk (parity with live Risk.ValidateSignal)
	DailyDrawdownLimit float64 // 0 = disabled; same as risk.daily_drawdown_limit
}

// Result holds backtest metrics.
type Result struct {
	TotalTrades        int
	Winners            int
	Losers             int
	Winrate            float64
	ProfitFactor       float64
	MaxDrawdown        float64
	MaxDrawdownPct     float64
	Expectancy         float64
	EquityCurve        []float64
	BarsSkippedSession int // bars skipped by trading_sessions (CanTrade false)
	BarsSkippedLondonNY int // bars skipped by block_london_ny_overlap
}

// maxProfitFactor caps PF when losers=0 to avoid overflow in reports/ranking.
const maxProfitFactor = 99.0

// ClosedTrade holds one resolved trade for CSV/metrics (segmentación).
type ClosedTrade struct {
	EntryTS    time.Time
	ExitTS     time.Time
	Direction  string
	Entry      float64
	SL         float64
	TP         float64
	ExitPrice  float64
	ExitReason string
	PnlR       float64
	PnlMoney   float64
	TrendH4    string
	TrendH1    string
	Session    string
	ATRBucket  string
	SweepType  string
}

// atrBucket returns a bucket label for ATR (segmentación, same as live journal).
func atrBucket(atr float64) string {
	switch {
	case atr < 8:
		return "0-8"
	case atr < 12:
		return "8-12"
	case atr < 18:
		return "12-18"
	default:
		return "18+"
	}
}

// resolveOutcome checks whether the candle hits SL and/or TP.
// Same-candle rule (conservative): if both are touched in one candle, SL is considered first.
// BUY: Low <= sl checked before High >= tp. SELL: High >= sl checked before Low <= tp.
func resolveOutcome(c models.Candle, direction string, entry, sl, tp float64) (hitSL, hitTP bool) {
	if direction == "BUY" {
		if c.Low <= sl {
			hitSL = true
			return
		}
		if c.High >= tp {
			hitTP = true
		}
		return
	}
	// SELL
	if c.High >= sl {
		hitSL = true
		return
	}
	if c.Low <= tp {
		hitTP = true
	}
	return
}

// Run runs the strategy pipeline on historical candles without API calls.
// Gating order (same as live): 1) state == READY 2) signal ok 3) H4 filter (regime) 4) H1 filter 5) openCount < MaxTrades 6) spread/slippage.
// Same-candle SL/TP: SL is considered first (conservative). Spread and slippage are applied to PnL.
// If journal != nil, writes position_closed per resolved trade (with segmentación). If tradesCSVPath != "", writes CSV of closed trades.
func Run(candlesM15 []models.Candle, candlesH1 []models.Candle, candlesH4 []models.Candle, cfg Config, initialBalance float64, jw journal.JournalWriter, tradesCSVPath string) ([]models.TradeSignal, Result) {
	var signals []models.TradeSignal
	res := Result{}
	if len(candlesM15) < 50 {
		return signals, res
	}
	balance := initialBalance
	peak := balance
	var grossProfit, grossLoss float64
	equity := []float64{initialBalance}
	var closedTrades []ClosedTrade
	dailyStartBalance := initialBalance
	var dailyReset time.Time
	var barsSkippedSession, barsSkippedLondonNY int

	sm := core.NewStateMachine()
	openCount := 0
	stateTTL := cfg.StateTTLCandles
	if stateTTL <= 0 {
		stateTTL = 12
	}
	maxTrades := cfg.MaxTrades
	if maxTrades <= 0 {
		maxTrades = 1
	}

	applyH1Filter := cfg.UseH1Filter && len(candlesH1) > 0
	applyH4Filter := cfg.UseH4Filter && len(candlesH4) > 0
	lookbackH1 := cfg.H1SwingLookback
	if lookbackH1 <= 0 {
		lookbackH1 = 5
	}
	var h1Idx int
	lookbackH4 := cfg.H4SwingLookback
	if lookbackH4 <= 0 {
		lookbackH4 = 5
	}
	var h4Idx int

	sweepOpts := &strategy.SweepOptions{
		CloseTolerance:  cfg.SweepCloseTolerance,
		MinPenetration:  cfg.SweepMinPenetration,
		MinWickRatio:    cfg.SweepMinWickRatio,
	}
	valuePerPoint := cfg.ValuePerPoint
	if valuePerPoint <= 0 {
		valuePerPoint = 1.0
	}

	for i := 50; i < len(candlesM15); i++ {
		barDay := candlesM15[i].Time.UTC().Truncate(24 * time.Hour)
		if barDay.After(dailyReset) {
			dailyReset = barDay
			dailyStartBalance = balance
		}
		slice := candlesM15[:i+1]
		lookback := cfg.SwingLookback
		if lookback <= 0 {
			lookback = 5
		}
		in := strategy.BuildComposerInput(slice, lookback, cfg.EntryDelayCandles, cfg.RSIPeriod, cfg.ATRPeriod, sweepOpts, cfg.UseRSIWilder)
		hasLiquidity := len(in.LiquidityZones) > 0
		hasSweep := strategy.HasSweep(in)
		atEquilibrium := in.EquilibriumTouch
		setupDirection := ""
		if in.LiquidityEvent != nil {
			if in.LiquidityEvent.Type == "BUY_SIDE" {
				setupDirection = "BUY"
			} else {
				setupDirection = "SELL"
			}
		} else if in.SweepBuy {
			setupDirection = "BUY"
		} else if in.SweepSell {
			setupDirection = "SELL"
		}
		lastCandleTime := time.Time{}
		if len(slice) > 0 {
			lastCandleTime = slice[len(slice)-1].Time
		}
		sm.Transition(hasLiquidity, hasSweep, atEquilibrium, false, core.TransitionParams{
			LastCandleTime:              lastCandleTime,
			StructureTrend:              in.Structure.Trend,
			SetupDirection:              setupDirection,
			StateTTLCandles:             stateTTL,
			InvalidateOnOppositeStructure: cfg.InvalidateOnOppositeStructure,
		})
		// Advance H1 and H4 indices for this bar so trendH1/trendH4 are available for journal and filter
		if applyH1Filter {
			for h1Idx < len(candlesH1) && !candlesH1[h1Idx].Time.After(candlesM15[i].Time) {
				h1Idx++
			}
		}
		if len(candlesH4) > 0 {
			for h4Idx < len(candlesH4) && !candlesH4[h4Idx].Time.After(candlesM15[i].Time) {
				h4Idx++
			}
		}
		h1Slice := candlesH1[:h1Idx]
		trendH1 := ""
		if applyH1Filter && len(h1Slice) >= 10 {
			trendH1 = marketstate.TrendFromCandles(h1Slice, lookbackH1)
		}
		h4Slice := candlesH4[:h4Idx]
		trendH4 := ""
		if applyH4Filter && len(h4Slice) >= 10 {
			trendH4 = marketstate.TrendFromCandles(h4Slice, lookbackH4)
		}
		session := strategy.GetSessionInfo(candlesM15[i].Time)
		if len(cfg.TradingSessions) > 0 && !strategy.CanTrade(candlesM15[i].Time, cfg.TradingSessions) {
			barsSkippedSession++
			equity = append(equity, balance)
			continue
		}
		if session == strategy.SessionLondonNY && cfg.BlockLondonNYOverlap {
			barsSkippedLondonNY++
			equity = append(equity, balance)
			continue
		}
		// P0-C: setup_evaluated when we have a setup candidate (hasLiquidity || hasSweep)
		if jw != nil && (hasLiquidity || hasSweep) {
			diagScore, diagDir, diagReasons := strategy.SignalDiagnostics(in, cfg.MinATR, cfg.MaxATR, cfg.ScoreThreshold, cfg.RSIBuyLow, cfg.RSIBuyHigh, cfg.RSISellLow, cfg.RSISellHigh, cfg.RSIRequired, cfg.RSIMode)
			signalOK := diagScore >= cfg.ScoreThreshold && diagDir != ""
			_ = jw.WriteSetupEvaluated(journal.SetupEvaluated{
				Ts:            time.Now().UTC().Format(time.RFC3339),
				Event:         journal.EventSetupEvaluated,
				Session:       session,
				State:         string(sm.State()),
				TrendH4:       trendH4,
				TrendH1:       trendH1,
				TrendM15:      in.Structure.Trend,
				ATR:           in.ATR,
				RSI:           in.RSI,
				ATRBucket:     atrBucket(in.ATR),
				HasLiquidity:  hasLiquidity,
				HasSweep:      hasSweep,
				AtEquilibrium: atEquilibrium,
				Score:         diagScore,
				Threshold:     cfg.ScoreThreshold,
				SignalOK:      signalOK,
				SignalDirection: diagDir,
				RejectReasons: diagReasons,
			})
		}
		if sm.State() != core.StateReady {
			equity = append(equity, balance)
			continue
		}
		signal, ok := strategy.SignalComposerWithZones(in, cfg.MinATR, cfg.MaxATR, cfg.ScoreThreshold, cfg.RSIBuyLow, cfg.RSIBuyHigh, cfg.RSISellLow, cfg.RSISellHigh, cfg.RSIRequired, cfg.RSIMode)
		if !ok || signal == nil {
			equity = append(equity, balance)
			continue
		}
		// H4 and session as score (no reject): finalScore = signal.Score + h4Score + sessionScore
		h4Score := 0
		if trendH4 != "" {
			if (signal.Direction == "BUY" && trendH4 == "BULLISH") || (signal.Direction == "SELL" && trendH4 == "BEARISH") {
				h4Score = 2
			} else if trendH4 == "RANGE" {
				h4Score = -1
			} else {
				h4Score = -2
			}
		}
		sessionScore := 0
		if strings.Contains(session, strategy.SessionNY) {
			sessionScore = 1
		}
		if strings.Contains(session, strategy.SessionLondon) {
			sessionScore--
		}
		finalScore := signal.Score + h4Score + sessionScore
		if finalScore < cfg.ScoreThreshold {
			if jw != nil {
				_ = jw.WriteSignalRejected(journal.SignalRejected{
					Ts:           time.Now().UTC().Format(time.RFC3339),
					Event:        journal.EventSignalRejected,
					State:        string(sm.State()),
					Session:      session,
					RejectReason: "score_after_context",
					Direction:    signal.Direction,
					Entry:        signal.Entry,
					Score:        signal.Score,
				})
			}
			equity = append(equity, balance)
			continue
		}
		if applyH1Filter && len(h1Slice) >= 10 {
			if cfg.BlockH1Range && trendH1 == "RANGE" {
				if jw != nil {
					_ = jw.WriteSignalRejected(journal.SignalRejected{
						Ts:           time.Now().UTC().Format(time.RFC3339),
						Event:        journal.EventSignalRejected,
						State:        string(sm.State()),
						Session:      session,
						RejectReason: journal.RejectH1Range,
						Direction:    signal.Direction,
						Entry:        signal.Entry,
						Score:        signal.Score,
					})
				}
				equity = append(equity, balance)
				continue
			}
			allowed := (signal.Direction == "BUY" && trendH1 == "BULLISH") || (signal.Direction == "SELL" && trendH1 == "BEARISH")
			if !allowed && trendH1 == "RANGE" && cfg.Crypto {
				extra := cfg.H1RangeExtraScore
				if extra <= 0 {
					extra = 1
				}
				minScore := cfg.ScoreThreshold + extra
				if finalScore >= minScore {
					allowed = true
				} else {
					if jw != nil {
						_ = jw.WriteSignalRejected(journal.SignalRejected{
							Ts:           time.Now().UTC().Format(time.RFC3339),
							Event:        journal.EventSignalRejected,
							State:        string(sm.State()),
							Session:      session,
							RejectReason: journal.RejectH1RangeCryptoScore,
							Direction:    signal.Direction,
							Entry:        signal.Entry,
							Score:        signal.Score,
						})
					}
					equity = append(equity, balance)
					continue
				}
			}
			if !allowed {
				if jw != nil {
					_ = jw.WriteSignalRejected(journal.SignalRejected{
						Ts:           time.Now().UTC().Format(time.RFC3339),
						Event:        journal.EventSignalRejected,
						State:        string(sm.State()),
						Session:      session,
						RejectReason: journal.RejectH1Filter,
						Direction:    signal.Direction,
						Entry:        signal.Entry,
						Score:        signal.Score,
					})
				}
				equity = append(equity, balance)
				continue
			}
		}
		if cfg.DailyDrawdownLimit > 0 && dailyStartBalance > 0 {
			ddPct := (dailyStartBalance - balance) / dailyStartBalance * 100
			if ddPct >= cfg.DailyDrawdownLimit {
				if jw != nil {
					_ = jw.WriteSignalRejected(journal.SignalRejected{
						Ts:           time.Now().UTC().Format(time.RFC3339),
						Event:        journal.EventSignalRejected,
						State:        string(sm.State()),
						Session:      session,
						RejectReason: journal.RejectRiskReject,
						Direction:    signal.Direction,
						Entry:        signal.Entry,
						Score:        signal.Score,
					})
				}
				equity = append(equity, balance)
				continue
			}
		}
		if openCount >= maxTrades {
			if jw != nil {
				_ = jw.WriteSignalRejected(journal.SignalRejected{
					Ts:           time.Now().UTC().Format(time.RFC3339),
					Event:        journal.EventSignalRejected,
					State:        string(sm.State()),
					Session:      session,
					RejectReason: journal.RejectMaxTrades,
					Direction:    signal.Direction,
					Entry:        signal.Entry,
					Score:        signal.Score,
				})
			}
			equity = append(equity, balance)
			continue
		}
		// Segmentación for journal/CSV (reuse session, trendH4, trendH1 from above)
		entrySession := session
		entryTrendH4 := trendH4
		entryTrendH1 := trendH1
		entryATRBucket := atrBucket(in.ATR)
		entrySweepType := setupDirection
		entryTs := candlesM15[i].Time

		// P0-C: signal_generated, order_attempt, order_result (simulated)
		if jw != nil {
			stopDist := signal.Entry - signal.StopLoss
			if stopDist < 0 {
				stopDist = -stopDist
			}
			if stopDist == 0 {
				stopDist = 1
			}
			tpDist := signal.TakeProfit - signal.Entry
			if tpDist < 0 {
				tpDist = -tpDist
			}
			rr := 0.0
			if stopDist > 0 {
				rr = tpDist / stopDist
			}
			liqEvent := in.LiquidityEvent != nil && in.LiquidityEvent.Strength > 0
			sweepStr := 0.0
			if in.LiquidityEvent != nil {
				sweepStr = in.LiquidityEvent.Strength
			}
			_ = jw.WriteSignalGenerated(journal.SignalGenerated{
				Ts:               time.Now().UTC().Format(time.RFC3339),
				Event:            journal.EventSignalGenerated,
				State:            string(sm.State()),
				Session:          session,
				TrendH4:          trendH4,
				TrendH1:          trendH1,
				Direction:        signal.Direction,
				Entry:            signal.Entry,
				SL:               signal.StopLoss,
				TP:               signal.TakeProfit,
				StopDist:         stopDist,
				RR:               rr,
				Score:            signal.Score,
				Confidence:       signal.Confidence,
				ATR:              in.ATR,
				RSI:              in.RSI,
				EquilibriumTouch: in.EquilibriumTouch,
				SweepType:        entrySweepType,
				SweepStrength:    sweepStr,
				LiquidityEvent:   liqEvent,
			})
			riskAmt := balance * (cfg.RiskPerTrade / 100)
			size := riskAmt / (stopDist * valuePerPoint)
			if size < 0.1 {
				size = 0.1
			}
			fillEntry := signal.Entry
			if signal.Direction == "BUY" {
				fillEntry += cfg.SlippagePoints
			} else {
				fillEntry -= cfg.SlippagePoints
			}
			_ = jw.WriteOrderAttempt(journal.OrderAttempt{
				Ts:        time.Now().UTC().Format(time.RFC3339),
				Event:     journal.EventOrderAttempt,
				Direction: signal.Direction,
				Size:     size,
				Entry:    signal.Entry,
				SL:       signal.StopLoss,
				TP:       signal.TakeProfit,
			})
			_ = jw.WriteOrderResult(journal.OrderResult{
				Ts:        time.Now().UTC().Format(time.RFC3339),
				Event:     journal.EventOrderResult,
				Status:    "FILLED",
				FillPrice: fillEntry,
				Slippage:  cfg.SlippagePoints,
			})
		}

		signals = append(signals, *signal)
		openCount++
		entry := signal.Entry
		sl := signal.StopLoss
		tp := signal.TakeProfit
		stopDist := entry - sl
		if stopDist < 0 {
			stopDist = -stopDist
		}
		if stopDist == 0 {
			stopDist = 1
		}
		riskAmount := balance * (cfg.RiskPerTrade / 100)
		size := riskAmount / (stopDist * valuePerPoint)
		if size < 0.1 {
			size = 0.1
		}
		slippage := cfg.SlippagePoints
		fillEntry := entry
		if signal.Direction == "BUY" {
			fillEntry += slippage
		} else {
			fillEntry -= slippage
		}
		effectiveSL := sl
		beLevel := fillEntry + stopDist*cfg.BreakEvenR
		if signal.Direction == "SELL" {
			beLevel = fillEntry - stopDist*cfg.BreakEvenR
		}
		outcome := 0.0
		hitSL := false
		hitTP := false
		var exitCandle *models.Candle
		for j := i + 1; j < len(candlesM15) && j < i+200; j++ {
			c := candlesM15[j]
			if cfg.UseBreakEven && cfg.BreakEvenR > 0 {
				if signal.Direction == "BUY" && c.High >= beLevel {
					effectiveSL = fillEntry
				} else if signal.Direction == "SELL" && c.Low <= beLevel {
					effectiveSL = fillEntry
				}
			}
			hitSL, hitTP = resolveOutcome(c, signal.Direction, fillEntry, effectiveSL, tp)
			if hitSL || hitTP {
				exitCandle = &candlesM15[j]
				break
			}
		}
		if hitTP && !hitSL {
			outcome = (tp - fillEntry) * size * valuePerPoint
			if signal.Direction == "SELL" {
				outcome = (fillEntry - tp) * size * valuePerPoint
			}
			grossProfit += outcome
			res.Winners++
		} else if hitSL {
			outcome = (effectiveSL - fillEntry) * size * valuePerPoint
			if signal.Direction == "SELL" {
				outcome = (fillEntry - effectiveSL) * size * valuePerPoint
			}
			grossLoss += math.Abs(outcome)
			res.Losers++
		}
		spreadCost := size * cfg.SpreadPoints * valuePerPoint
		outcome -= spreadCost
		balance += outcome
		openCount-- // trade resolved (or no hit within lookback; treat as closed for simplicity)

		// Write journal and collect for CSV when we have a resolved exit (hit SL or TP)
		if exitCandle != nil && (hitSL || hitTP) {
			exitReason := "TP"
			exitPrice := tp
			if hitSL {
				exitReason = "SL"
				exitPrice = effectiveSL
			}
			pnlR := 0.0
			if riskAmount > 0 {
				pnlR = outcome / riskAmount
			}
			entryTSStr := entryTs.UTC().Format(time.RFC3339)
			exitTSStr := exitCandle.Time.UTC().Format(time.RFC3339)
			if jw != nil {
				_ = jw.WritePositionClosed(journal.PositionClosed{
					Ts:         time.Now().UTC().Format(time.RFC3339),
					Event:      journal.EventPositionClosed,
					Epic:       "",
					Direction:  signal.Direction,
					Entry:      fillEntry,
					SL:         effectiveSL,
					TP:         tp,
					ExitPrice:  exitPrice,
					ExitReason: exitReason,
					PnlR:       pnlR,
					PnlMoney:   outcome,
					TrendH4:    entryTrendH4,
					TrendH1:    entryTrendH1,
					Session:    entrySession,
					ATRBucket:  entryATRBucket,
					SweepType:  entrySweepType,
					EntryTS:    entryTSStr,
					ExitTS:     exitTSStr,
				})
			}
			if tradesCSVPath != "" {
				closedTrades = append(closedTrades, ClosedTrade{
					EntryTS:    entryTs,
					ExitTS:     exitCandle.Time,
					Direction:  signal.Direction,
					Entry:      fillEntry,
					SL:         effectiveSL,
					TP:         tp,
					ExitPrice:  exitPrice,
					ExitReason: exitReason,
					PnlR:       pnlR,
					PnlMoney:   outcome,
					TrendH4:    entryTrendH4,
					TrendH1:    entryTrendH1,
					Session:    entrySession,
					ATRBucket:  entryATRBucket,
					SweepType:  entrySweepType,
				})
			}
		}
		if balance > peak {
			peak = balance
		}
		dd := (peak - balance) / peak * 100
		if dd > res.MaxDrawdownPct {
			res.MaxDrawdownPct = dd
			res.MaxDrawdown = peak - balance
		}
		equity = append(equity, balance)
	}
	res.TotalTrades = len(signals)
	if res.TotalTrades > 0 {
		res.Winrate = float64(res.Winners) / float64(res.TotalTrades) * 100
	}
	if grossLoss > 0 {
		res.ProfitFactor = grossProfit / grossLoss
		if res.ProfitFactor > maxProfitFactor {
			res.ProfitFactor = maxProfitFactor
		}
	} else if grossProfit > 0 {
		res.ProfitFactor = maxProfitFactor
	}
	if res.TotalTrades > 0 {
		res.Expectancy = (grossProfit - grossLoss) / float64(res.TotalTrades)
	}
	res.EquityCurve = equity
	res.BarsSkippedSession = barsSkippedSession
	res.BarsSkippedLondonNY = barsSkippedLondonNY
	if tradesCSVPath != "" && len(closedTrades) > 0 {
		if err := writeTradesCSV(tradesCSVPath, closedTrades); err != nil {
			// best-effort: don't fail Run, caller may log
			_ = err
		}
	}
	return signals, res
}

// writeTradesCSV writes closed trades to a CSV file for heatmap/segmentación analysis.
func writeTradesCSV(path string, trades []ClosedTrade) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	header := []string{"entry_ts", "exit_ts", "direction", "entry", "sl", "tp", "exit_price", "exit_reason", "pnl_r", "pnl_money", "trend_h4", "trend_h1", "session", "atr_bucket", "sweep_type"}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, t := range trades {
		row := []string{
			t.EntryTS.UTC().Format(time.RFC3339),
			t.ExitTS.UTC().Format(time.RFC3339),
			t.Direction,
			strconv.FormatFloat(t.Entry, 'f', -1, 64),
			strconv.FormatFloat(t.SL, 'f', -1, 64),
			strconv.FormatFloat(t.TP, 'f', -1, 64),
			strconv.FormatFloat(t.ExitPrice, 'f', -1, 64),
			t.ExitReason,
			strconv.FormatFloat(t.PnlR, 'f', -1, 64),
			strconv.FormatFloat(t.PnlMoney, 'f', -1, 64),
			t.TrendH4,
			t.TrendH1,
			t.Session,
			t.ATRBucket,
			t.SweepType,
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// CandleRow is used for JSON file format (time as RFC3339 or Unix ms).
type CandleRow struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// CandleFile is the JSON file format with metadata (epic, resolution, source, generated_at).
// LoadCandlesFromFile accepts both CandleFile (object with "candles") and legacy []CandleRow (array only).
type CandleFile struct {
	Epic        string      `json:"epic,omitempty"`
	Resolution  string      `json:"resolution,omitempty"`
	Source      string      `json:"source,omitempty"`
	GeneratedAt string      `json:"generated_at,omitempty"`
	Candles     []CandleRow `json:"candles"`
}

// SaveCandlesToFile writes candles to path as JSON with metadata. Creates parent directory if needed.
func SaveCandlesToFile(path string, epic, resolution, source string, candles []models.Candle) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create backtest dir: %w", err)
	}
	rows := make([]CandleRow, 0, len(candles))
	for _, c := range candles {
		rows = append(rows, CandleRow{
			Time:   c.Time.UTC().Format(time.RFC3339),
			Open:   c.Open,
			High:   c.High,
			Low:    c.Low,
			Close:  c.Close,
			Volume: c.Volume,
		})
	}
	out := CandleFile{
		Epic:        epic,
		Resolution:  resolution,
		Source:      source,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Candles:     rows,
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal candles: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write candles: %w", err)
	}
	return nil
}

// LoadCandlesFromFile loads candles from JSON (array of {time, open, high, low, close, volume}) or CSV (header: time,open,high,low,close,volume).
func LoadCandlesFromFile(path string) ([]models.Candle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".json") {
		return loadCandlesJSON(data)
	}
	return loadCandlesCSV(data)
}

func loadCandlesJSON(data []byte) ([]models.Candle, error) {
	// Try metadata format first (CandleFile with "candles" array).
	var cf CandleFile
	if err := json.Unmarshal(data, &cf); err == nil && len(cf.Candles) > 0 {
		return candleRowsToCandles(cf.Candles), nil
	}
	// Fall back to legacy format (array of CandleRow only).
	var rows []CandleRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	return candleRowsToCandles(rows), nil
}

func candleRowsToCandles(rows []CandleRow) []models.Candle {
	candles := make([]models.Candle, 0, len(rows))
	for _, r := range rows {
		t, err := parseTime(r.Time)
		if err != nil {
			continue
		}
		candles = append(candles, models.Candle{
			Time:   t,
			Open:   r.Open,
			High:   r.High,
			Low:    r.Low,
			Close:  r.Close,
			Volume: r.Volume,
		})
	}
	return candles
}

func loadCandlesCSV(data []byte) ([]models.Candle, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, nil
	}
	header := records[0]
	timeIdx, openIdx, highIdx, lowIdx, closeIdx, volIdx := -1, -1, -1, -1, -1, -1
	for i, h := range header {
		h = strings.ToLower(strings.TrimSpace(h))
		switch h {
		case "time", "timestamp", "date":
			timeIdx = i
		case "open":
			openIdx = i
		case "high":
			highIdx = i
		case "low":
			lowIdx = i
		case "close":
			closeIdx = i
		case "volume", "vol":
			volIdx = i
		}
	}
	if timeIdx < 0 || openIdx < 0 || closeIdx < 0 {
		return nil, nil
	}
	if highIdx < 0 {
		highIdx = openIdx
	}
	if lowIdx < 0 {
		lowIdx = openIdx
	}
	candles := make([]models.Candle, 0, len(records)-1)
	for _, rec := range records[1:] {
		if len(rec) <= max(timeIdx, openIdx, highIdx, lowIdx, closeIdx) {
			continue
		}
		t, err := parseTime(strings.TrimSpace(rec[timeIdx]))
		if err != nil {
			continue
		}
		o, _ := strconv.ParseFloat(strings.TrimSpace(rec[openIdx]), 64)
		hi, _ := strconv.ParseFloat(strings.TrimSpace(rec[highIdx]), 64)
		lo, _ := strconv.ParseFloat(strings.TrimSpace(rec[lowIdx]), 64)
		c, _ := strconv.ParseFloat(strings.TrimSpace(rec[closeIdx]), 64)
		vol := 0.0
		if volIdx >= 0 && volIdx < len(rec) {
			vol, _ = strconv.ParseFloat(strings.TrimSpace(rec[volIdx]), 64)
		}
		candles = append(candles, models.Candle{Time: t, Open: o, High: hi, Low: lo, Close: c, Volume: vol})
	}
	return candles, nil
}

func max(a, b, c, d, e int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	if d > m {
		m = d
	}
	if e > m {
		m = e
	}
	return m
}

func parseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if unix, err := strconv.ParseInt(s, 10, 64); err == nil {
		if unix > 1e12 {
			return time.UnixMilli(unix), nil
		}
		return time.Unix(unix, 0), nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", s)
}
