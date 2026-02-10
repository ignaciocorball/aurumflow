// Package main is the entrypoint for AurumFlow trading bot.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"aurumflow/config"
	"aurumflow/internal/backtest"
	"github.com/google/uuid"
	"aurumflow/internal/core"
	"aurumflow/internal/journal"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/notifications"
	"aurumflow/pkg/models"
)

func main() {
	backtestPath := flag.String("backtest", "", "run backtest from candles file (JSON or CSV) and exit")
	backtestFrom := flag.String("backtest-from", "", "start date for backtest download (YYYY-MM-DD)")
	backtestTo := flag.String("backtest-to", "", "end date for backtest download (YYYY-MM-DD)")
	backtestEpic := flag.String("backtest-epic", "", "epic for backtest download (default: GOLD or resolve 'gold')")
	backtestDir := flag.String("backtest-dir", "backtesting", "directory to save/load backtest JSON files")
	backtestOutdir := flag.String("backtest-outdir", "", "if set, write journal and CSV under this dir (jsonl/ and csv/ subdirs) instead of candles dir")
	backtestForce := flag.Bool("backtest-force", false, "re-download even if backtest file already exists")
	flag.Parse()

	if *backtestFrom != "" && *backtestTo != "" {
		runBacktestDownload(*backtestFrom, *backtestTo, *backtestEpic, *backtestDir, *backtestOutdir, *backtestForce)
		return
	}
	if *backtestPath != "" {
		cfg := loadConfigForBacktest()
		runBacktest(*backtestPath, cfg, "")
		return
	}

	ctx := context.Background()
	configPath := "config/config.json"
	if p := os.Getenv("AURUMFLOW_CONFIG"); p != "" {
		configPath = p
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("load config: %v", err)
		os.Exit(1)
	}
	logDir := os.Getenv("AURUMFLOW_LOG_DIR")
	if logDir == "" {
		logDir = cfg.Logging.LogDir
	}
	if logDir == "" {
		logDir = "logs"
	}
	if err := logger.InitFile(logDir); err != nil {
		logger.Warn("log file disabled: %v", err)
	}
	defer logger.CloseFile()
	logger.Info("config loaded from %s (mode=%s)", configPath, cfg.API.Mode)

	client := market.NewClient(cfg.API.BaseURL)
	client.APIKey = cfg.API.APIKey
	session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
	if err != nil {
		logger.Error("create session: %v", err)
		os.Exit(1)
	}
	logger.Info("session created, currentAccountId=%s", session.CurrentAccountID)

	// List all accounts at startup
	for i, acc := range session.Accounts {
		marker := ""
		if acc.AccountID == session.CurrentAccountID {
			marker = " (current)"
		}
		logger.Info("account[%d] id=%s balance=%.2f available=%.2f%s",
			i, acc.AccountID, acc.Balance.Balance, acc.Balance.Available, marker)
	}

	// Switch to configured account if set
	if cfg.API.AccountID != "" && cfg.API.AccountID != session.CurrentAccountID {
		if err := client.SwitchAccount(ctx, cfg.API.AccountID); err != nil {
			logger.Error("switch account to %s: %v", cfg.API.AccountID, err)
			os.Exit(1)
		}
		logger.Info("switched to account %s", cfg.API.AccountID)
		// Re-fetch session/accounts for balance of switched account
		if ar, err := client.GetAccounts(ctx); err == nil {
			for _, acc := range ar.Accounts {
				if acc.AccountID == cfg.API.AccountID {
					session.CurrentAccountID = acc.AccountID
					session.Accounts = ar.Accounts
					break
				}
			}
		}
	}

	var balance float64
	for _, acc := range session.Accounts {
		if acc.AccountID == session.CurrentAccountID {
			balance = acc.Balance.Balance
			logger.Info("balance=%.2f available=%.2f", acc.Balance.Balance, acc.Balance.Available)
			break
		}
	}

	epic := os.Getenv("AURUMFLOW_EPIC")
	if epic == "" {
		epic, err = client.ResolveEpic(ctx, "gold")
		if err != nil {
			logger.Error("resolve epic: %v", err)
			os.Exit(1)
		}
		logger.Info("resolved epic for gold: %s", epic)
	} else {
		logger.Info("using epic from env: %s", epic)
	}

	if details, err := client.GetMarketDetails(ctx, epic); err == nil {
		logger.Info("market %s: minDealSize=%.2f maxDealSize=%.2f",
			details.Instrument.Epic, details.DealingRules.MinDealSize.Value, details.DealingRules.MaxDealSize.Value)
	}

	chileLoc, _ := time.LoadLocation("America/Santiago")
	now := time.Now().UTC()
	to := now
	from24h := now.Add(-24 * time.Hour)
	from10h := now.Add(-10 * time.Hour)
	// M5 at startup: use 10h window and max 300 so API returns most recent candles (24h + max 100 returns first 100 → stale).
	resolutions := []struct {
		name string
		res  string
		from time.Time
		to   time.Time
		max  int
	}{
		{"H1", cfg.Timeframes.H1, from24h, to, 100},
		{"M15", cfg.Timeframes.M15, from24h, to, 100},
		{"M5", cfg.Timeframes.M5, from10h, to, 300},
	}

	var lastM5, lastM15, lastH1 time.Time
	for _, tf := range resolutions {
		candles, err := client.GetPrices(ctx, epic, tf.res, tf.max, tf.from, tf.to)
		if err != nil {
			logger.Warn("prices %s: %v", tf.name, err)
			continue
		}
		logger.Info("%s: fetched %d candles (first=%s last=%s)",
			tf.name, len(candles), formatCandleTime(candles), formatCandleTimeLast(candles))
		if len(candles) > 0 {
			last := candles[len(candles)-1]
			lastUTC := last.Time.UTC()
			lastCl := lastUTC.Add(-3 * time.Hour)
			if chileLoc != nil {
				lastCl = lastUTC.In(chileLoc)
			}
			switch tf.name {
			case "M5":
				lastM5 = last.Time
			case "M15":
				lastM15 = last.Time
			case "H1":
				lastH1 = last.Time
			}
			logger.Info("%s last candle: O=%.2f H=%.2f L=%.2f C=%.2f last_utc=%s last_cl=%s",
				tf.name, last.Open, last.High, last.Low, last.Close,
				lastUTC.Format("2006-01-02T15:04"), lastCl.Format("2006-01-02T15:04"))
		}
	}
	// Time alignment: M5 must be within 15m of M15 or pipeline coherence is at risk.
	if !lastM5.IsZero() && !lastM15.IsZero() {
		diff := lastM15.Sub(lastM5)
		if diff < 0 {
			diff = -diff
		}
		if diff > 15*time.Minute {
			logger.Warn("DATA_STALE(M5): M5 last=%s M15 last=%s (diff=%.0fm > 15m); pipeline coherence at risk",
				lastM5.UTC().Format("2006-01-02T15:04"), lastM15.UTC().Format("2006-01-02T15:04"), diff.Minutes())
			if os.Getenv("AURUMFLOW_BLOCK_ON_DATA_STALE") == "1" {
				logger.Error("AURUMFLOW_BLOCK_ON_DATA_STALE=1: exiting due to DATA_STALE(M5)")
				os.Exit(1)
			}
		}
	}
	_ = lastH1

	openCount := 0
	if pos, err := client.GetPositions(ctx); err == nil {
		// Filter positions by epic (only count positions for the instrument we're trading)
		for _, p := range pos.Positions {
			if p.GetEpic() == epic {
				openCount++
			}
		}
		totalPositions := len(pos.Positions)
		if totalPositions > openCount {
			logger.Info("open positions: %d total, %d for epic %s (ignoring %d in other instruments)",
				totalPositions, openCount, epic, totalPositions-openCount)
		} else {
			logger.Info("open positions: %d", openCount)
		}
	}

	reloginFn := func() error {
		sess, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
		if err != nil {
			return err
		}
		if cfg.API.AccountID != "" && cfg.API.AccountID != sess.CurrentAccountID {
			return client.SwitchAccount(ctx, cfg.API.AccountID)
		}
		return nil
	}

	liveConfirmRequired := cfg.API.Mode == config.ModeLive && (os.Getenv("AURUMFLOW_LIVE_CONFIRM") != "1" && os.Getenv("AURUMFLOW_LIVE_CONFIRM") != "true")

	instanceID := epic
	if cfg.Notifications != nil && cfg.Notifications.InstanceID != "" {
		instanceID = cfg.Notifications.InstanceID
	}
	runID := uuid.New().String()

	var globalStore notifications.GlobalSentStore
	if cfg.Notifications != nil && cfg.Notifications.GlobalDedup != nil && cfg.Notifications.GlobalDedup.Type == "redis" && cfg.Notifications.GlobalDedup.Redis != nil {
		rs, err := notifications.NewRedisGlobalStore(cfg.Notifications.GlobalDedup.Redis)
		if err != nil {
			logger.Warn("global dedup Redis disabled: %v", err)
		} else {
			globalStore = rs
			defer rs.Close()
		}
	}
	notifEmitter := notifications.NewEmitter(cfg, globalStore)
	if notifEmitter.Active() {
		notifEmitter.Emit(notifications.NotifEvent{
			Type:       notifications.TypeBotStart,
			Category:   notifications.CategoryHealth,
			Severity:   notifications.SeverityImportant,
			Instrument: epic,
			Timestamp:  time.Now().UTC(),
			Payload: map[string]any{
				"version": "aurumflow",
				"mode":    cfg.API.Mode,
			},
		})
	}
	defer func() {
		if notifEmitter != nil && notifEmitter.Active() {
			notifEmitter.Emit(notifications.NotifEvent{
				Type:       notifications.TypeBotStop,
				Category:   notifications.CategoryHealth,
				Severity:   notifications.SeverityImportant,
				Instrument: epic,
				Timestamp:  time.Now().UTC(),
				Payload:    map[string]any{"mode": cfg.API.Mode},
			})
			notifEmitter.Close()
		}
	}()

	loop := core.NewLoop(cfg, client, epic, balance, openCount)
	loop.InstanceID = instanceID
	loop.RunID = runID
	loop.ReloginFn = reloginFn
	loop.LiveConfirmRequired = liveConfirmRequired
	loop.NotifEmitter = notifEmitter

	if cfg.Logging.JournalEnabled {
		jpath := filepath.Join(cfg.Logging.JournalDir, "trades.jsonl")
		jw, err := journal.NewFileWriter(jpath)
		if err != nil {
			logger.Warn("journal disabled: %v", err)
		} else {
			loop.Journal = jw
			defer func() {
				if err := jw.Close(); err != nil {
					logger.Warn("journal close: %v", err)
				}
			}()
			logger.Info("journal writing to %s", jpath)
		}
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		logger.Info("shutdown signal received")
		cancel()
	}()
	logger.Info("event loop starting (interval=60s)")
	loop.Run(runCtx)
	logger.Info("stopped")
}

func loadConfigForBacktest() *config.Config {
	configPath := "config/config.json"
	if p := os.Getenv("AURUMFLOW_CONFIG"); p != "" {
		configPath = p
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("load config: %v", err)
		os.Exit(1)
	}
	return cfg
}

func runBacktest(path string, cfg *config.Config, outDir string) {
	if cfg == nil {
		cfg = loadConfigForBacktest()
	}
	candlesM15, err := backtest.LoadCandlesFromFile(path)
	if err != nil {
		logger.Error("load candles: %v", err)
		os.Exit(1)
	}
	logger.Info("backtest: loaded %d candles from %s", len(candlesM15), path)
	if len(candlesM15) < 50 {
		logger.Error("backtest: need at least 50 candles")
		os.Exit(1)
	}
	var candlesH1 []models.Candle
	var candlesH4 []models.Candle
	if strings.Contains(path, "_M15_") {
		pathH1 := strings.Replace(path, "_M15_", "_H1_", 1)
		if _, err := os.Stat(pathH1); err == nil {
			candlesH1, err = backtest.LoadCandlesFromFile(pathH1)
			if err != nil {
				logger.Warn("backtest: could not load H1 file %s: %v (running without H1)", pathH1, err)
			} else {
				logger.Info("backtest: loaded %d H1 candles from %s", len(candlesH1), pathH1)
			}
		}
		pathH4 := strings.Replace(path, "_M15_", "_H4_", 1)
		if _, err := os.Stat(pathH4); err == nil {
			candlesH4, err = backtest.LoadCandlesFromFile(pathH4)
			if err != nil {
				logger.Warn("backtest: could not load H4 file %s: %v (running without H4)", pathH4, err)
			} else {
				logger.Info("backtest: loaded %d H4 candles from %s", len(candlesH4), pathH4)
			}
		}
	}
	btCfg := backtest.Config{
		RSIPeriod:            cfg.Indicators.RSIPeriod,
		ATRPeriod:            cfg.Indicators.ATRPeriod,
		MinATR:               cfg.Indicators.MinATR,
		MaxATR:               cfg.Indicators.MaxATR,
		SwingLookback:        cfg.Strategy.SwingLookback,
		EntryDelayCandles:    cfg.Strategy.EntryDelayCandles,
		ScoreThreshold:       cfg.Strategy.ScoreThreshold,
		RiskPerTrade:         cfg.Risk.RiskPerTrade,
		RSIRequired:          cfg.Strategy.RSIRequired,
		RSIMode:               cfg.Strategy.RSIMode,
		RSIBuyLow:             cfg.Strategy.RSIBuyLow,
		RSIBuyHigh:           cfg.Strategy.RSIBuyHigh,
		RSISellLow:            cfg.Strategy.RSISellLow,
		RSISellHigh:           cfg.Strategy.RSISellHigh,
		SweepCloseTolerance:  cfg.Strategy.SweepCloseTolerance,
		SweepMinPenetration:  cfg.Strategy.SweepMinPenetration,
		SweepMinWickRatio:    cfg.Strategy.SweepMinWickRatio,
		UseH1Filter:          cfg.Strategy.UseH1Filter,
		BlockH1Range:         cfg.Strategy.BlockH1Range,
		Crypto:               cfg.Strategy.Crypto,
		H1RangeExtraScore:    cfg.Strategy.H1RangeExtraScore,
		H1SwingLookback:      cfg.Strategy.H1SwingLookback,
		UseH4Filter:          cfg.Strategy.UseH4Filter,
		BlockH4Range:         cfg.Strategy.BlockH4Range,
		H4SwingLookback:      cfg.Strategy.H4SwingLookback,
		UseRSIWilder:         cfg.Indicators.UseRSIWilder,
		ValuePerPoint:        cfg.Risk.ValuePerPoint,
		MaxTrades:            cfg.Risk.MaxTrades,
		StateTTLCandles:      cfg.Strategy.StateTTLCandles,
		InvalidateOnOppositeStructure: cfg.Strategy.InvalidateOnOppositeStructure,
		BlockLondonNYOverlap: cfg.Strategy.BlockLondonNYOverlap == nil || *cfg.Strategy.BlockLondonNYOverlap,
		TradingSessions:      cfg.Strategy.TradingSessions,
		SpreadPoints:         0,
		SlippagePoints:       0,
		UseBreakEven:         cfg.Strategy.UseBreakEven,
		BreakEvenR:           cfg.Strategy.BreakEvenR,
		DailyDrawdownLimit:   cfg.Risk.DailyDrawdownLimit,
	}
	if btCfg.ValuePerPoint <= 0 {
		btCfg.ValuePerPoint = 1.0
	}
	if btCfg.H1SwingLookback <= 0 {
		btCfg.H1SwingLookback = 5
	}
	if btCfg.H4SwingLookback <= 0 {
		btCfg.H4SwingLookback = 5
	}
	// Journal and CSV for metrics/segmentación (heatmaps, trend_h1, session, atr_bucket, sweep_type)
	baseDir := filepath.Dir(path)
	baseName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	var journalPath, csvPath string
	if outDir != "" {
		journalPath = filepath.Join(outDir, "jsonl", "backtest_"+baseName+".jsonl")
		csvPath = filepath.Join(outDir, "csv", "trades_"+baseName+".csv")
	} else {
		journalPath = filepath.Join(baseDir, "journals", "backtest_"+baseName+".jsonl")
		csvPath = filepath.Join(baseDir, "trades_"+baseName+".csv")
	}
	var jw journal.JournalWriter
	if fw, err := journal.NewFileWriter(journalPath); err == nil {
		jw = fw
		defer func() { _ = fw.Close() }()
		logger.Info("backtest: writing journal to %s", journalPath)
	} else {
		logger.Warn("backtest: journal disabled: %v", err)
	}
	logger.Info("backtest: writing trades CSV to %s", csvPath)

	initialBalance := 10000.0
	var signals []models.TradeSignal
	var res backtest.Result
	signals, res = backtest.Run(candlesM15, candlesH1, candlesH4, btCfg, initialBalance, jw, csvPath)
	logger.Info("backtest: totalTrades=%d winners=%d losers=%d winrate=%.1f%% profitFactor=%.2f maxDrawdownPct=%.2f expectancy=%.2f",
		res.TotalTrades, res.Winners, res.Losers, res.Winrate, res.ProfitFactor, res.MaxDrawdownPct, res.Expectancy)
	if res.BarsSkippedSession > 0 || res.BarsSkippedLondonNY > 0 {
		logger.Info("backtest: barsSkippedSession=%d barsSkippedLondonNY=%d", res.BarsSkippedSession, res.BarsSkippedLondonNY)
	}
	_ = signals
}

func runBacktestDownload(backtestFrom, backtestTo, backtestEpic, backtestDir, backtestOutdir string, force bool) {
	cfg := loadConfigForBacktest()
	ctx := context.Background()

	client := market.NewClient(cfg.API.BaseURL)
	client.APIKey = cfg.API.APIKey
	session, err := client.CreateSession(ctx, cfg.API.Identifier, cfg.API.Password)
	if err != nil {
		logger.Error("create session: %v", err)
		os.Exit(1)
	}
	_ = session

	epic := backtestEpic
	if epic == "" {
		epic = os.Getenv("AURUMFLOW_EPIC")
	}
	if epic == "" {
		epic, err = client.ResolveEpic(ctx, "gold")
		if err != nil {
			logger.Error("resolve epic: %v", err)
			os.Exit(1)
		}
		logger.Info("resolved epic for gold: %s", epic)
	}
	epic = strings.ToUpper(epic)

	fromDt, err := time.Parse("2006-01-02", backtestFrom)
	if err != nil {
		logger.Error("invalid backtest-from: %v (use YYYY-MM-DD)", err)
		os.Exit(1)
	}
	toDt, err := time.Parse("2006-01-02", backtestTo)
	if err != nil {
		logger.Error("invalid backtest-to: %v (use YYYY-MM-DD)", err)
		os.Exit(1)
	}
	fromDt = time.Date(fromDt.Year(), fromDt.Month(), fromDt.Day(), 0, 0, 0, 0, time.UTC)
	toDt = time.Date(toDt.Year(), toDt.Month(), toDt.Day(), 23, 59, 59, 0, time.UTC)

	pathM15 := filepath.Join(backtestDir, epic+"_M15_"+backtestFrom+"_"+backtestTo+".json")
	pathH1 := filepath.Join(backtestDir, epic+"_H1_"+backtestFrom+"_"+backtestTo+".json")
	pathH4 := filepath.Join(backtestDir, epic+"_H4_"+backtestFrom+"_"+backtestTo+".json")
	if !force {
		if _, errM := os.Stat(pathM15); errM == nil {
			if _, errH := os.Stat(pathH1); errH == nil {
				logger.Info("backtest: using cached files %s and %s", pathM15, pathH1)
				runBacktest(pathM15, cfg, backtestOutdir)
				return
			}
		}
	}

	const chunkDays = 7
	const maxPerRequest = 1000

	// Download M15
	var allCandlesM15 []models.Candle
	currentStart := fromDt
	for currentStart.Before(toDt) {
		chunkEnd := currentStart.AddDate(0, 0, chunkDays)
		if chunkEnd.After(toDt) {
			chunkEnd = toDt
		}
		chunk, err := client.GetPrices(ctx, epic, market.ResolutionMinute15, maxPerRequest, currentStart, chunkEnd)
		if err != nil {
			logger.Warn("get prices M15 %s -> %s: %v", currentStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"), err)
			currentStart = chunkEnd
			time.Sleep(300 * time.Millisecond)
			continue
		}
		allCandlesM15 = append(allCandlesM15, chunk...)
		logger.Info("backtest: fetched M15 %d candles %s -> %s (total %d)", len(chunk), currentStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"), len(allCandlesM15))
		currentStart = chunkEnd
		time.Sleep(300 * time.Millisecond)
	}
	if len(allCandlesM15) == 0 {
		logger.Error("backtest: no M15 candles downloaded; check dates and epic")
		os.Exit(1)
	}
	sort.Slice(allCandlesM15, func(i, j int) bool { return allCandlesM15[i].Time.Before(allCandlesM15[j].Time) })
	if err := backtest.SaveCandlesToFile(pathM15, epic, "MINUTE_15", "capital.com", allCandlesM15); err != nil {
		logger.Error("save M15 candles: %v", err)
		os.Exit(1)
	}
	logger.Info("backtest: saved %d M15 candles to %s", len(allCandlesM15), pathM15)

	// Download H1 (same date range)
	var allCandlesH1 []models.Candle
	currentStart = fromDt
	for currentStart.Before(toDt) {
		chunkEnd := currentStart.AddDate(0, 0, chunkDays)
		if chunkEnd.After(toDt) {
			chunkEnd = toDt
		}
		chunk, err := client.GetPrices(ctx, epic, market.ResolutionHour, maxPerRequest, currentStart, chunkEnd)
		if err != nil {
			logger.Warn("get prices H1 %s -> %s: %v", currentStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"), err)
			currentStart = chunkEnd
			time.Sleep(300 * time.Millisecond)
			continue
		}
		allCandlesH1 = append(allCandlesH1, chunk...)
		logger.Info("backtest: fetched H1 %d candles %s -> %s (total %d)", len(chunk), currentStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"), len(allCandlesH1))
		currentStart = chunkEnd
		time.Sleep(300 * time.Millisecond)
	}
	if len(allCandlesH1) > 0 {
		sort.Slice(allCandlesH1, func(i, j int) bool { return allCandlesH1[i].Time.Before(allCandlesH1[j].Time) })
		if err := backtest.SaveCandlesToFile(pathH1, epic, "HOUR", "capital.com", allCandlesH1); err != nil {
			logger.Error("save H1 candles: %v", err)
			os.Exit(1)
		}
		logger.Info("backtest: saved %d H1 candles to %s", len(allCandlesH1), pathH1)
	}

	// Download H4 (regime; same date range)
	var allCandlesH4 []models.Candle
	currentStart = fromDt
	for currentStart.Before(toDt) {
		chunkEnd := currentStart.AddDate(0, 0, chunkDays)
		if chunkEnd.After(toDt) {
			chunkEnd = toDt
		}
		chunk, err := client.GetPrices(ctx, epic, market.ResolutionHour4, maxPerRequest, currentStart, chunkEnd)
		if err != nil {
			logger.Warn("get prices H4 %s -> %s: %v", currentStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"), err)
			currentStart = chunkEnd
			time.Sleep(300 * time.Millisecond)
			continue
		}
		allCandlesH4 = append(allCandlesH4, chunk...)
		logger.Info("backtest: fetched H4 %d candles %s -> %s (total %d)", len(chunk), currentStart.Format("2006-01-02"), chunkEnd.Format("2006-01-02"), len(allCandlesH4))
		currentStart = chunkEnd
		time.Sleep(300 * time.Millisecond)
	}
	if len(allCandlesH4) > 0 {
		sort.Slice(allCandlesH4, func(i, j int) bool { return allCandlesH4[i].Time.Before(allCandlesH4[j].Time) })
		if err := backtest.SaveCandlesToFile(pathH4, epic, "HOUR_4", "capital.com", allCandlesH4); err != nil {
			logger.Error("save H4 candles: %v", err)
			os.Exit(1)
		}
		logger.Info("backtest: saved %d H4 candles to %s", len(allCandlesH4), pathH4)
	}

	runBacktest(pathM15, cfg, backtestOutdir)
}

func formatCandleTime(c []models.Candle) string {
	if len(c) == 0 {
		return "n/a"
	}
	return c[0].Time.Format("2006-01-02 15:04")
}

func formatCandleTimeLast(c []models.Candle) string {
	if len(c) == 0 {
		return "n/a"
	}
	return c[len(c)-1].Time.Format("2006-01-02 15:04")
}
