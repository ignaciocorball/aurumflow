package core

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"aurumflow/config"
	"aurumflow/internal/execution"
	"aurumflow/internal/journal"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/marketstate"
	"aurumflow/internal/money"
	"aurumflow/internal/notifications"
	"aurumflow/internal/risk"
	"aurumflow/internal/strategy"
	"aurumflow/internal/stratrade"
	"aurumflow/internal/telemetry"
	"aurumflow/pkg/models"
)

const (
	marketStatusCacheTTL = 5 * time.Minute
	idleLogInterval      = 5 * time.Minute
	idleProgressBarWidth = 24
)

// Loop runs the main event loop: fetch candles, run structure -> strategy -> composer -> state machine -> risk -> execution.
type Loop struct {
	Config              *config.Config
	Client              *market.Client
	Epic                string
	Risk                *risk.Manager
	Exec                *execution.Executor
	State               *StateMachine
	Journal             journal.JournalWriter  // optional; if set, writes setup/signal/order/position events
	NotifEmitter        *notifications.Emitter // optional; if set and active, sends Pushover notifications
	Interval            time.Duration
	PingEvery           time.Duration
	lastPing            time.Time
	ReloginFn           func() error
	LiveConfirmRequired bool // deprecated in P1; LIVE is fail-closed at startup
	Provider            execution.ExecutionProvider
	Kill                *killswitch.Switch
	Spec                market.InstrumentSpec
	Money               money.MonetaryInstrumentSpec
	RequireRuntimeMoney bool
	UnknownPositions    int
	Decision            *strategy.DecisionContext
	LastSignalText      string
	// OnLegacySignal is SHADOW observation only. It must not open, close, filter, or veto.
	OnLegacySignal      func(sig *models.TradeSignal, now time.Time)
	AccountID           string
	tickCount           int
	lastStatusLog       time.Time
	lastIdleLog         time.Time // when we last logged an idle line (OFF_HOURS or MARKET_CLOSED)
	// lastOpened* used for PositionClosed best-effort when openCount drops to 0
	lastOpenedDealRef string
	lastOpenedSignal  *models.TradeSignal
	lastOpenedAt      time.Time
	// reopenWarmupCandlesRemaining: after market closed/data frozen, wait this many ticks before evaluating signals again
	reopenWarmupCandlesRemaining int
	// lastKnownMarketStatus / lastKnownMarketStatusAt: cache for market status when API fails (e.g. 500)
	lastKnownMarketStatus   string
	lastKnownMarketStatusAt time.Time
	// prevM15LastTime: last M15 candle time seen when decrementing warmup (warmup decrements only on new M15)
	prevM15LastTime time.Time
	// lastDecision for HEARTBEAT: e.g. "idle", "in_trade", "rejected:H1_RANGE", "waiting:warmup"
	lastDecision  string
	lastSweepType string // for SWEEP_CONFIRMED typeChanged detection
	// InstanceID and RunID for notification payloads (tracing; set from main)
	InstanceID string
	RunID      string
	// lastHeartbeatBoundary: last aligned boundary when we emitted heartbeat (when heartbeat_interval_minutes + align_top are set)
	lastHeartbeatBoundary time.Time
	// Telemetry: optional live publisher (e.g. Firebase RTDB); set from main when telemetry.firebase enabled
	Telemetry telemetry.LivePublisher
	// BotTag: directory name of bot config (e.g. "eth-offensive-alpha"); empty if config is in config/ or root
	BotTag           string
	BotTagNormalized string
	Forensic         *stratrade.Recorder
	Week             *stratrade.WeekState
	ObserveIntel     func() stratrade.IntelligenceContext
	GitCommit        string
	StrategyVersion  string
	LastScanAt       time.Time
	LastSignalAt     time.Time
	LastSignalID     string
}

// NewLoop creates the event loop with wired dependencies.
func NewLoop(cfg *config.Config, client *market.Client, epic string, balance float64, openCount int) *Loop {
	rm := risk.NewManager(cfg.Risk.RiskPerTrade, cfg.Risk.MaxTrades, cfg.Risk.DailyDrawdownLimit, balance)
	rm.SetOpenCount(openCount)
	exec := execution.NewExecutor(client, epic, 0.1, 0.1)
	return &Loop{
		Config:    cfg,
		Client:    client,
		Epic:      epic,
		Risk:      rm,
		Exec:      exec,
		Provider:  exec,
		State:     NewStateMachine(),
		Interval:  60 * time.Second,
		PingEvery: 5 * time.Minute,
	}
}

// Run fetches candles periodically, runs the pipeline, and may send an order when READY and risk allows.
func (l *Loop) Run(ctx context.Context) {
	logger.Info("event loop started: epic=%s interval=%v maxTrades=%d riskPerTrade=%.2f%%",
		l.Epic, l.Interval, l.Config.Risk.MaxTrades, l.Config.Risk.RiskPerTrade)
	if l.Telemetry != nil {
		meta := telemetry.MetaPayload{
			RunID: l.RunID, Epic: l.Epic,
			StartedAt:     time.Now().UTC().Format(time.RFC3339),
			SchemaVersion: 1, AnalyticsVersion: 1,
		}
		if l.BotTag != "" {
			meta.BotTag = l.BotTag
			meta.BotTagNormalized = l.BotTagNormalized
		}
		if err := l.Telemetry.PublishMeta(ctx, l.InstanceID, l.RunID, meta); err != nil {
			logger.Warn("telemetry publish meta failed: %v", err)
		}
	}
	ticker := time.NewTicker(l.Interval)
	defer ticker.Stop()
	// First tick immediately
	l.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.Info("loop stopped")
			return
		case <-ticker.C:
			l.tick(ctx)
		}
	}
}

func (l *Loop) tick(ctx context.Context) {
	l.tickCount++
	now := time.Now().UTC()
	currentState := l.State.State()
	l.lastDecision = string(currentState)
	if currentState == StateInTrade {
		l.lastDecision = "in_trade"
	}

	// Status log / heartbeat: aligned (if configured) or every 5 ticks / 10 min
	doHeartbeat := false
	if l.Config.Notifications != nil && l.Config.Notifications.HeartbeatIntervalMinutes > 0 {
		intervalMin := l.Config.Notifications.HeartbeatIntervalMinutes
		alignTop := l.Config.Notifications.HeartbeatAlignTop
		nextBoundary := nextHeartbeatBoundary(now, intervalMin, alignTop)
		if !nextBoundary.IsZero() && !now.Before(nextBoundary) && (l.lastHeartbeatBoundary.IsZero() || l.lastHeartbeatBoundary.Before(nextBoundary)) {
			doHeartbeat = true
			l.lastHeartbeatBoundary = nextBoundary
		}
	} else {
		doHeartbeat = time.Since(l.lastStatusLog) > 10*time.Minute || l.tickCount%5 == 0
	}
	if doHeartbeat {
		// Populate market status before HEARTBEAT so notification shows TRADEABLE/CLOSED instead of UNKNOWN
		if l.lastKnownMarketStatus == "" && l.Config.DataQuality != nil {
			if details, err := l.Client.GetMarketDetails(ctx, l.Epic); err == nil && details != nil && details.Snapshot.MarketStatus != "" {
				l.lastKnownMarketStatus = details.Snapshot.MarketStatus
				l.lastKnownMarketStatusAt = now
			}
		}
		l.logStatusSummary(ctx, now)
		l.lastStatusLog = now
	}

	// Market Readiness Gate: skip signal evaluation when market is not TRADEABLE (data quality)
	if l.Config.DataQuality != nil {
		details, err := l.Client.GetMarketDetails(ctx, l.Epic)
		if err != nil {
			if strings.Contains(err.Error(), "401") && l.ReloginFn != nil {
				_ = l.ReloginFn()
			}
			cacheValid := l.lastKnownMarketStatus != "" && time.Since(l.lastKnownMarketStatusAt) <= marketStatusCacheTTL
			if cacheValid {
				if l.lastKnownMarketStatus != "TRADEABLE" {
					l.lastDecision = "waiting:MARKET_CLOSED"
					if l.lastIdleLog.IsZero() || time.Since(l.lastIdleLog) >= idleLogInterval {
						l.logIdleLine("MARKET_CLOSED", now)
					}
					return
				}
			} else {
				l.lastDecision = "waiting:MARKET_CLOSED"
				if l.lastIdleLog.IsZero() || time.Since(l.lastIdleLog) >= idleLogInterval {
					l.logIdleLine("MARKET_CLOSED", now)
				}
				return
			}
		} else {
			if details != nil && details.Snapshot.MarketStatus != "" {
				l.lastKnownMarketStatus = details.Snapshot.MarketStatus
				l.lastKnownMarketStatusAt = now
			}
			if details != nil && details.Snapshot.MarketStatus != "" && details.Snapshot.MarketStatus != "TRADEABLE" {
				l.lastDecision = "waiting:MARKET_CLOSED"
				l.reopenWarmupCandlesRemaining = l.Config.DataQuality.ReopenWarmupM15Candles
				if l.lastIdleLog.IsZero() || time.Since(l.lastIdleLog) >= idleLogInterval {
					l.logIdleLine("MARKET_CLOSED", now)
				}
				return
			}
		}
	}

	// Check trading session filter
	if !strategy.CanTrade(now, l.Config.Strategy.TradingSessions) {
		l.lastDecision = "waiting:OFF_HOURS"
		if l.lastIdleLog.IsZero() || time.Since(l.lastIdleLog) >= idleLogInterval {
			l.logIdleLine("OFF_HOURS", now)
		}
		return
	}
	l.LastScanAt = now
	logger.Info("scanning for signals... [state=%s tick=%d session=%s]", currentState, l.tickCount, strategy.GetSessionInfo(now))

	if time.Since(l.lastPing) > l.PingEvery {
		if err := l.Client.Ping(ctx); err != nil {
			if strings.Contains(err.Error(), "401") && l.ReloginFn != nil {
				if reloginErr := l.ReloginFn(); reloginErr != nil {
					logger.Error("relogin after ping 401: %v", reloginErr)
				} else {
					logger.Success("relogin successful after 401")
				}
			} else {
				logger.Warn("ping failed: %v", err)
			}
		} else {
			l.lastPing = now
		}
	}

	// Refresh balance of the selected account only (never accounts[0]).
	if ar, err := l.Client.GetAccounts(ctx); err == nil {
		acc, selErr := risk.SelectAccount(ar.Accounts, l.Config.API.AccountID, l.AccountID)
		if selErr != nil {
			logger.Warn("loop: account selection: %v", selErr)
		} else {
			l.Risk.UpdateBalance(acc.Balance.Balance)
			l.AccountID = acc.AccountID
		}
	}
	openCount := 0
	var positions []market.PositionItem
	if pos, err := l.Client.GetPositions(ctx); err != nil {
		if strings.Contains(err.Error(), "401") && l.ReloginFn != nil {
			_ = l.ReloginFn()
		}
	} else {
		// Filter positions by epic (only count positions for the instrument we're trading)
		var filteredPositions []market.PositionItem
		for _, p := range pos.Positions {
			epic := p.GetEpic()
			if epic == l.Epic {
				filteredPositions = append(filteredPositions, p)
			}
		}
		openCount = len(filteredPositions)
		positions = filteredPositions
		l.Risk.SetOpenCount(openCount)

		// Log if there are positions in other instruments
		if len(pos.Positions) > openCount {
			logger.Info("positions filtered: %d total positions, %d for epic %s (ignoring %d in other instruments)",
				len(pos.Positions), openCount, l.Epic, len(pos.Positions)-openCount)
		}

		if openCount > 0 && l.State.State() == StateInTrade {
			l.sampleStrategyHeartbeat(now, positions)
		}
		if l.State.State() == StateInTrade && openCount > 1 {
			l.haltForensic("unknown extra GOLD position")
		}
		// Transition IN_TRADE -> COOLDOWN when all positions closed
		if l.State.State() == StateInTrade && openCount == 0 {
			minutesInTrade := 0
			if !l.lastOpenedAt.IsZero() {
				minutesInTrade = int(time.Since(l.lastOpenedAt).Minutes())
			}
			if l.Journal != nil && l.lastOpenedSignal != nil {
				_ = l.Journal.WritePositionClosed(journal.PositionClosed{
					Epic:           l.Epic,
					DealID:         l.lastOpenedDealRef,
					Direction:      l.lastOpenedSignal.Direction,
					Entry:          l.lastOpenedSignal.Entry,
					SL:             l.lastOpenedSignal.StopLoss,
					TP:             l.lastOpenedSignal.TakeProfit,
					ExitPrice:      0,
					ExitReason:     "UNKNOWN",
					MinutesInTrade: minutesInTrade,
				})
			}
			if l.NotifEmitter != nil && l.lastOpenedSignal != nil {
				closedAt := time.Now().UTC()
				l.NotifEmitter.Emit(notifications.NotifEvent{
					Type: notifications.TypePositionClosed, Category: notifications.CategoryExecution, Severity: notifications.SeverityCritical,
					Instrument: l.Epic, Timestamp: closedAt,
					Payload: map[string]any{
						"exit_price":       0,
						"exit_reason":      "UNKNOWN",
						"pnl_pct":          0,
						"pnl_r":            0,
						"duration_minutes": minutesInTrade,
						"deal_ref":         l.lastOpenedDealRef,
						"event_id":         uuid.New().String(),
						"event_ts_utc":     closedAt.Format(time.RFC3339),
						"instance_id":      l.InstanceID,
						"run_id":           l.RunID,
					},
				})
			}
			l.finalizeStrategyTrade("UNKNOWN", 0, 0, 0)
			l.State.ToCooldown()
			logger.Success("all positions closed, state -> COOLDOWN")
			telemetry.PublishPositionsBestEffort(l.Telemetry, ctx, l.InstanceID, l.Risk.GetOpenCount(), l.Config.Risk.MaxTrades, l.lastOpenedDealRef)
		}
		// Log position details if in trade
		if openCount > 0 && l.State.State() == StateInTrade {
			logger.Info("position status: %d of %d open", openCount, l.Config.Risk.MaxTrades)
			for i, p := range positions {
				logger.Info("  position[%d]: %s size=%.2f entry=%.2f dealId=%s epic=%s",
					i+1, p.Position.Direction, p.Position.Size, p.Position.Level, p.Position.DealID, p.GetEpic())
			}
		}
	}

	from := now.Add(-24 * time.Hour)
	to := now
	max := 200

	candlesM15, err := l.Client.GetPrices(ctx, l.Epic, l.Config.Timeframes.M15, max, from, to)
	if err != nil {
		if strings.Contains(err.Error(), "401") && l.ReloginFn != nil {
			_ = l.ReloginFn()
		}
		logger.Warn("loop: get prices M15: %v", err)
		return
	}
	if len(candlesM15) < 15 {
		logger.Warn("loop: not enough M15 candles (%d)", len(candlesM15))
		return
	}
	lastM15 := candlesM15[len(candlesM15)-1].Time

	var trendH1 string
	var lastH1 time.Time
	if l.Config.Strategy.UseH1Filter {
		candlesH1, errH1 := l.Client.GetPrices(ctx, l.Epic, l.Config.Timeframes.H1, max, from, to)
		if errH1 != nil {
			if strings.Contains(errH1.Error(), "401") && l.ReloginFn != nil {
				_ = l.ReloginFn()
			}
			logger.Warn("loop: get prices H1: %v (continuing without H1 filter)", errH1)
		} else if len(candlesH1) < 10 {
			logger.Warn("loop: not enough H1 candles (%d), continuing without H1 filter", len(candlesH1))
		} else {
			lastH1 = candlesH1[len(candlesH1)-1].Time
			lookbackH1 := l.Config.Strategy.H1SwingLookback
			if lookbackH1 <= 0 {
				lookbackH1 = 5
			}
			trendH1 = marketstate.TrendFromCandles(candlesH1, lookbackH1)
		}
	}
	const minH4CandlesForContext = 60
	var lastH4 time.Time // below this, H4 context is insufficient for production
	var trendH4 string
	var h4Ready bool
	if l.Config.Strategy.UseH4Filter {
		fromH4 := now.Add(-14 * 24 * time.Hour) // 14 days for H4 warmup (60+ candles)
		maxH4 := 1000
		candlesH4, errH4 := l.Client.GetPrices(ctx, l.Epic, l.Config.Timeframes.H4, maxH4, fromH4, to)
		if errH4 != nil {
			if strings.Contains(errH4.Error(), "401") && l.ReloginFn != nil {
				_ = l.ReloginFn()
			}
			logger.Warn("loop: get prices H4: %v (continuing without H4 filter)", errH4)
		} else if len(candlesH4) < minH4CandlesForContext {
			logger.Warn("H4 readiness=false, h4Candles=%d, minRequired=%d, context_degraded=true; skipping signal evaluation (skip_reason=H4_NOT_READY)",
				len(candlesH4), minH4CandlesForContext)
			return
		} else {
			h4Ready = true
			lastH4 = candlesH4[len(candlesH4)-1].Time
			lookbackH4 := l.Config.Strategy.H4SwingLookback
			if lookbackH4 <= 0 {
				lookbackH4 = 5
			}
			trendH4 = marketstate.TrendFromCandles(candlesH4, lookbackH4)
		}
	} else {
		h4Ready = true
	}
	_ = h4Ready

	// M5 refiner: fetch M5 with short window for timing in READY only (degradation, no block)
	var candlesM5 []models.Candle
	var lastM5 time.Time
	m5Status := ""
	if l.Config.M5Refiner != nil && l.Config.M5Refiner.Enabled {
		fromM5 := now.Add(-10 * time.Hour)
		maxM5 := 300
		candlesM5, err = l.Client.GetPrices(ctx, l.Epic, l.Config.Timeframes.M5, maxM5, fromM5, to)
		if err != nil {
			if strings.Contains(err.Error(), "401") && l.ReloginFn != nil {
				_ = l.ReloginFn()
			}
			m5Status = "DEGRADED"
			if l.Config.M5Refiner.DegradeIfUnavailable {
				logger.Warn("m5_status=DEGRADED; get prices M5 failed: %v (proceeding M15-only)", err)
			}
		} else if len(candlesM5) < 10 {
			m5Status = "DEGRADED"
			if l.Config.M5Refiner.DegradeIfUnavailable {
				logger.Warn("m5_status=DEGRADED; not enough M5 candles (%d) (proceeding M15-only)", len(candlesM5))
			}
		} else {
			lastM5 = candlesM5[len(candlesM5)-1].Time
			maxAge := time.Duration(l.Config.M5Refiner.M5MaxAgeMinutes) * time.Minute
			coherenceDiff := time.Duration(l.Config.M5Refiner.M5CoherenceDiffMinutes) * time.Minute
			ageM5 := now.Sub(lastM5)
			diffM15M5 := lastM15.Sub(lastM5)
			if diffM15M5 < 0 {
				diffM15M5 = -diffM15M5
			}
			if ageM5 > maxAge || diffM15M5 > coherenceDiff {
				m5Status = "DEGRADED"
				if l.Config.M5Refiner.DegradeIfUnavailable {
					logger.Warn("m5_status=DEGRADED; M5 stale or misaligned (age=%.0fm diff_m15_m5=%.0fm) (proceeding M15-only)", ageM5.Minutes(), diffM15M5.Minutes())
				}
			} else {
				m5Status = "OK"
			}
		}
	}

	// Freshness check: skip signal evaluation if last candle is too old (only when market is TRADEABLE)
	if l.Config.DataQuality != nil {
		maxAgeM15 := time.Duration(l.Config.DataQuality.MaxAgeM15Minutes) * time.Minute
		maxAgeH1 := time.Duration(l.Config.DataQuality.MaxAgeH1Minutes) * time.Minute
		maxAgeH4 := time.Duration(l.Config.DataQuality.MaxAgeH4Minutes) * time.Minute
		ageM15 := now.Sub(lastM15)
		if ageM15 > maxAgeM15 {
			l.reopenWarmupCandlesRemaining = l.Config.DataQuality.ReopenWarmupM15Candles
			l.prevM15LastTime = lastM15
			l.lastDecision = "data_frozen:M15"
			if l.NotifEmitter != nil {
				l.NotifEmitter.Emit(notifications.NotifEvent{
					Type: notifications.TypeDataStale, Category: notifications.CategoryHealth, Severity: notifications.SeverityCritical,
					Instrument: l.Epic, Timestamp: now,
					Payload: map[string]any{
						"timeframe": "M15", "age_minutes": ageM15.Minutes(), "max_age_minutes": maxAgeM15.Minutes(),
						"last_ts": lastM15.Format("2006-01-02T15:04"), "block_on_data_frozen": l.Config.DataQuality.BlockOnDataFrozen,
						"prolonged": ageM15 > maxAgeM15*2,
					},
				})
			}
			logger.Warn("DATA_FROZEN: epic=%s tf=M15 last=%s age=%.0fm > maxAge=%.0fm; skipping signal evaluation (op_state=OP_DATA_FROZEN)",
				l.Epic, lastM15.Format("2006-01-02T15:04"), ageM15.Minutes(), maxAgeM15.Minutes())
			if l.Config.DataQuality.BlockOnDataFrozen {
				return
			}
		}
		if l.Config.Strategy.UseH1Filter && !lastH1.IsZero() {
			ageH1 := now.Sub(lastH1)
			if ageH1 > maxAgeH1 {
				l.reopenWarmupCandlesRemaining = l.Config.DataQuality.ReopenWarmupM15Candles
				l.prevM15LastTime = lastM15
				l.lastDecision = "data_frozen:H1"
				if l.NotifEmitter != nil {
					l.NotifEmitter.Emit(notifications.NotifEvent{
						Type: notifications.TypeDataStale, Category: notifications.CategoryHealth, Severity: notifications.SeverityCritical,
						Instrument: l.Epic, Timestamp: now,
						Payload: map[string]any{
							"timeframe": "H1", "age_minutes": ageH1.Minutes(), "max_age_minutes": maxAgeH1.Minutes(),
							"last_ts": lastH1.Format("2006-01-02T15:04"), "block_on_data_frozen": l.Config.DataQuality.BlockOnDataFrozen,
							"prolonged": ageH1 > maxAgeH1*2,
						},
					})
				}
				logger.Warn("DATA_FROZEN: epic=%s tf=H1 last=%s age=%.0fm > maxAge=%.0fm; skipping signal evaluation (op_state=OP_DATA_FROZEN)",
					l.Epic, lastH1.Format("2006-01-02T15:04"), ageH1.Minutes(), maxAgeH1.Minutes())
				if l.Config.DataQuality.BlockOnDataFrozen {
					return
				}
			}
		}
		if l.Config.Strategy.UseH4Filter && !lastH4.IsZero() {
			ageH4 := now.Sub(lastH4)
			if ageH4 > maxAgeH4 {
				l.reopenWarmupCandlesRemaining = l.Config.DataQuality.ReopenWarmupM15Candles
				l.prevM15LastTime = lastM15
				l.lastDecision = "data_frozen:H4"
				if l.NotifEmitter != nil {
					l.NotifEmitter.Emit(notifications.NotifEvent{
						Type: notifications.TypeDataStale, Category: notifications.CategoryHealth, Severity: notifications.SeverityCritical,
						Instrument: l.Epic, Timestamp: now,
						Payload: map[string]any{
							"timeframe": "H4", "age_minutes": ageH4.Minutes(), "max_age_minutes": maxAgeH4.Minutes(),
							"last_ts": lastH4.Format("2006-01-02T15:04"), "block_on_data_frozen": l.Config.DataQuality.BlockOnDataFrozen,
							"prolonged": ageH4 > maxAgeH4*2,
						},
					})
				}
				logger.Warn("DATA_FROZEN: epic=%s tf=H4 last=%s age=%.0fm > maxAge=%.0fm; skipping signal evaluation (op_state=OP_DATA_FROZEN)",
					l.Epic, lastH4.Format("2006-01-02T15:04"), ageH4.Minutes(), maxAgeH4.Minutes())
				if l.Config.DataQuality.BlockOnDataFrozen {
					return
				}
			}
		}
	}

	// Reopen warmup: after market closed or data frozen, wait N new M15 candles before evaluating signals again
	if l.Config.DataQuality != nil && l.reopenWarmupCandlesRemaining > 0 {
		l.lastDecision = "waiting:warmup"
		logger.Info("MARKET_REOPEN_WARMUP: waiting for %d candles; skipping signal evaluation", l.reopenWarmupCandlesRemaining)
		if !lastM15.IsZero() && lastM15 != l.prevM15LastTime {
			l.reopenWarmupCandlesRemaining--
			l.prevM15LastTime = lastM15
		}
		return
	}

	m15LastCl := lastM15.UTC().Add(-3 * time.Hour)
	if chileLoc, err := time.LoadLocation("America/Santiago"); err == nil {
		m15LastCl = lastM15.UTC().In(chileLoc)
	}
	logger.Info("evaluating: marketStatus=%s m15_last_utc=%s m15_last_cl=%s",
		l.lastKnownMarketStatus, lastM15.UTC().Format("2006-01-02T15:04"), m15LastCl.Format("2006-01-02T15:04"))

	lookback := l.Config.Strategy.SwingLookback
	if lookback <= 0 {
		lookback = 5
	}
	sweepOpts := &strategy.SweepOptions{
		CloseTolerance: l.Config.Strategy.SweepCloseTolerance,
		MinPenetration: l.Config.Strategy.SweepMinPenetration,
		MinWickRatio:   l.Config.Strategy.SweepMinWickRatio,
	}
	in := strategy.BuildComposerInput(
		candlesM15,
		lookback,
		l.Config.Strategy.EntryDelayCandles,
		l.Config.Indicators.RSIPeriod,
		l.Config.Indicators.ATRPeriod,
		sweepOpts,
		l.Config.Indicators.UseRSIWilder,
	)
	hasLiquidity := len(in.LiquidityZones) > 0
	hasSweep := strategy.HasSweep(in)
	atEquilibrium := in.EquilibriumTouch

	// setupDirection for state machine: from LiquidityEvent, else SweepBuy/SweepSell
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
	if len(candlesM15) > 0 {
		lastCandleTime = candlesM15[len(candlesM15)-1].Time
	}
	stateTTL := l.Config.Strategy.StateTTLCandles
	if stateTTL <= 0 {
		stateTTL = 12
	}

	l.Risk.ResetDailyIfNewDay(now)
	inTrade := l.State.State() == StateInTrade
	prevState := l.State.State()
	l.State.Transition(hasLiquidity, hasSweep, atEquilibrium, inTrade, TransitionParams{
		LastCandleTime:                lastCandleTime,
		StructureTrend:                in.Structure.Trend,
		SetupDirection:                setupDirection,
		StateTTLCandles:               stateTTL,
		InvalidateOnOppositeStructure: l.Config.Strategy.InvalidateOnOppositeStructure,
	})
	newState := l.State.State()
	if prevState != newState {
		logger.Info("state transition: %s -> %s", prevState, newState)
	}

	// Log market analysis
	if hasLiquidity {
		logger.Info("market analysis: liquidity zones detected (%d)", len(in.LiquidityZones))
	}
	if hasSweep {
		if in.LiquidityEvent != nil {
			logger.Info("market analysis: liquidity sweep detected (type=%s strength=%.2f)", in.LiquidityEvent.Type, in.LiquidityEvent.Strength)
		} else {
			sweepType := "SELL_SIDE"
			if in.SweepBuy {
				sweepType = "BUY_SIDE"
			}
			logger.Info("market analysis: liquidity sweep detected (type=%s from flags)", sweepType)
		}
	}
	if atEquilibrium {
		logger.Info("market analysis: price at equilibrium zone")
	}
	if in.Structure.Trend != "" {
		logger.Info("market analysis: structure=%s (HH=%v HL=%v LH=%v LL=%v)", in.Structure.Trend, in.Structure.HH, in.Structure.HL, in.Structure.LH, in.Structure.LL)
	}
	if !math.IsNaN(in.RSI) {
		logger.Info("market analysis: RSI=%.2f ATR=%.2f", in.RSI, in.ATR)
	}

	if l.Config.Strategy.SweepDebugLog && in.SweepDiagnostics != nil {
		d := in.SweepDiagnostics
		nBuy, nSell := 0, 0
		for _, z := range in.LiquidityZones {
			if z.Type == strategy.BuyLiquidity {
				nBuy++
			} else if z.Type == strategy.SellLiquidity {
				nSell++
			}
		}
		logger.Info("[LiquidityEngine] zones: %d (buy=%d sell=%d)", len(in.LiquidityZones), nBuy, nSell)
		logger.Info("[SweepAnalyzer] wickUp=%.2f wickDown=%.2f body=%.2f wickRatioUp=%.2f wickRatioDown=%.2f brokeBuy=%v brokeSell=%v closeInsideBuy=%v closeInsideSell=%v",
			d.WickSizeUp, d.WickSizeDown, d.BodySize, d.WickRatioUp, d.WickRatioDown, d.BrokeBuyZone, d.BrokeSellZone, d.CloseInsideBuy, d.CloseInsideSell)
		if d.SweepBuy || d.SweepSell {
			sweepType := "SELL_SIDE"
			if d.SweepBuy {
				sweepType = "BUY_SIDE"
			}
			logger.Success("SWEEP CONFIRMED type=%s zoneBuy=%.2f zoneSell=%.2f", sweepType, d.ZonePriceBuy, d.ZonePriceSell)
			typeChanged := (l.lastSweepType != "" && sweepType != l.lastSweepType && ((l.lastSweepType == "BUY_SIDE" && sweepType == "SELL_SIDE") || (l.lastSweepType == "SELL_SIDE" && sweepType == "BUY_SIDE")))
			l.lastSweepType = sweepType
			if l.NotifEmitter != nil {
				strength := 0.0
				if in.LiquidityEvent != nil {
					strength = in.LiquidityEvent.Strength
				} else {
					strength = in.IntentStrength
				}
				l.NotifEmitter.Emit(notifications.NotifEvent{
					Type: notifications.TypeSweepConfirmed, Category: notifications.CategoryMarket, Severity: notifications.SeverityInfo,
					Instrument: l.Epic, Session: strategy.GetSessionInfo(now), Timestamp: now,
					Payload: map[string]any{
						"sweep_type": sweepType, "strength": strength, "zoneBuy": d.ZonePriceBuy, "zoneSell": d.ZonePriceSell,
						"at_equilibrium": atEquilibrium, "typeChanged": typeChanged,
					},
				})
			}
		}
	}

	in.Context = l.Decision
	signal, ok := strategy.SignalComposerWithZones(in, l.Config.Indicators.MinATR, l.Config.Indicators.MaxATR, l.Config.Strategy.ScoreThreshold,
		l.Config.Strategy.RSIBuyLow, l.Config.Strategy.RSIBuyHigh, l.Config.Strategy.RSISellLow, l.Config.Strategy.RSISellHigh, l.Config.Strategy.RSIRequired, l.Config.Strategy.RSIMode)
	diagScore, diagDir, reasons := strategy.SignalDiagnostics(in, l.Config.Indicators.MinATR, l.Config.Indicators.MaxATR, l.Config.Strategy.ScoreThreshold,
		l.Config.Strategy.RSIBuyLow, l.Config.Strategy.RSIBuyHigh, l.Config.Strategy.RSISellLow, l.Config.Strategy.RSISellHigh, l.Config.Strategy.RSIRequired, l.Config.Strategy.RSIMode)
	trendM15 := marketstate.TrendFromCandles(candlesM15, lookback)
	session := strategy.GetSessionInfo(now)
	blockOverlap := l.Config.Strategy.BlockLondonNYOverlap == nil || *l.Config.Strategy.BlockLondonNYOverlap
	if session == strategy.SessionLondonNY && blockOverlap {
		logger.Info("scanning skipped: session overlap LONDON+NY blocked")
		return
	}
	if l.Journal != nil {
		score := diagScore
		sigDir := diagDir
		rejectReasons := reasons
		if ok && signal != nil {
			score = signal.Score
			sigDir = signal.Direction
			rejectReasons = nil
		}
		_ = l.Journal.WriteSetupEvaluated(journal.SetupEvaluated{
			Epic:            l.Epic,
			TfEntry:         "M15",
			TfBias:          "H1",
			Session:         session,
			State:           string(newState),
			TrendH4:         trendH4,
			TrendH1:         trendH1,
			TrendM15:        trendM15,
			ATR:             in.ATR,
			RSI:             in.RSI,
			ATRBucket:       atrBucket(in.ATR),
			HasLiquidity:    hasLiquidity,
			HasSweep:        hasSweep,
			AtEquilibrium:   atEquilibrium,
			Score:           score,
			Threshold:       l.Config.Strategy.ScoreThreshold,
			SignalOK:        ok && signal != nil,
			SignalDirection: sigDir,
			RejectReasons:   rejectReasons,
		})
	}
	if !ok || signal == nil {
		reasonCode := notifications.ReasonToCode(journal.RejectNoSignal)
		l.lastDecision = "rejected:" + reasonCode
		if len(reasons) > 0 {
			logger.Info("no signal: score=%d/%d direction=%s reasons=[%s]", diagScore, l.Config.Strategy.ScoreThreshold, diagDir, strings.Join(reasons, ", "))
		} else {
			logger.Info("no signal: score=%d/%d direction=%s", diagScore, l.Config.Strategy.ScoreThreshold, diagDir)
		}
		if l.Journal != nil {
			_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
				Epic:         l.Epic,
				State:        string(newState),
				Session:      session,
				RejectReason: journal.RejectNoSignal,
				Direction:    diagDir,
				Score:        diagScore,
			})
		}
		// InTrade: do not notify REJ_NO_SIGNAL (noise; journal and log kept)
		emitReject := true
		if l.State.State() == StateInTrade && reasonCode == notifications.RejNoSignal {
			emitReject = false
		}
		if emitReject && l.NotifEmitter != nil {
			l.NotifEmitter.Emit(notifications.NotifEvent{
				Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
				Instrument: l.Epic, Session: session, Timestamp: now,
				Payload: map[string]any{"reason_code": reasonCode, "direction": diagDir, "score": diagScore, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": 0},
			})
		}
		return
	}

	logger.Success("signal generated: direction=%s entry=%.2f sl=%.2f tp=%.2f score=%d confidence=%.2f",
		signal.Direction, signal.Entry, signal.StopLoss, signal.TakeProfit, signal.Score, signal.Confidence)
	stopDist := risk.StopDistance(signal)
	if stopDist <= 0 {
		stopDist = 1
	}
	rr := 0.0
	if stopDist > 0 {
		tpDist := signal.TakeProfit - signal.Entry
		if tpDist < 0 {
			tpDist = -tpDist
		}
		rr = tpDist / stopDist
	}
	sweepType := ""
	sweepStrength := 0.0
	liquidityEvent := in.LiquidityEvent != nil
	if in.LiquidityEvent != nil {
		sweepType = in.LiquidityEvent.Type
		sweepStrength = in.LiquidityEvent.Strength
	} else if in.SweepBuy {
		sweepType = "BUY_SIDE"
		sweepStrength = in.IntentStrength
	} else if in.SweepSell {
		sweepType = "SELL_SIDE"
		sweepStrength = in.IntentStrength
	}
	if l.Journal != nil {
		_ = l.Journal.WriteSignalGenerated(journal.SignalGenerated{
			Epic:             l.Epic,
			State:            string(l.State.State()),
			Session:          session,
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
			EquilibriumTouch: atEquilibrium,
			SweepType:        sweepType,
			SweepStrength:    sweepStrength,
			LiquidityEvent:   liquidityEvent,
		})
	}
	if signal != nil {
		l.LastSignalAt = now
		l.LastSignalID = stratrade.SignalID(l.Epic, now, signal.Direction)
	}
	if l.OnLegacySignal != nil && signal != nil {
		l.OnLegacySignal(signal, now)
	}
	if l.NotifEmitter != nil {
		l.NotifEmitter.Emit(notifications.NotifEvent{
			Type: notifications.TypeSignalGenerated, Category: notifications.CategorySignals, Severity: notifications.SeverityImportant,
			Instrument: l.Epic, Session: session, Timestamp: now,
			Payload: map[string]any{
				"direction": signal.Direction, "entry": signal.Entry, "sl": signal.StopLoss, "tp": signal.TakeProfit,
				"score": signal.Score, "confidence": signal.Confidence, "structure_m15": trendM15, "trend_h1": trendH1, "rsi": in.RSI, "atr": in.ATR,
			},
		})
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
	if finalScore < l.Config.Strategy.ScoreThreshold {
		reasonCode := notifications.ReasonToCode("score_after_context")
		l.lastDecision = "rejected:" + reasonCode
		if l.Journal != nil {
			_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
				Epic:         l.Epic,
				State:        string(l.State.State()),
				Session:      session,
				RejectReason: "score_after_context",
				Direction:    signal.Direction,
				Entry:        signal.Entry,
				Score:        signal.Score,
			})
		}
		if l.NotifEmitter != nil {
			l.NotifEmitter.Emit(notifications.NotifEvent{
				Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
				Instrument: l.Epic, Session: session, Timestamp: now,
				Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": signal.Score, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
			})
		}
		return
	}
	if l.Config.Strategy.UseH1Filter && trendH1 != "" {
		if l.Config.Strategy.BlockH1Range && trendH1 == "RANGE" {
			reasonCode := notifications.ReasonToCode(journal.RejectH1Range)
			l.lastDecision = "rejected:" + reasonCode
			logger.Info("signal rejected: H1 RANGE blocked (no trade when H1 is RANGE)")
			if l.Journal != nil {
				_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
					Epic:         l.Epic,
					State:        string(l.State.State()),
					Session:      session,
					RejectReason: journal.RejectH1Range,
					Direction:    signal.Direction,
					Entry:        signal.Entry,
					Score:        signal.Score,
				})
			}
			if l.NotifEmitter != nil {
				l.NotifEmitter.Emit(notifications.NotifEvent{
					Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
					Instrument: l.Epic, Session: session, Timestamp: now,
					Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": signal.Score, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
				})
			}
			return
		}
		allowed := (signal.Direction == "BUY" && trendH1 == "BULLISH") || (signal.Direction == "SELL" && trendH1 == "BEARISH")
		if !allowed && trendH1 == "RANGE" && l.Config.Strategy.Crypto {
			extra := l.Config.Strategy.H1RangeExtraScore
			if extra <= 0 {
				extra = 1
			}
			minScore := l.Config.Strategy.ScoreThreshold + extra
			if finalScore >= minScore {
				allowed = true
				logger.Info("H1 RANGE override allowed (crypto): score=%d minScore=%d threshold=%d extra=%d dir=%s trendH1=%s",
					finalScore, minScore, l.Config.Strategy.ScoreThreshold, extra, signal.Direction, trendH1)
			} else {
				reasonCode := notifications.ReasonToCode(journal.RejectH1RangeCryptoScore)
				l.lastDecision = "rejected:" + reasonCode
				logger.Info("signal rejected: H1 RANGE (crypto) requires score>=%d (got=%d) direction=%s trendH1=%s",
					minScore, finalScore, signal.Direction, trendH1)
				if l.Journal != nil {
					_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
						Epic:         l.Epic,
						State:        string(l.State.State()),
						Session:      session,
						RejectReason: journal.RejectH1RangeCryptoScore,
						Direction:    signal.Direction,
						Entry:        signal.Entry,
						Score:        signal.Score,
					})
				}
				if l.NotifEmitter != nil {
					l.NotifEmitter.Emit(notifications.NotifEvent{
						Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
						Instrument: l.Epic, Session: session, Timestamp: now,
						Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": finalScore, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
					})
				}
				return
			}
		}
		if !allowed {
			reasonCode := notifications.ReasonToCode(journal.RejectH1Filter)
			l.lastDecision = "rejected:" + reasonCode
			logger.Info("signal rejected: H1 bias filter (trendH1=%s direction=%s)", trendH1, signal.Direction)
			if l.Journal != nil {
				_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
					Epic:         l.Epic,
					State:        string(l.State.State()),
					Session:      session,
					RejectReason: journal.RejectH1Filter,
					Direction:    signal.Direction,
					Entry:        signal.Entry,
					Score:        signal.Score,
				})
			}
			if l.NotifEmitter != nil {
				l.NotifEmitter.Emit(notifications.NotifEvent{
					Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
					Instrument: l.Epic, Session: session, Timestamp: now,
					Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": signal.Score, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
				})
			}
			return
		}
	}

	if !l.State.CanSendOrder() {
		reasonCode := notifications.ReasonToCode(journal.RejectStateNotReady)
		l.lastDecision = "rejected:" + reasonCode
		logger.Info("signal valid but state=%s (need READY to send order)", l.State.State())
		if l.Journal != nil {
			_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
				Epic:         l.Epic,
				State:        string(l.State.State()),
				Session:      session,
				RejectReason: journal.RejectStateNotReady,
				Direction:    signal.Direction,
				Entry:        signal.Entry,
				Score:        signal.Score,
			})
		}
		if l.NotifEmitter != nil {
			l.NotifEmitter.Emit(notifications.NotifEvent{
				Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
				Instrument: l.Epic, Session: session, Timestamp: now,
				Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": signal.Score, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
			})
		}
		return
	}

	// M5 refiner: timing only in READY (soft reject recycles to WAIT_PULLBACK)
	if l.Config.M5Refiner != nil && l.Config.M5Refiner.Enabled {
		if m5Status == "DEGRADED" && l.Config.M5Refiner.DegradeIfUnavailable {
			logger.Info("m5_status=DEGRADED; proceeding M15-only")
		} else if m5Status == "OK" && len(candlesM5) >= 10 {
			m5Lookback := 5
			m5Result := marketstate.M5TimingAndScore(candlesM5, signal.Direction, m5Lookback)
			if m5Result.Timing == marketstate.M5TimingReject {
				reasonCode := notifications.ReasonToCode("M5_TIMING_REJECT")
				l.lastDecision = "rejected:" + reasonCode
				l.State.ToWaitPullback()
				logger.Warn("M5_TIMING_REJECT -> WAIT_PULLBACK (soft)")
				if l.Journal != nil {
					_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
						Epic:         l.Epic,
						State:        string(l.State.State()),
						Session:      session,
						RejectReason: "M5_TIMING_REJECT",
						Direction:    signal.Direction,
						Entry:        signal.Entry,
						Score:        signal.Score,
					})
				}
				if l.NotifEmitter != nil {
					l.NotifEmitter.Emit(notifications.NotifEvent{
						Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
						Instrument: l.Epic, Session: session, Timestamp: now,
						Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": signal.Score, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
					})
				}
				return
			}
			logger.Info("M5_TIMING_CONFIRM m5_score=%d", m5Result.Score)
		}
	}

	if ok && signal != nil {
		l.LastSignalText = signal.Direction
	}
	if l.UnknownPositions > 0 {
		l.lastDecision = "rejected:" + journal.RejectUnknownPos
		logger.Warn("skip order: %s count=%d", journal.RejectUnknownPos, l.UnknownPositions)
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: journal.RejectUnknownPos, Direction: signal.Direction})
		return
	}
	if l.RequireRuntimeMoney && l.Money.ValidationStatus != money.RuntimeValidated {
		l.lastDecision = "rejected:" + journal.RejectInstrumentSpec
		logger.Warn("skip order: monetary spec not strategy-executable")
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: journal.RejectInstrumentSpec, Direction: signal.Direction})
		return
	}
	if l.Kill != nil && l.Kill.HaltNewOrders() {
		l.lastDecision = "rejected:" + journal.RejectKillSwitch
		logger.Warn("skip order: %s", journal.RejectKillSwitch)
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: journal.RejectKillSwitch, Direction: signal.Direction, Entry: signal.Entry, SL: signal.StopLoss, TP: signal.TakeProfit, Score: signal.Score})
		if l.Journal != nil {
			_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
				Epic: l.Epic, State: string(l.State.State()), Session: session,
				RejectReason: journal.RejectKillSwitch, Direction: signal.Direction, Entry: signal.Entry, Score: signal.Score,
			})
		}
		return
	}

	execMode := l.Config.ExecMode()
	if execMode == config.ExecutionDisabled {
		l.lastDecision = "rejected:" + journal.RejectExecDisabled
		l.journalLife(journal.Lifecycle{
			Event: journal.EventOrderSkipped, Reason: journal.RejectExecDisabled,
			Direction: signal.Direction, Entry: signal.Entry, SL: signal.StopLoss, TP: signal.TakeProfit, Score: signal.Score,
		})
		return
	}

	if l.Week != nil && l.Week.BlocksNewStrategy() {
		l.lastDecision = "rejected:MAX_STRATEGY_TRADES"
		logger.Warn("skip order: strategy maxTrades=%d already completed=%d", l.Week.Snapshot().MaxTrades, l.Week.Snapshot().CompletedStrategyTrades)
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: "MAX_STRATEGY_TRADES", Direction: signal.Direction, Score: signal.Score})
		return
	}
	if l.Week != nil && l.Week.HasOpen() {
		if l.Forensic != nil {
			l.Forensic.NoteDuplicateOpen()
		}
		l.lastDecision = "rejected:DUPLICATE_OPEN"
		logger.Warn("skip order: duplicate_open_attempts strategy trade already open")
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: "DUPLICATE_OPEN", Direction: signal.Direction, Score: signal.Score})
		return
	}

	valid, err := l.Risk.ValidateSignal(signal)
	if err != nil || !valid {
		reasonCode := notifications.ReasonToCode(journal.RejectRiskReject)
		l.lastDecision = "rejected:" + reasonCode
		l.emitForensic(stratrade.EvRiskEvaluated, "FLAT", string(l.State.State()), "risk evaluated")
		l.emitForensic(stratrade.EvRiskRejected, "FLAT", string(l.State.State()), errString(err))
		logger.Warn("loop: risk reject: %v", err)
		if l.Journal != nil {
			_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
				Epic:         l.Epic,
				State:        string(l.State.State()),
				Session:      session,
				RejectReason: journal.RejectRiskReject,
				Direction:    signal.Direction,
				Entry:        signal.Entry,
				Score:        signal.Score,
			})
		}
		if l.NotifEmitter != nil {
			l.NotifEmitter.Emit(notifications.NotifEvent{
				Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
				Instrument: l.Epic, Session: session, Timestamp: now,
				Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": signal.Score, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
			})
		}
		return
	}

	// Spread check: skip if spread above threshold (safety)
	spread := 0.0
	if l.Config.API.MaxSpread > 0 {
		if details, err := l.Client.GetMarketDetails(ctx, l.Epic); err == nil && details.Snapshot.Offer > details.Snapshot.Bid {
			spread = details.Snapshot.Offer - details.Snapshot.Bid
			if spread > l.Config.API.MaxSpread {
				reasonCode := notifications.ReasonToCode(journal.RejectSpreadReject)
				l.lastDecision = "rejected:" + reasonCode
				logger.Warn("skip order: spread %.2f > max %.2f", spread, l.Config.API.MaxSpread)
				if l.Journal != nil {
					_ = l.Journal.WriteSignalRejected(journal.SignalRejected{
						Epic:         l.Epic,
						State:        string(l.State.State()),
						Session:      session,
						RejectReason: journal.RejectSpreadReject,
						Direction:    signal.Direction,
						Entry:        signal.Entry,
						Score:        signal.Score,
					})
				}
				if l.NotifEmitter != nil {
					l.NotifEmitter.Emit(notifications.NotifEvent{
						Type: notifications.TypeSignalRejected, Category: notifications.CategorySignals, Severity: notifications.SeverityInfo,
						Instrument: l.Epic, Session: session, Timestamp: now,
						Payload: map[string]any{"reason_code": reasonCode, "direction": signal.Direction, "score": signal.Score, "threshold": l.Config.Strategy.ScoreThreshold, "trend_h1": trendH1, "trend_h4": trendH4, "confidence": signal.Confidence},
					})
				}
				return
			}
		}
	}

	if !l.Spec.SizingComplete() {
		if details, dErr := l.Client.GetMarketDetails(ctx, l.Epic); dErr == nil {
			if spec, sErr := market.SpecFromDetails(details, l.Config.Risk.ValuePerPoint); sErr == nil {
				l.Spec = spec
			}
		}
	}
	if !l.Spec.SizingComplete() {
		l.lastDecision = "rejected:" + journal.RejectInstrumentSpec
		logger.Warn("skip order: %s", journal.RejectInstrumentSpec)
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: journal.RejectInstrumentSpec, Direction: signal.Direction, Score: signal.Score})
		return
	}
	size, err := risk.ComputeSize(l.Risk.GetBalance(), l.Config.Risk.RiskPerTrade, stopDist, l.Spec)
	if err != nil {
		reason := journal.RejectRiskReject
		errText := err.Error()
		if strings.Contains(errText, risk.ReasonMinSizeExceedsRisk) {
			reason = journal.RejectMinSizeRisk
		} else if strings.Contains(errText, risk.ReasonInstrumentSpec) {
			reason = journal.RejectInstrumentSpec
		} else if strings.Contains(errText, risk.ReasonMaxSizeExceeded) {
			reason = journal.RejectMaxSize
		}
		l.lastDecision = "rejected:" + reason
		logger.Warn("loop: position size: %v", err)
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: reason, Error: errText, Direction: signal.Direction, Score: signal.Score})
		return
	}

	l.captureStrategySnapshot(now, session, signal, size, stopDist, spread)
	l.emitForensic(stratrade.EvSignalObserved, l.lastKnownMarketStatus, string(l.State.State()), "legacy signal")
	l.emitForensic(stratrade.EvRiskEvaluated, l.lastKnownMarketStatus, string(l.State.State()), "risk evaluated")
	l.emitForensic(stratrade.EvRiskAccepted, l.lastKnownMarketStatus, string(l.State.State()), "risk accepted")
	l.journalLife(journal.Lifecycle{
		Event: journal.EventOrderIntent, Direction: signal.Direction, Size: size,
		Entry: signal.Entry, SL: signal.StopLoss, TP: signal.TakeProfit, Score: signal.Score,
	})
	l.emitForensic(stratrade.EvOrderIntent, l.lastKnownMarketStatus, string(l.State.State()), "")
	if execMode == config.ExecutionDryRun {
		l.lastDecision = "dry_run"
		logger.Info("DRY_RUN would_have_sent epic=%s direction=%s size=%.4f entry=%.2f sl=%.2f tp=%.2f score=%d state=%s",
			l.Epic, signal.Direction, size, signal.Entry, signal.StopLoss, signal.TakeProfit, signal.Score, l.State.State())
		l.journalLife(journal.Lifecycle{
			Event: journal.EventOrderDryRun, Reason: "would_have_sent",
			Direction: signal.Direction, Size: size, Entry: signal.Entry, SL: signal.StopLoss, TP: signal.TakeProfit, Score: signal.Score,
		})
		return
	}

	if l.Journal != nil {
		_ = l.Journal.WriteOrderAttempt(journal.OrderAttempt{
			Epic:                l.Epic,
			Direction:           signal.Direction,
			Size:                size,
			Entry:               signal.Entry,
			SL:                  signal.StopLoss,
			TP:                  signal.TakeProfit,
			MaxSpread:           l.Config.API.MaxSpread,
			Spread:              spread,
			LiveConfirmRequired: false,
		})
	}
	if l.NotifEmitter != nil {
		l.NotifEmitter.Emit(notifications.NotifEvent{
			Type: notifications.TypeOrderSent, Category: notifications.CategoryExecution, Severity: notifications.SeverityImportant,
			Instrument: l.Epic, Session: session, Timestamp: now,
			Payload: map[string]any{"direction": signal.Direction, "size": size, "entry": signal.Entry, "sl": signal.StopLoss, "tp": signal.TakeProfit},
		})
	}

	if l.Kill != nil && l.Kill.HaltNewOrders() {
		l.lastDecision = "rejected:" + journal.RejectKillSwitch
		logger.Warn("skip order immediately before broker mutation: %s", journal.RejectKillSwitch)
		l.journalLife(journal.Lifecycle{Event: journal.EventOrderSkipped, Reason: journal.RejectKillSwitch, Direction: signal.Direction, Size: size})
		return
	}

	prov := l.Provider
	if prov == nil {
		prov = l.Exec
	}
	l.emitForensic(stratrade.EvBrokerOpenRequest, l.lastKnownMarketStatus, string(l.State.State()), "open request")
	openRes, err := prov.OpenPosition(ctx, execution.OpenRequest{
		Epic: l.Epic, Direction: signal.Direction, Size: size,
		StopLoss: signal.StopLoss, TakeProfit: signal.TakeProfit, Signal: signal,
	})
	dealRef := ""
	if openRes != nil {
		dealRef = openRes.DealReference
	}
	if err != nil {
		if strings.Contains(err.Error(), "401") && l.ReloginFn != nil {
			_ = l.ReloginFn()
		}
		l.State.ToCooldown()
		execution.LogTrade(signal, size, "", err)
		logger.Error("failed to open position: %v", err)
		if l.Journal != nil {
			_ = l.Journal.WriteOrderResult(journal.OrderResult{
				Epic:   l.Epic,
				Status: "REJECTED",
				Error:  err.Error(),
			})
		}
		if l.NotifEmitter != nil {
			l.NotifEmitter.Emit(notifications.NotifEvent{
				Type: notifications.TypeOrderRejected, Category: notifications.CategoryExecution, Severity: notifications.SeverityCritical,
				Instrument: l.Epic, Session: session, Timestamp: time.Now().UTC(),
				Payload: map[string]any{"error": err.Error(), "deal_ref": ""},
			})
		}
		return
	}
	if l.Journal != nil {
		_ = l.Journal.WriteOrderResult(journal.OrderResult{
			Epic:    l.Epic,
			Status:  "ACCEPTED",
			DealRef: dealRef,
		})
	}
	l.journalLife(journal.Lifecycle{
		Event: journal.EventOrderSubmitted, DealRef: dealRef, Direction: signal.Direction,
		Size: size, Entry: signal.Entry, SL: signal.StopLoss, TP: signal.TakeProfit, Status: "SUBMITTED",
	})
	l.State.ToInTrade()
	l.lastOpenedDealRef = dealRef
	l.lastOpenedSignal = signal
	l.lastOpenedAt = now
	if l.Forensic != nil {
		l.Forensic.Bind(l.LastSignalID, dealRef, "")
	}
	if l.Week != nil && l.Forensic != nil {
		_ = l.Week.MarkOpen(l.Forensic.ID(), l.LastSignalID)
	}
	execution.LogTrade(signal, size, dealRef, nil)

	// Refresh positions count to get accurate "X of Y" (filtered by epic)
	currentOpenCount := 0
	if pos, err := l.Client.GetPositions(ctx); err == nil {
		for _, p := range pos.Positions {
			if p.GetEpic() == l.Epic {
				currentOpenCount++
			}
		}
		l.Risk.SetOpenCount(currentOpenCount)
		logger.Success("position opened: %d of %d (dealRef=%s direction=%s entry=%.2f sl=%.2f tp=%.2f size=%.2f epic=%s)",
			currentOpenCount, l.Config.Risk.MaxTrades, dealRef, signal.Direction, signal.Entry, signal.StopLoss, signal.TakeProfit, size, l.Epic)
	} else {
		currentOpenCount = 1
		logger.Success("position opened: dealRef=%s direction=%s entry=%.2f size=%.2f (unable to refresh count)",
			dealRef, signal.Direction, signal.Entry, size)
	}
	if l.NotifEmitter != nil {
		now := time.Now().UTC()
		l.NotifEmitter.Emit(notifications.NotifEvent{
			Type: notifications.TypePositionOpened, Category: notifications.CategoryExecution, Severity: notifications.SeverityCritical,
			Instrument: l.Epic, Session: session, Timestamp: now,
			Payload: map[string]any{
				"direction":      signal.Direction,
				"fill_price":     signal.Entry,
				"sl":             signal.StopLoss,
				"tp":             signal.TakeProfit,
				"open_positions": currentOpenCount,
				"deal_ref":       dealRef,
				"event_id":       uuid.New().String(),
				"event_ts_utc":   now.Format(time.RFC3339),
				"instance_id":    l.InstanceID,
				"run_id":         l.RunID,
			},
		})
	}
	telemetry.PublishPositionsBestEffort(l.Telemetry, ctx, l.InstanceID, currentOpenCount, l.Config.Risk.MaxTrades, dealRef)

	confirm, err := l.Exec.ConfirmDeal(ctx, dealRef)
	if err == nil && confirm != nil {
		if l.Forensic != nil {
			l.Forensic.Bind(l.LastSignalID, confirm.DealReference, confirm.DealID)
		}
		l.emitForensic(stratrade.EvBrokerConfirm, confirm.Status, string(l.State.State()), "confirm")
		l.journalLife(journal.Lifecycle{
			Event: journal.EventOrderConfirmed, DealRef: confirm.DealReference, DealID: confirm.DealID,
			Direction: confirm.Direction, Size: size, ConfirmedSize: confirm.Size, FillPrice: confirm.Level,
			Status: confirm.Status, Entry: signal.Entry, SL: signal.StopLoss, TP: signal.TakeProfit,
		})
		l.journalLife(journal.Lifecycle{
			Event: journal.EventPositionOpen, DealRef: confirm.DealReference, DealID: confirm.DealID,
			Direction: confirm.Direction, Size: confirm.Size, FillPrice: confirm.Level, Status: confirm.Status,
		})
		if confirm.DealID != "" {
			l.lastOpenedDealRef = confirm.DealID
		}
		l.emitForensic(stratrade.EvPositionResolved, confirm.Status, string(l.State.State()), confirm.DealID)
		l.reconcileOpenedPosition(ctx, signal, size, confirm)
	}
	if err != nil {
		logger.Warn("loop: confirm deal: %v", err)
	} else if l.Journal != nil && confirm != nil {
		fillPrice := confirm.Level
		slippage := 0.0
		if signal.Direction == "BUY" && fillPrice > signal.Entry {
			slippage = fillPrice - signal.Entry
		} else if signal.Direction == "SELL" && fillPrice < signal.Entry {
			slippage = signal.Entry - fillPrice
		}
		_ = l.Journal.WriteOrderResult(journal.OrderResult{
			Epic:      l.Epic,
			Status:    confirm.Status,
			DealRef:   confirm.DealReference,
			DealID:    confirm.DealID,
			FillPrice: fillPrice,
			Slippage:  slippage,
		})
	}
}

// logIdleLine logs a single throttled idle line (OFF_HOURS or MARKET_CLOSED) with countdown and progress bar.
// Call only when throttle allows (lastIdleLog.IsZero() || time.Since(lastIdleLog) >= idleLogInterval).
func (l *Loop) logIdleLine(reason string, now time.Time) {
	nextSession, nextAt, ok := strategy.NextSessionStart(now, l.Config.Strategy.TradingSessions)
	if !ok {
		logger.Info("idle: %s | next session unknown (allowed=%v)", reason, l.Config.Strategy.TradingSessions)
	} else {
		remaining := nextAt.Sub(now)
		logger.Info("%s", formatIdleLine(reason, nextSession, remaining, idleProgressBarWidth))
	}
	l.lastIdleLog = now
}

// formatIdleLine builds a single log line for idle state (OFF_HOURS or MARKET_CLOSED) with countdown to next session.
// barWidth is used for progress bar (optional; 0 means countdown only).
func formatIdleLine(reason, nextSession string, remaining time.Duration, barWidth int) string {
	countdown := formatCountdown(remaining)
	msg := fmt.Sprintf("idle: %s | next %s in %s", reason, nextSession, countdown)
	if barWidth > 0 {
		pct := 0.0
		const maxWindow = 24 * time.Hour
		if remaining < maxWindow {
			pct = 100 * (1 - float64(remaining)/float64(maxWindow))
			if pct < 0 {
				pct = 0
			}
			if pct > 100 {
				pct = 100
			}
		}
		msg += " " + formatProgressBar(pct, barWidth)
	}
	return msg
}

func formatCountdown(d time.Duration) string {
	if d <= 0 {
		return "0m"
	}
	d = d.Round(time.Minute)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h >= 24 {
		days := h / 24
		h = h % 24
		return fmt.Sprintf("%dd %dh", days, h)
	}
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func formatProgressBar(pct float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int((pct / 100) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	empty := width - filled
	const fullBlock = "\u2593" // dark block
	const emptyDot = "\u00B7"  // middle dot
	bar := strings.Repeat(fullBlock, filled) + strings.Repeat(emptyDot, empty)
	return fmt.Sprintf("[%s] %d%%", bar, int(pct))
}

// nextHeartbeatBoundary returns the next heartbeat boundary time (>= now). Used when heartbeat_interval_minutes and heartbeat_align_top are set.
// For alignTop and 30 min: boundaries are :00 and :30 each hour. For 60 min: top of each hour. For !alignTop: boundaries every intervalMin from midnight.
func nextHeartbeatBoundary(now time.Time, intervalMin int, alignTop bool) time.Time {
	if intervalMin <= 0 {
		return time.Time{}
	}
	y, m, d := now.Date()
	hr, min, _ := now.Clock()
	if alignTop {
		if intervalMin == 30 {
			if min < 30 {
				return time.Date(y, m, d, hr, 30, 0, 0, time.UTC)
			}
			return time.Date(y, m, d, hr+1, 0, 0, 0, time.UTC)
		}
		if intervalMin == 60 {
			return time.Date(y, m, d, hr+1, 0, 0, 0, time.UTC)
		}
		// other interval: align to 0, intervalMin, ... within hour then next hour
		if min < intervalMin {
			return time.Date(y, m, d, hr, intervalMin, 0, 0, time.UTC)
		}
		return time.Date(y, m, d, hr+1, 0, 0, 0, time.UTC)
	}
	midnight := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	intervalDur := time.Duration(intervalMin) * time.Minute
	elapsed := now.Sub(midnight)
	n := int(elapsed / intervalDur)
	next := midnight.Add(time.Duration(n+1) * intervalDur)
	return next
}

// logStatusSummary logs a periodic status summary and emits HEARTBEAT (with marketStatus + lastDecision) and optionally DAILY_DD_UPDATE.
func (l *Loop) logStatusSummary(ctx context.Context, now time.Time) {
	state := l.State.State()
	balance := l.Risk.GetBalance()
	openCount := l.Risk.GetOpenCount()
	ddPct := l.Risk.DailyDrawdownPct()
	logger.Info("status summary: state=%s balance=%.2f openPositions=%d/%d dailyDD=%.2f%% (account-level) tick=%d",
		state, balance, openCount, l.Config.Risk.MaxTrades, ddPct, l.tickCount)
	marketStatus := l.lastKnownMarketStatus
	if marketStatus == "" {
		marketStatus = "UNKNOWN"
	}
	if l.NotifEmitter != nil {
		l.NotifEmitter.Emit(notifications.NotifEvent{
			Type: notifications.TypeHeartbeat, Category: notifications.CategoryHealth, Severity: notifications.SeverityInfo,
			Instrument: l.Epic, Timestamp: now,
			Payload: map[string]any{
				"state": string(state), "balance": balance, "open_positions": openCount, "daily_dd": ddPct, "tick": l.tickCount,
				"market_status": marketStatus, "last_decision": l.lastDecision,
			},
		})
		// DAILY_DD_UPDATE when crossing 50%, 80%, 100% of limit
		limit := l.Config.Risk.DailyDrawdownLimit
		if limit > 0 && ddPct >= limit*0.5 {
			action := "monitor"
			if ddPct >= limit {
				action = "halt"
				l.NotifEmitter.Emit(notifications.NotifEvent{
					Type: notifications.TypeTradingHalted, Category: notifications.CategoryRisk, Severity: notifications.SeverityCritical,
					Instrument: l.Epic, Timestamp: now,
					Payload: map[string]any{"reason": "DD limit reached", "state": string(state), "dd_pct": ddPct, "limit": limit},
				})
			} else if ddPct >= limit*0.8 {
				action = "reduce"
			}
			l.NotifEmitter.Emit(notifications.NotifEvent{
				Type: notifications.TypeDailyDDUpdate, Category: notifications.CategoryRisk, Severity: notifications.SeverityImportant,
				Instrument: l.Epic, Timestamp: now,
				Payload: map[string]any{"dd_pct": ddPct, "limit": limit, "action": action},
			})
		}
		// MAX_TRADES_REACHED when openCount == maxTrades
		if openCount >= l.Config.Risk.MaxTrades {
			l.NotifEmitter.Emit(notifications.NotifEvent{
				Type: notifications.TypeMaxTradesReached, Category: notifications.CategoryRisk, Severity: notifications.SeverityInfo,
				Instrument: l.Epic, Timestamp: now,
				Payload: map[string]any{"open_positions": openCount, "max_trades": l.Config.Risk.MaxTrades},
			})
		}
	}
	if l.Telemetry != nil {
		telemetry.PublishStatusBestEffort(l.Telemetry, ctx, l.InstanceID, l.RunID, telemetry.StatusSnapshot{
			State:            string(state),
			Balance:          balance,
			OpenPositions:    openCount,
			LastDecision:     l.lastDecision,
			MarketStatus:     marketStatus,
			DailyDrawdownPct: ddPct,
			Tick:             l.tickCount,
			UpdatedAt:        now.Format(time.RFC3339),
			RunID:            l.RunID,
			SchemaVersion:    1,
		})
	}
}

// CandlesForStrategy returns M15 candles for the strategy (used by loop).
func CandlesForStrategy(client *market.Client, ctx context.Context, epic, resolution string, max int, from, to time.Time) ([]models.Candle, error) {
	return client.GetPrices(ctx, epic, resolution, max, from, to)
}

// atrBucket returns a bucket label for ATR (for journal segmentación).
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

func (l *Loop) DecisionText() string { return l.lastDecision }

func (l *Loop) OpenedAt() time.Time { return l.lastOpenedAt }

func (l *Loop) OpenedSignal() *models.TradeSignal { return l.lastOpenedSignal }

func (l *Loop) SessionView(broker string) stratrade.SessionReport {
	allowed := []string{}
	if l.Config != nil {
		allowed = l.Config.Strategy.TradingSessions
	}
	return stratrade.ObserveSession(time.Now().UTC(), allowed, broker)
}

func (l *Loop) emitForensic(ev, broker, local, note string) {
	if l.Forensic == nil {
		return
	}
	_ = l.Forensic.Append(stratrade.Event{Event: ev, BrokerState: broker, LocalState: local, Note: note})
}

func (l *Loop) haltForensic(reason string) {
	logger.Error("CRITICAL %s — HALT_NEW_ORDERS", reason)
	if l.Forensic != nil {
		l.Forensic.Halt(reason)
		_ = l.Forensic.Append(stratrade.Event{Event: stratrade.EvPositionReconciled, Severity: "CRITICAL", Note: reason})
	}
	if l.Kill != nil {
		_ = l.Kill.HaltPersist()
	}
}

func (l *Loop) captureStrategySnapshot(now time.Time, session string, signal *models.TradeSignal, size, stopDist, spread float64) {
	if signal == nil {
		return
	}
	sigID := l.LastSignalID
	if sigID == "" {
		sigID = stratrade.SignalID(l.Epic, now, signal.Direction)
		l.LastSignalID = sigID
	}
	id := stratrade.TradeID(sigID, "", "", l.Epic)
	if l.Forensic == nil {
		l.Forensic = stratrade.NewRecorder(filepath.Join("research", "strategy-trades"), id)
	} else {
		l.Forensic.Bind(sigID, "", "")
	}
	mpu := l.Money.MoneyPerPriceUnit
	mid := signal.Entry
	ask := mid
	bid := mid
	if spread > 0 {
		ask = mid + spread/2
		bid = mid - spread/2
	}
	journalState := "MISSING"
	if l.Journal != nil {
		journalState = "OK"
	}
	ver := l.StrategyVersion
	if ver == "" {
		ver = "LEGACY"
	}
	snap := stratrade.FreezePreSignal(stratrade.PreSignalSnapshot{
		SignalID: sigID, StrategyTradeID: id, Timestamp: now, GitCommit: l.GitCommit,
		StrategyVersion: ver, Instrument: l.Epic, Direction: signal.Direction,
		Bid: bid, Ask: ask, Mid: mid, Spread: spread, MarketStatus: l.lastKnownMarketStatus,
		H1State: "", H4State: "", LegacyScore: signal.Score, ATR: 0, RSI: 0, Session: session,
		EntryCandidate: signal.Entry, StopLoss: signal.StopLoss, TakeProfit: signal.TakeProfit,
		StopDistance: stopDist, PositionSize: size, MoneyPerPriceUnit: mpu,
		ExpectedAccountRisk: stratrade.ExpectedRiskUSD(size, stopDist, mpu),
		AccountBalance: l.Risk.GetBalance(), DailyDrawdown: l.Risk.DailyDrawdownPct(),
		OpenPositionsBefore: l.Risk.GetOpenCount(),
		KillSwitch: l.Kill != nil && l.Kill.HaltNewOrders(),
		BrokerConnection: "CONNECTED", JournalState: journalState,
	})
	_ = l.Forensic.PersistPreSignal(snap)
	if l.ObserveIntel != nil {
		_ = l.Forensic.AttachIntel(l.ObserveIntel())
	}
}

func (l *Loop) sampleStrategyHeartbeat(now time.Time, positions []market.PositionItem) {
	if l.Forensic == nil || l.lastOpenedSignal == nil {
		return
	}
	var p market.PositionItem
	if len(positions) > 0 {
		p = positions[0]
	}
	brokerUPL := p.Position.Upnl
	if brokerUPL == 0 {
		brokerUPL = p.Position.ProfitLoss
	}
	closeable := p.Position.Level
	if closeable == 0 {
		closeable = l.lastOpenedSignal.Entry
	}
	exp := stratrade.ExpectedUPL(l.lastOpenedSignal.Direction, p.Position.Size, p.Position.Level, closeable, l.Money.MoneyPerPriceUnit)
	_ = l.Forensic.Sample(stratrade.Heartbeat{
		Timestamp: now, BrokerUPL: brokerUPL, ExpectedUPL: exp,
		SL: l.lastOpenedSignal.StopLoss, TP: l.lastOpenedSignal.TakeProfit,
		PositionState: string(l.State.State()),
	})
	l.emitForensic(stratrade.EvMonitoring, l.lastKnownMarketStatus, string(l.State.State()), "heartbeat")
}

func (l *Loop) reconcileOpenedPosition(ctx context.Context, signal *models.TradeSignal, size float64, confirm *execution.ConfirmResponse) {
	if confirm == nil {
		return
	}
	l.emitForensic(stratrade.EvPositionReconciled, confirm.Status, string(l.State.State()), "identity resolved")
	brokerSL, brokerTP := 0.0, 0.0
	brokerDir, brokerEpic := confirm.Direction, confirm.Epic
	brokerSize := confirm.Size
	if confirm.DealID != "" && l.Client != nil {
		if item, err := l.Client.GetPosition(ctx, confirm.DealID); err == nil && item != nil {
			brokerSL = item.Position.StopLevel
			brokerTP = item.Position.ProfitLevel
			if item.Position.Direction != "" {
				brokerDir = item.Position.Direction
			}
			if item.GetEpic() != "" {
				brokerEpic = item.GetEpic()
			}
			if item.Position.Size > 0 {
				brokerSize = item.Position.Size
			}
		}
	}
	dirOK := strings.EqualFold(brokerDir, signal.Direction)
	epicOK := brokerEpic == "" || strings.EqualFold(brokerEpic, l.Epic)
	sizeOK := brokerSize == 0 || math.Abs(brokerSize-size) <= math.Max(0.0001, size*0.01)
	protKnown := brokerSL != 0 || brokerTP != 0
	protOK := !protKnown || stratrade.ProtectiveOK(signal.StopLoss, signal.TakeProfit, brokerSL, brokerTP, 0.05)
	if !dirOK || !epicOK || !sizeOK || !protOK {
		l.haltForensic("protective/identity mismatch vs broker")
		return
	}
	note := "protective levels verified"
	if !protKnown {
		note = "protective levels requested; broker did not expose stop/profit"
	}
	l.emitForensic(stratrade.EvProtectiveVerified, confirm.Status, string(l.State.State()), note)
}

func (l *Loop) finalizeStrategyTrade(exitReason string, brokerPnL float64, localCount, brokerCount int) {
	reason := stratrade.ExitReasonFromEvidence(exitReason)
	l.emitForensic(stratrade.EvExitCondition, "FLAT", string(l.State.State()), reason)
	l.emitForensic(stratrade.EvBrokerCloseRequest, "FLAT", string(l.State.State()), "broker close observed")
	l.emitForensic(stratrade.EvBrokerCloseConfirm, "FLAT", string(l.State.State()), "close confirmation inferred")
	l.emitForensic(stratrade.EvPositionClosed, "FLAT", "COOLDOWN", reason)
	dup := 0
	if l.Forensic != nil {
		dup = l.Forensic.DuplicateOpens()
	}
	mismatch := !stratrade.CountsAgree(localCount, brokerCount)
	if mismatch {
		l.haltForensic("final broker/local mismatch")
	}
	ops := stratrade.OperationalOutcome(mismatch, dup > 0, false, false)
	if l.Forensic != nil {
		halted, _ := l.Forensic.Halted()
		if halted {
			ops = stratrade.OpsFailed
		}
	}
	l.emitForensic(stratrade.EvFinalReconciliation, "FLAT", "COOLDOWN", ops)
	if l.Forensic != nil {
		halted, _ := l.Forensic.Halted()
		_ = l.Forensic.Finish(stratrade.FinalRecord{
			Direction:          "",
			GrossBrokerPnL:     brokerPnL,
			ExitReason:         reason,
			TradeOutcome:       stratrade.TradeOutcome(brokerPnL, 0.01),
			OperationalOutcome: ops,
			LocalPositions:     localCount,
			BrokerPositions:    brokerCount,
			CountsAgree:        !mismatch,
			ProtectiveOK:       !halted,
		})
		_ = stratrade.WriteReport(l.Forensic.Dir())
	}
	if l.Week != nil && l.Forensic != nil {
		_ = l.Week.MarkClosed(l.Forensic.ID())
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
