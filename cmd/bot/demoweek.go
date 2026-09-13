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
	"aurumflow/internal/core"
	"aurumflow/internal/execution"
	"aurumflow/internal/exhaustion"
	"aurumflow/internal/gates"
	"aurumflow/internal/journal"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/money"
	"aurumflow/internal/ops"
	"aurumflow/internal/radar"
	"aurumflow/internal/recovery"
	"aurumflow/internal/research"
	"aurumflow/internal/strategy"
	"aurumflow/pkg/models"
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
	st.BookCapability = "BOOK_CAPABILITY_LIMITED"
	st.FeedFreshness = "CAPITAL_REST"
	started := time.Now().UTC()
	exh := exhaustion.NewEngine(epic)
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
	loop.OnLegacySignal = func(sig *models.TradeSignal, now time.Time) {
		if sig == nil || exhaustion.MayMutateBroker() || radar.MayMutateBroker(radar.ModeShadow) {
			return
		}
		dir := 1
		if sig.Direction == "SELL" {
			dir = -1
		}
		p := 0.0
		if loop.Decision != nil {
			p = loop.Decision.Pressure
		}
		snap := exh.Observe(now, dir, float64(sig.Score), p, 0)
		if loop.Decision != nil {
			loop.Decision.V1Class = snap.Classification
			loop.Decision.FlowEfficiency = snap.Features.FlowEffNorm
			loop.Decision.ImpactFailure = snap.Features.ImpactFailure
			loop.Decision.ExhaustionEvidence = snap.ExhaustionEvidence
		}
		if loop.Journal != nil {
			_ = loop.Journal.WriteLifecycle(journal.Lifecycle{
				Event: journal.EventExhaustionSnapshot, Epic: epic, Direction: sig.Direction, Score: sig.Score,
				Reason: snap.Classification, Status: snap.Mode,
			})
			if snap.Classification == exhaustion.ClassExhaustion {
				_ = loop.Journal.WriteLifecycle(journal.Lifecycle{
					Event: journal.EventFlowExhaustionSignal, Epic: epic, Direction: sig.Direction, Score: sig.Score,
					SignalID: now.UTC().Format(time.RFC3339Nano) + "-" + epic,
				})
			}
		}
		if !strings.Contains(strings.ToUpper(epic), "BTC") {
			return
		}
		id := now.UTC().Format("20060102T150405Z") + "-" + epic + "-" + snap.Classification
		_ = research.AppendInput(research.ProspectiveDir, research.ProspectiveInput{
			SignalID: id, RecordedAt: time.Now().UTC(), Timestamp: now.UTC(), Instrument: "BTCUSDT",
			LegacyDirection: dir, LegacyScore: sig.Score, PressureScore: p,
			DirectionalPressure: snap.DirectionalPressure, V1Classification: snap.Classification,
			Features: snap.Features, GitCommit: gitHead(), FeatureVersion: exhaustion.FeatureVersion,
			SpecHash: exhaustion.V1SpecHash, OutcomeKnown: false,
		})
		if loop.Journal != nil {
			_ = loop.Journal.WriteLifecycle(journal.Lifecycle{Event: journal.EventProspectiveInput, Epic: epic, SignalID: id})
		}
	}
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
					cur.LastPressure = loop.Decision.Pressure
					cur.Confidence = loop.Decision.Confidence
					cur.LastV1Class = loop.Decision.V1Class
					cur.FlowEfficiency = loop.Decision.FlowEfficiency
					cur.ImpactFailure = loop.Decision.ImpactFailure
				}
				if loop.LastSignalText == "BUY" {
					cur.LastLegacyDir = 1
				} else if loop.LastSignalText == "SELL" {
					cur.LastLegacyDir = -1
				}
				cur.BookCapability = "BOOK_CAPABILITY_LIMITED"
				pst := research.ReadProspectiveStatus(research.ProspectiveDir)
				cur.ProspectiveTotal = pst.Signals
				cur.ProspectiveExh = pst.Exhaustion
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
