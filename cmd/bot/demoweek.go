package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"aurumflow/config"
	"aurumflow/internal/execution"
	"aurumflow/internal/gates"
	"aurumflow/internal/journal"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/money"
	"aurumflow/internal/ops"
	"aurumflow/internal/radar"
	"aurumflow/internal/recovery"
	"aurumflow/internal/core"
	"aurumflow/internal/strategy"
)

func runDemoWeek(ctx context.Context, epic string, statusAddr string) {
	epic = strings.ToUpper(strings.TrimSpace(epic))
	if epic == "" {
		epic = "GOLD"
	}
	if !demoEnvConfigured() {
		logger.Error("DEMO-WEEK refused: missing AURUMFLOW_DEMO_* (will not load LIVE config.json)")
		os.Exit(1)
	}
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	ks := killswitch.New(false, killswitch.DefaultFile)
	pos, err := sess.client.GetPositions(ctx)
	if err != nil {
		logger.Error("GetPositions: %v — HALT_NEW_ORDERS", err)
		_ = ks.HaltPersist()
		os.Exit(1)
	}
	var broker []recovery.BrokerPos
	for _, p := range pos.Positions {
		broker = append(broker, recovery.BrokerPos{DealID: p.Position.DealID, Epic: p.GetEpic(), Direction: p.Position.Direction, Size: p.Position.Size})
	}
	classified, unknown := recovery.Reconcile(broker, nil)
	for _, c := range classified {
		logger.Info("RECOVERY dealId=%s epic=%s class=%s", c.Pos.DealID, c.Pos.Epic, c.Class)
	}
	if unknown > 0 {
		logger.Warn("UNKNOWN positions=%d — new orders blocked until resolved", unknown)
		_ = ks.HaltPersist()
	}

	details, err := sess.client.GetMarketDetails(ctx, epic)
	if err != nil {
		logger.Error("GetMarketDetails: %v", err)
		os.Exit(1)
	}
	liveSpec := money.FromMarketDetails(details)
	cached, _ := money.LoadSpec(money.CachePath("", epic))
	if cached.Epic != "" && money.MetadataChanged(cached, liveSpec) {
		logger.Warn("InstrumentSpec cache stale: %s", money.InvalidateReason(cached, liveSpec))
		cached = liveSpec
	}
	if cached.Epic == "" {
		cached = liveSpec
	}

	journalOK := true
	jpath := filepath.Join("journals", "demo-week.jsonl")
	jw, jerr := journal.NewFileWriter(jpath)
	if jerr != nil {
		journalOK = false
		logger.Warn("journal: %v", jerr)
	} else {
		defer func() { _ = jw.Close() }()
	}

	gateErr := gates.DemoWeekTradeAllowed(gates.Input{
		Environment: "demo", Host: sess.client.BaseURL, ExecutionMode: string(config.ExecutionDemo),
		AccountOK: true, KillSwitch: ks.HaltNewOrders(), MarketStatus: details.Snapshot.MarketStatus,
		Spec: cached, JournalOK: journalOK, DailyDDBlocked: false, DataStale: false, UnknownPositions: unknown,
	})
	execMode := config.ExecutionDisabled
	if gateErr != nil {
		logger.Warn("DEMO-WEEK execution gated: %v — analysis may continue, no new orders", gateErr)
	} else {
		execMode = config.ExecutionDemo
		sess.cfg.API.ExecutionMode = string(config.ExecutionDemo)
		logger.Info("DEMO-WEEK gates PASS — DEMO execution enabled for %s", epic)
	}

	st := ops.NewStatus()
	st.APIEnvironment = "demo"
	st.ExecutionMode = string(execMode)
	st.Host = sess.client.BaseURL
	st.KillSwitch = ks.HaltNewOrders()
	st.OpenPositions = len(pos.Positions)
	st.ExecutionEpic = epic
	st.MarketStatus = details.Snapshot.MarketStatus
	st.RadarMode = radar.ModeShadow
	started := time.Now().UTC()
	srv := ops.NewServer(statusAddr, st)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Warn("status server: %v", err)
		}
	}()
	logger.Info("status surface listening (healthz/status)")

	loop := core.NewLoop(sess.cfg, sess.client, epic, sess.acc.Balance.Balance, len(pos.Positions))
	loop.InstanceID = epic + "-demo-week"
	loop.RunID = uuid.New().String()
	loop.Kill = ks
	loop.Journal = jw
	loop.UnknownPositions = unknown
	loop.RequireRuntimeMoney = true
	loop.Money = cached
	loop.Decision = &strategy.DecisionContext{RadarMode: radar.ModeShadow}
	if execMode == config.ExecutionDryRun {
		loop.Provider = &execution.DryRunProvider{Inner: loop.Exec}
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()
	go func() {
		t := time.NewTicker(15 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-t.C:
				cur := srv.Get()
				cur.UptimeSeconds = ops.AgeSeconds(started)
				cur.KillSwitch = ks.HaltNewOrders()
				cur.LastStrategy = loop.LastSignalText
				if loop.Decision != nil {
					cur.RadarState = loop.Decision.RadarState
					cur.Pressure = loop.Decision.Pressure
					cur.Confidence = loop.Decision.Confidence
				}
				if pr, err := sess.client.GetPositions(runCtx); err == nil {
					cur.OpenPositions = len(pr.Positions)
				}
				srv.Set(cur)
			}
		}
	}()
	logger.Info("demo-week loop starting execution_mode=%s radar=SHADOW live=impossible", execMode)
	loop.Run(runCtx)
	logger.Info("demo-week stopped")
}
