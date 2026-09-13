package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"aurumflow/config"
	"aurumflow/internal/core"
	"aurumflow/internal/execacct"
	"aurumflow/internal/execution"
	"aurumflow/internal/exhaustion"
	"aurumflow/internal/gates"
	"aurumflow/internal/journal"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/money"
	"aurumflow/internal/ops"
	"aurumflow/internal/radar"
	"aurumflow/internal/optrust"
	"aurumflow/internal/recovery"
	"aurumflow/internal/research"
	"aurumflow/internal/strategy"
	"aurumflow/internal/strathist"
	"aurumflow/internal/stratrade"
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

	if !sess.Identity.MayTrade {
		logger.Warn("DEMO-WEEK account not explicitly verified — %s", execacct.Instruction(sess.Identity))
	}
	histReq := strathist.RequirementsFromConfig(sess.cfg)
	histStarted := time.Now().UTC()
	hist := strathist.Seed(ctx, sess.client, epic, histReq, histStarted)
	logger.Info("history seed status=%s M5=%d/%d oldest=%s latest=%s M15=%d/%d oldest=%s latest=%s H1=%d/%d oldest=%s latest=%s H4=%d/%d oldest=%s latest=%s warmup=%s",
		hist.Status,
		hist.M5Count, histReq.M5Required, rfcOrDash(hist.M5Oldest), rfcOrDash(hist.M5Latest),
		hist.M15Count, histReq.M15Required, rfcOrDash(hist.M15Oldest), rfcOrDash(hist.M15Latest),
		hist.H1Count, histReq.H1Required, rfcOrDash(hist.H1Oldest), rfcOrDash(hist.H1Latest),
		hist.H4Count, histReq.H4Required, rfcOrDash(hist.H4Oldest), rfcOrDash(hist.H4Latest),
		time.Since(histStarted).Round(time.Millisecond))
	gateErr := gates.DemoWeekTradeAllowed(gates.Input{
		Environment: "demo", Host: sess.client.BaseURL, ExecutionMode: string(config.ExecutionDemo),
		AccountOK: sess.Identity.MayTrade, KillSwitch: ks.HaltNewOrders(), MarketStatus: details.Snapshot.MarketStatus,
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

	week, werr := stratrade.LoadWeek(stratrade.WeekPath("journals", epic), epic, sess.cfg.Risk.MaxTrades)
	if werr != nil {
		logger.Warn("strategy week state: %v — starting empty", werr)
	}
	if week.BlocksNewStrategy() {
		logger.Warn("strategy week already completed %d/%d trades — no additional strategy opens", week.Snapshot().CompletedStrategyTrades, week.Snapshot().MaxTrades)
	}
	loop := core.NewLoop(sess.cfg, sess.client, epic, sess.acc.Balance.Balance, len(pos.Positions))
	loop.InstanceID = epic + "-demo-week"
	loop.RunID = uuid.New().String()
	loop.Kill = ks
	loop.Journal = jw
	loop.UnknownPositions = unknown
	loop.RequireRuntimeMoney = true
	loop.Money = cached
	loop.AccountID = sess.cfg.API.AccountID
	loop.ExecAccount = sess.Identity
	loop.History = &hist
	loop.HistoryReq = histReq
	loop.Week = week
	loop.GitCommit = gitHead()
	loop.StrategyVersion = "LEGACY"
	loop.Forensic = stratrade.NewRecorder(filepath.Join("research", "strategy-trades"), "")
	loop.ObserveIntel = func() stratrade.IntelligenceContext {
		return peekIntelligence("http://127.0.0.1:8766")
	}
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
				cur.LastExecution = loop.DecisionText()
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
					n := 0
					for _, p := range pr.Positions {
						if p.GetEpic() == epic {
							n++
							cur.GoldEntry = p.Position.Level
							cur.GoldUPnL = p.Position.ProfitLoss
							if p.Position.Upnl != 0 {
								cur.GoldUPnL = p.Position.Upnl
							}
							cur.GoldSL = p.Position.StopLevel
							cur.GoldTP = p.Position.ProfitLevel
						}
					}
					cur.OpenPositions = n
					cur.PositionsKnown = true
				}
				if md, err := sess.client.GetMarketDetails(runCtx, epic); err == nil && md != nil {
					cur.MarketStatus = md.Snapshot.MarketStatus
					cur.GoldBid = md.Snapshot.Bid
					cur.GoldAsk = md.Snapshot.Offer
					if cur.GoldAsk > cur.GoldBid {
						cur.GoldSpread = cur.GoldAsk - cur.GoldBid
					}
					cur.GoldValidation = cached.ValidationStatus
				}
				sessv := loop.SessionView(cur.MarketStatus)
				cur.StrategySession = sessv.StrategySession
				cur.SessionPolicy = sessv.ConfigPolicy
				cur.SessionEligible = sessv.Eligible
				cur.SessionReason = sessv.Reason
				cur.StrategyWaiting = sessv.StrategyWaiting
				cur.NextSession = sessv.NextSession
				if !sessv.NextSessionAt.IsZero() {
					cur.NextSessionAt = sessv.NextSessionAt.UTC().Format(time.RFC3339)
				}
				cur.AccountMasked = execacct.Mask(sess.Identity.AccountID)
				cur.AccountType = sess.Identity.Type
				cur.AccountCurrency = sess.Identity.Currency
				if sess.Identity.MayTrade {
					cur.AccountVerified = "VERIFIED"
				} else {
					cur.AccountVerified = sess.Identity.Resolution
				}
				if loop.History != nil {
					cur.HistoryStatus = loop.History.Status
					cur.HistoryM5 = fmt.Sprintf("%d/%d", loop.History.M5Count, histReq.M5Required)
					cur.HistoryM15 = fmt.Sprintf("%d/%d", loop.History.M15Count, histReq.M15Required)
					cur.HistoryH1 = fmt.Sprintf("%d/%d", loop.History.H1Count, histReq.H1Required)
					cur.HistoryH4 = fmt.Sprintf("%d/%d", loop.History.H4Count, histReq.H4Required)
					cur.HistoryReady = loop.History.Status == strathist.HistoryReady
				}
				cur.BrokerReady = cur.MarketStatus == "TRADEABLE"
				cur.AccountReady = sess.Identity.MayTrade
				cur.MonetaryReady = cached.ValidationStatus == money.RuntimeValidated
				cur.RiskReady = !ks.HaltNewOrders()
				cur.StrategyReady = cur.BrokerReady && cur.AccountReady && cur.MonetaryReady && cur.HistoryReady && sessv.Eligible && cur.RiskReady
				if !loop.LastScanAt.IsZero() {
					cur.LastScan = loop.LastScanAt.UTC().Format(time.RFC3339)
				}
				cur.LastSignal = loop.LastSignalID
				if !loop.LastSignalAt.IsZero() && cur.LastSignal == "" {
					cur.LastSignal = loop.LastSignalAt.UTC().Format(time.RFC3339)
				}
				wk := week.Snapshot()
				cur.TradeCount = wk.CompletedStrategyTrades
				cur.MaxTrades = wk.MaxTrades
				if cur.MaxTrades <= 0 {
					cur.MaxTrades = sess.cfg.Risk.MaxTrades
				}
				cur.MonetaryStatus = cached.ValidationStatus
				if sig := loop.OpenedSignal(); sig != nil {
					cur.GoldEntry = sig.Entry
					if cur.GoldSL == 0 {
						cur.GoldSL = sig.StopLoss
					}
					if cur.GoldTP == 0 {
						cur.GoldTP = sig.TakeProfit
					}
					stop := sig.Entry - sig.StopLoss
					if stop < 0 {
						stop = -stop
					}
					cur.ExpectedRisk = stratrade.ExpectedRiskUSD(0, stop, cached.MoneyPerPriceUnit)
					if !loop.OpenedAt().IsZero() {
						cur.HoldingSeconds = int64(time.Since(loop.OpenedAt()).Seconds())
					}
				}
				cur.ObservationalNote = stratrade.LabelObservational
				_, g := optrust.ForMarket(epic, true, true, cached.ValidationStatus == money.RuntimeValidated, cur.MarketStatus == "TRADEABLE")
				sc := stratrade.EmptyScorecard()
				if g.Monetary {
					sc.Monetary = stratrade.TrustPASS
				}
				if week.BlocksNewStrategy() {
					sc.Strategy = stratrade.TrustPASS
				}
				sc.Overall = stratrade.Overall(sc)
				cur.OperationalTrust = sc.Overall
				if halted, _ := loop.Forensic.Halted(); halted {
					cur.OperationalTrust = stratrade.TrustFAIL
				}
				srv.Set(cur)
			}
		}
	}()
	logger.Info("demo-week loop starting execution_mode=%s radar=SHADOW live=impossible", execMode)
	loop.Run(runCtx)
	logger.Info("demo-week stopped")
}

func peekIntelligence(base string) stratrade.IntelligenceContext {
	ctx := stratrade.NewObservational(stratrade.IntelligenceContext{DataQuality: "UNAVAILABLE"})
	client := &http.Client{Timeout: 1500 * time.Millisecond}
	resp, err := client.Get(strings.TrimRight(base, "/") + "/api/world")
	if err != nil {
		return ctx
	}
	defer resp.Body.Close()
	var raw map[string]any
	if json.NewDecoder(resp.Body).Decode(&raw) != nil {
		return ctx
	}
	ctx.WorldHash = asString(raw["Hash"])
	if ctx.WorldHash == "" {
		ctx.WorldHash = asString(raw["hash"])
	}
	if liq, ok := raw["Liquidity"].(map[string]any); ok {
		ctx.Liquidity = asString(liq["Class"])
	}
	if risk, ok := raw["Risk"].(map[string]any); ok {
		ctx.Risk = asString(risk["Class"])
	}
	ctx.DataQuality = "PEER_READ"
	return ctx
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func rfcOrDash(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.UTC().Format(time.RFC3339)
}
