package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/capitalhist"
	"aurumflow/internal/capsched"
	"aurumflow/internal/demomut"
	"aurumflow/internal/execacct"
	"aurumflow/internal/execution"
	"aurumflow/internal/indicators"
	"aurumflow/internal/killswitch"
	"aurumflow/internal/legacynorm"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/mirror"
	"aurumflow/internal/money"
	"aurumflow/internal/ops"
	"aurumflow/internal/pilotrisk"
	"aurumflow/internal/portfoliorisk"
	"aurumflow/internal/promote"
	"aurumflow/internal/research"
	"aurumflow/internal/strategy"
	"aurumflow/internal/stratrade"
	"aurumflow/pkg/models"
)

func runUS100MirrorOnly(ctx context.Context, statusAddr string, stay bool) {
	spec, err := money.LoadSpec(money.CachePath("", "US100"))
	if err != nil || !money.EvidenceComplete(spec) {
		logger.Error("US100 DEMO_MIRROR blocked: monetary evidence incomplete")
		return
	}
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDemo)
	if err != nil || !sess.Identity.MayTrade {
		logger.Error("US100 DEMO_MIRROR blocked: explicit DEMO account")
		return
	}
	runUS100Mirror(ctx, statusAddr, 3, stay, spec, sess.Identity)
}

func runUS100Mirror(ctx context.Context, statusAddr string, shadowMin int, stay bool, spec money.MonetaryInstrumentSpec, ident execacct.Identity) {
	if shadowMin <= 0 {
		shadowMin = 4
	}
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDemo)
	if err != nil {
		logger.Error("US100 mirror session: %v", err)
		return
	}
	ident = sess.Identity
	want := strings.TrimSpace(os.Getenv("AURUMFLOW_DEMO_ACCOUNT_ID"))
	coord := demomut.New()
	ks := killswitch.New(false, killswitch.DefaultFile)
	st := ops.NewStatus()
	st.APIEnvironment = "DEMO"
	st.ExecutionMode = "DEMO"
	st.CapitalEnv = "DEMO"
	st.LivePossible = "IMPOSSIBLE / FAIL-CLOSED"
	st.AccountMasked = execacct.Mask(ident.AccountID)
	st.AccountVerified = verifiedLabel(ident)
	st.AccountCurrency = ident.Currency
	st.US100Origin = demomut.OriginMirror
	st.US100Status = "SHADOW"
	st.US100DemoEligible = "YES"
	st.US100MirrorArmed = "ON"
	st.US100Shadow = "RUNNING"
	st.MonetaryStatus = spec.ValidationStatus
	srv := ops.NewServer(statusAddr, st)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Warn("US100 ops: %v", err)
		}
	}()
	eq := sess.acc.Balance.Balance
	logger.Info("US100 live SHADOW then DEMO_MIRROR · port=%s · strategy=%s · account=%s · equity=%.2f pilot_pos=%.2f pilot_agg=%.2f · GOLD :8765 untouched",
		statusAddr, promote.FrozenStrategy, execacct.Mask(ident.AccountID), eq, pilotrisk.PerPositionCap(eq), pilotrisk.EffectiveAggregate(eq, pilotrisk.HardAggregateUSD))

	diskM15 := loadP89Candles(filepath.Join("data", "research", "p89", "capital"), "US100", "MINUTE_15")
	diskH1 := loadP89Candles(filepath.Join("data", "research", "p89", "capital"), "US100", "HOUR")
	diskH4 := loadP89Candles(filepath.Join("data", "research", "p89", "capital"), "US100", "HOUR_4")
	sched := capsched.New(2, 300*time.Millisecond)
	seen := map[string]bool{}
	week, _ := stratrade.LoadWeek(stratrade.WeekPath("journals", "US100"), "US100", pilotrisk.MaxTradesMarket)
	journalOK := writeUS100Shadow(map[string]any{"event": "SHADOW_START", "t": time.Now().UTC(), "strategy": promote.FrozenStrategy, "hash": legacynorm.SpecHash, "pilot": pilotrisk.Policy})
	armedAfter := time.Now().UTC()
	proveUntil := time.Now().Add(time.Duration(shadowMin) * time.Minute)
	scans := 0
	var rec *stratrade.Recorder
	halt := ks.HaltNewOrders()
	st.MaxTrades = week.MaxTrades
	st.TradeCount = week.CompletedStrategyTrades

	tick := func() {
		now := time.Now().UTC()
		md, quotesOK, mstat := us100Details(ctx, sched, sess.client)
		m15 := mergeUS100(ctx, sched, sess.client, diskM15, market.ResolutionMinute15)
		h1 := mergeUS100(ctx, sched, sess.client, diskH1, market.ResolutionHour)
		h4 := mergeUS100(ctx, sched, sess.client, diskH4, market.ResolutionHour4)
		histReady := len(m15) >= 80 && len(h1) >= 20 && len(h4) >= 8
		age := 0.0
		if len(m15) > 0 {
			age = now.Sub(m15[len(m15)-1].Time).Minutes()
		}
		fresh := quotesOK && (mstat == "TRADEABLE" && age < 45 || mstat != "TRADEABLE")
		sessEv := strategy.EvaluateSession(now, legacynorm.SessionsOf("US100"))
		nextName, nextAt := strategy.NextEligibleLabel(sessEv, now, legacynorm.SessionsOf("US100"))
		sigs := []research.SignalRow{}
		if histReady {
			sigs = legacynorm.ScanNormalized(m15, h1, h4)
		}
		scans++
		lastSig := ""
		if n := len(sigs); n > 0 {
			lastSig = sigs[n-1].Time.UTC().Format(time.RFC3339) + " dir=" + itoa(sigs[n-1].Direction)
		}
		pos, _ := sess.client.GetPositions(ctx)
		goldN := countEpic(pos, "GOLD")
		usN := countEpic(pos, "US100")
		if rec != nil && usN > 0 {
			sampleUS100(rec, spec, pos, md)
		}
		if rec != nil && usN == 0 && st.US100Lifecycle != "" && st.US100Lifecycle != "CLOSED" {
			if mirror.HaltOnMismatch(0, usN, 0) && st.OpenPositions > 0 {
				// still open locally? treat as closed
			}
			_ = rec.Append(stratrade.Event{Event: stratrade.EvPositionClosed, Note: "broker US100=0"})
			_ = rec.Append(stratrade.Event{Event: stratrade.EvFinalReconciliation})
			if err := mirror.Reconcile(0, usN, 0); err != nil {
				halt = true
				_ = ks.HaltPersist()
				logger.Error("US100 CLOSE mismatch — HALT_NEW_ORDERS")
			}
			st.US100Lifecycle = "CLOSED"
			_ = week.MarkClosed(rec.ID())
			st.TradeCount = week.CompletedStrategyTrades
			rec = nil
		}
		st.HistoryReady = histReady
		st.HistoryStatus = "M15=" + itoa(len(m15))
		st.LiveDataFresh = fresh
		st.M15AgeMinutes = age
		st.SessionEligible = sessEv.Eligible
		st.SessionReason = sessEv.Reason
		st.StrategySession = sessEv.ClockSession
		st.SessionPolicy = sessEv.ConfigPolicy
		st.NextSession = nextName
		if !nextAt.IsZero() {
			st.NextSessionAt = nextAt.UTC().Format(time.RFC3339)
		}
		st.TradeCount = week.CompletedStrategyTrades
		st.MaxTrades = week.MaxTrades
		if week.BlocksNewStrategy() {
			st.US100MirrorArmed = "COMPLETE"
		}
		st.StrategyReady = histReady && journalOK
		st.LastScan = now.Format(time.RFC3339)
		st.LastSignal = lastSig
		st.US100Status = "SHADOW"
		if st.US100MirrorArmed == "ON" {
			st.US100Status = "DEMO_ELIGIBLE"
		}
		st.OpenPositions = goldN + usN
		st.PositionsKnown = true
		st.US100RiskUsed = 0
		if usN > 0 {
			st.US100Status = "OPEN"
			fillUS100Pos(&st, pos)
		}
		if md != nil {
			st.US100Current = (md.Snapshot.Bid + md.Snapshot.Offer) / 2
		}
		st.HaltNewOrders = halt || ks.HaltNewOrders()
		st.PortfolioOpen = goldN + usN
		st.PreciousGroup = goldN
		st.USEquityGroup = usN
		srv.Set(st)
		_ = writeUS100Shadow(map[string]any{
			"event": "SCAN", "t": now, "history_ready": histReady, "fresh": fresh,
			"session_eligible": sessEv.Eligible, "session": sessEv.ClockSession,
			"signals": len(sigs), "last": lastSig, "broker_mutation": false,
			"journal": journalOK, "gold": goldN, "us100": usN,
		})
		if scans == 1 {
			logger.Info("US100 SHADOW history_ready=%v fresh=%v session=%s eligible=%v scan=RUNNING journal=%v mutation=0",
				histReady, fresh, sessEv.ClockSession, sessEv.Eligible, journalOK)
		}
		if st.US100MirrorArmed == "ON" && rec == nil && usN == 0 && !halt && !ks.HaltNewOrders() && !week.BlocksNewStrategy() {
			if cand, ok := latestLiveSignal(sigs, armedAfter, sessEv.Eligible); ok {
				tryUS100Open(ctx, sess, coord, ks, spec, ident, want, cand, md, m15, pos, goldN, seen, week, &rec, &st, &halt)
				srv.Set(st)
			}
		}
	}

	tick()
	tkr := time.NewTicker(45 * time.Second)
	defer tkr.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tkr.C:
			tick()
			if !stay && time.Now().After(proveUntil) && rec == nil {
				logger.Info("US100 SHADOW prove complete scans=%d · no fake signal · DEMO_MIRROR armed=%s", scans, st.US100MirrorArmed)
				return
			}
		}
	}
}

func us100Details(ctx context.Context, sched *capsched.Scheduler, client *market.Client) (*market.MarketDetailsResponse, bool, string) {
	var md *market.MarketDetailsResponse
	_ = sched.Do(ctx, func(ctx context.Context) error {
		var err error
		md, err = client.GetMarketDetails(ctx, "US100")
		return err
	})
	if md == nil {
		return nil, false, ""
	}
	return md, md.Snapshot.Bid > 0 && md.Snapshot.Offer > 0, md.Snapshot.MarketStatus
}

func mergeUS100(ctx context.Context, sched *capsched.Scheduler, client *market.Client, disk []models.Candle, res string) []models.Candle {
	var live []models.Candle
	_ = sched.Do(ctx, func(ctx context.Context) error {
		got, err := client.GetPrices(ctx, "US100", res, 80, time.Time{}, time.Time{})
		live = got
		return err
	})
	return capitalhist.DedupSort(append(append([]models.Candle{}, disk...), live...))
}

func latestLiveSignal(sigs []research.SignalRow, after time.Time, sessionOK bool) (research.SignalRow, bool) {
	if !sessionOK {
		return research.SignalRow{}, false
	}
	for i := len(sigs) - 1; i >= 0; i-- {
		if sigs[i].Time.After(after) && sigs[i].Direction != 0 {
			return sigs[i], true
		}
	}
	return research.SignalRow{}, false
}

func tryUS100Open(ctx context.Context, sess *demoSession, coord *demomut.Coordinator, ks *killswitch.Switch, spec money.MonetaryInstrumentSpec, ident execacct.Identity, want string, sig research.SignalRow, md *market.MarketDetailsResponse, m15 []models.Candle, pos *market.PositionsResponse, goldN int, seen map[string]bool, week *stratrade.WeekState, rec **stratrade.Recorder, st *ops.Status, halt *bool) {
	if err := mirror.AllowOrder(mirror.Intent{Source: mirror.SourceNormalized, Signal: sig}); err != nil {
		return
	}
	dir := "BUY"
	if sig.Direction < 0 {
		dir = "SELL"
	}
	sid := stratrade.SignalID("US100", sig.Time, dir)
	if err := mirror.OneSignalOneOrder(seen, sid); err != nil {
		return
	}
	if md == nil || !marketAllowsExecution(md.Snapshot.MarketStatus) {
		logger.Info("US100 signal %s held: market %s", sid, "")
		return
	}
	if !money.EvidenceComplete(spec) || week.BlocksNewStrategy() {
		return
	}
	idx := len(m15) - 1
	if idx < 20 {
		return
	}
	atr := indicators.ATR(m15[:idx+1], 14)
	if atr <= 0 {
		return
	}
	entry := (md.Snapshot.Bid + md.Snapshot.Offer) / 2
	if entry <= 0 {
		entry = m15[idx].Close
	}
	sl := entry - float64(sig.Direction)*atr
	tp := entry + float64(sig.Direction)*atr*1.5
	stopDist := atr
	if ar, aerr := sess.client.GetAccounts(ctx); aerr == nil {
		for _, a := range ar.Accounts {
			if a.AccountID == ident.AccountID && a.Balance.Balance > 0 {
				sess.acc = a
				break
			}
		}
	}
	bal := sess.acc.Balance.Balance
	if bal <= 0 {
		logger.Info("US100 signal %s risk reject: unknown DEMO balance", sid)
		return
	}
	goldRisk := goldOpenRisk(pos)
	openRisk := 0.0
	if goldRisk.RiskKnown {
		openRisk = goldRisk.RiskMoney
	} else if goldN > 0 {
		logger.Info("US100 signal %s blocked: GOLD risk unknown", sid)
		return
	}
	pd := pilotrisk.Evaluate(pilotrisk.Input{
		Market: "US100", Equity: bal, StopDistance: stopDist, MPU: spec.MoneyPerPriceUnit,
		MinDealSize: spec.MinDealSize, SizeStep: spec.SizeIncrement, MaxDealSize: spec.MaxDealSize,
		OpenRisk: openRisk, TradesThisMarket: week.CompletedStrategyTrades, OpenStrategy: goldN + countEpic(pos, "US100"),
		GroupOpen: map[string]int{portfoliorisk.GroupPrecious: goldN, portfoliorisk.GroupUSEquity: countEpic(pos, "US100")},
	})
	if !pd.Pass {
		logger.Info("US100 signal %s %s pilot=%s", sid, pd.Block, pilotrisk.Policy)
		return
	}
	size := pd.Size
	expRisk := pd.PlannedRisk
	pm := portfoliorisk.New()
	pm.Caps.MaxOpenPositions = pilotrisk.MaxStrategyOpen
	var open []portfoliorisk.Position
	if goldN > 0 {
		open = append(open, goldRisk)
	}
	pr := pm.Evaluate(open, portfoliorisk.Position{Canonical: "US100", Region: "UNITED_STATES", Asset: "EQUITIES", RiskMoney: expRisk, RiskKnown: true})
	if !pr.Pass {
		logger.Info("US100 signal %s portfolio reject: %v", sid, pr.Blocks)
		return
	}
	req := demomut.Request{
		Origin: demomut.OriginMirror, Market: "US100", Env: "DEMO", Account: ident, WantAccountID: want,
		Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true,
		Halt: ks.HaltNewOrders() || *halt, OpenStrategy: goldN + countEpic(pos, "US100"),
		GroupOpen: map[string]int{portfoliorisk.GroupPrecious: goldN, portfoliorisk.GroupUSEquity: countEpic(pos, "US100")},
	}
	if err := coord.Reserve(req); err != nil {
		logger.Info("US100 signal %s coordinator: %v", sid, err)
		return
	}
	defer coord.Release()
	seen[sid] = true
	r := stratrade.NewRecorder(filepath.Join("research", "strategy-trades"), sid)
	*rec = r
	mid := (md.Snapshot.Bid + md.Snapshot.Offer) / 2
	_ = r.PersistPreSignal(stratrade.PreSignalSnapshot{
		SignalID: sid, Timestamp: time.Now().UTC(), StrategyVersion: promote.FrozenStrategy, StrategyHash: legacynorm.SpecHash,
		Origin: demomut.OriginMirror, Instrument: "US100", Direction: dir,
		Bid: md.Snapshot.Bid, Ask: md.Snapshot.Offer, Mid: mid, Spread: md.Snapshot.Offer - md.Snapshot.Bid,
		MarketStatus: md.Snapshot.MarketStatus, ATR: atr, Session: "NY", HistoryReady: true,
		EntryCandidate: entry, StopLoss: sl, TakeProfit: tp,
		StopDistance: stopDist, PositionSize: size, MoneyPerPriceUnit: spec.MoneyPerPriceUnit,
		ExpectedAccountRisk: expRisk, AccountBalance: bal, Immutable: true,
		ObservationalNote: stratrade.LabelObservational, PilotPolicy: pilotrisk.Policy,
		PilotRiskLimit: pd.PerPosCap, PortfolioRiskBefore: openRisk,
	})
	_ = r.AttachIntel(stratrade.IntelligenceContext{WorldHash: "", Attention: 0, Coverage: "OBSERVATIONAL", ObservationalOnly: true})
	_ = r.Append(stratrade.Event{Event: stratrade.EvSignalObserved, SignalID: sid})
	_ = r.Append(stratrade.Event{Event: stratrade.EvRiskEvaluated, Note: "account_currency_risk"})
	_ = r.Append(stratrade.Event{Event: stratrade.EvRiskAccepted})
	_ = r.Append(stratrade.Event{Event: stratrade.EvOrderIntent})
	ex := execution.NewExecutor(sess.client, "US100", spec.MinDealSize, spec.SizeIncrement)
	_ = r.Append(stratrade.Event{Event: stratrade.EvBrokerOpenRequest})
	ref, err := ex.SendOrder(ctx, &models.TradeSignal{Direction: dir, Entry: entry, StopLoss: sl, TakeProfit: tp, Score: sig.Score}, size)
	if err != nil {
		logger.Error("US100 OPEN fail: %v", err)
		return
	}
	conf, err := ex.ConfirmDeal(ctx, ref)
	if err != nil || conf == nil {
		logger.Error("US100 CONFIRM fail: %v", err)
		return
	}
	r.Bind(sid, ref, conf.DealID)
	_ = r.Append(stratrade.Event{Event: stratrade.EvBrokerConfirm, DealReference: ref, DealID: conf.DealID})
	_ = r.Append(stratrade.Event{Event: stratrade.EvPositionResolved})
	gotEpic, gotDir, gotSize, gotSL, gotTP := conf.Epic, conf.Direction, conf.Size, 0.0, 0.0
	if fresh, ferr := sess.client.GetPositions(ctx); ferr == nil {
		for _, p := range fresh.Positions {
			if strings.EqualFold(p.GetEpic(), "US100") {
				gotEpic, gotDir, gotSize = p.GetEpic(), p.Position.Direction, p.Position.Size
				gotSL, gotTP = p.Position.StopLevel, p.Position.ProfitLevel
			}
		}
	}
	if !mirror.ProtectionOK("US100", gotEpic, dir, gotDir, size, gotSize, sl, gotSL, tp, gotTP, atr*0.15) {
		*halt = true
		_ = ks.HaltPersist()
		logger.Error("US100 protection mismatch — HALT_NEW_ORDERS")
		_ = r.Append(stratrade.Event{Event: stratrade.EvProtectiveVerified, Severity: "HALT", Note: "SL/TP or identity mismatch"})
		return
	}
	_ = r.Append(stratrade.Event{Event: stratrade.EvPositionReconciled})
	_ = r.Append(stratrade.Event{Event: stratrade.EvProtectiveVerified})
	_ = r.Append(stratrade.Event{Event: stratrade.EvMonitoring})
	_ = week.MarkOpen(sid, sid)
	fill := conf.Level
	if fill == 0 {
		fill = entry
	}
	actualStop := fill - sl
	if actualStop < 0 {
		actualStop = -actualStop
	}
	actualRisk := pilotrisk.ModeledRisk(gotSize, actualStop, spec.MoneyPerPriceUnit)
	_ = stratrade.WriteRisk(r.Dir(), "", stratrade.RiskRecord{
		Policy: pilotrisk.Policy, Equity: bal, PerPositionCap: pd.PerPosCap, AggregateCap: pd.AggregateCap,
		PlannedRisk: expRisk, ActualModeledRisk: actualRisk, Difference: actualRisk - expRisk,
		StopDistance: actualStop, Size: gotSize, MoneyPerPriceUnit: spec.MoneyPerPriceUnit,
		PilotStillHolds: actualRisk <= pd.PerPosCap+1e-6,
	})
	slip := mirror.EntrySlippage(dir, md.Snapshot.Bid, md.Snapshot.Offer, fill)
	st.US100Status = "OPEN"
	st.US100Direction = dir
	st.US100Entry = conf.Level
	if st.US100Entry == 0 {
		st.US100Entry = entry
	}
	st.US100SL = sl
	st.US100TP = tp
	st.US100Size = conf.Size
	if st.US100Size == 0 {
		st.US100Size = size
	}
	st.US100Risk = expRisk
	st.US100Lifecycle = "MONITORING"
	st.US100DealRef = ref
	st.US100DealID = conf.DealID
	st.ExpectedRisk = expRisk
	logger.Info("US100 DEMO_MIRROR OPEN dir=%s size=%.4f entry=%.2f sl=%.2f tp=%.2f planned=%.4f actual=%.4f slip=%.4f ref=%s",
		dir, st.US100Size, st.US100Entry, sl, tp, expRisk, actualRisk, slip, ref)
	_ = writeUS100Shadow(map[string]any{
		"event": "DEMO_MIRROR_OPEN", "origin": demomut.OriginMirror, "market": "US100",
		"strategy": promote.FrozenStrategy, "signal_id": sid, "dealReference": ref, "dealId": conf.DealID,
		"direction": dir, "size": st.US100Size, "entry": st.US100Entry, "sl": sl, "tp": tp, "risk": expRisk,
	})
}

func fillUS100Pos(st *ops.Status, pos *market.PositionsResponse) {
	if pos == nil {
		return
	}
	for _, p := range pos.Positions {
		if !strings.EqualFold(p.GetEpic(), "US100") {
			continue
		}
		st.US100Direction = p.Position.Direction
		st.US100Entry = p.Position.Level
		st.US100SL = p.Position.StopLevel
		st.US100TP = p.Position.ProfitLevel
		st.US100Size = p.Position.Size
		st.US100UPL = money.BrokerUPL(p.Position.Upnl, p.Position.ProfitLoss)
		st.US100DealID = p.Position.DealID
	}
}

func sampleUS100(rec *stratrade.Recorder, spec money.MonetaryInstrumentSpec, pos *market.PositionsResponse, md *market.MarketDetailsResponse) {
	if pos == nil {
		return
	}
	for _, p := range pos.Positions {
		if !strings.EqualFold(p.GetEpic(), "US100") {
			continue
		}
		cur := p.Position.Level
		if md != nil {
			if strings.EqualFold(p.Position.Direction, "SELL") {
				cur = md.Snapshot.Offer
			} else if md.Snapshot.Bid > 0 {
				cur = md.Snapshot.Bid
			}
		}
		exp, _ := money.ExpectedPnLDirected(spec, p.Position.Size, p.Position.Level, cur, p.Position.Direction)
		_ = rec.Sample(stratrade.Heartbeat{
			Bid: cur, BrokerUPL: money.BrokerUPL(p.Position.Upnl, p.Position.ProfitLoss),
			ExpectedUPL: exp, SL: p.Position.StopLevel, TP: p.Position.ProfitLevel, PositionState: "OPEN",
		})
	}
}

func goldOpenRisk(pos *market.PositionsResponse) portfoliorisk.Position {
	p := portfoliorisk.Position{Canonical: "GOLD", Region: "GLOBAL", Asset: "PRECIOUS_METALS"}
	if pos == nil {
		return p
	}
	gold, err := money.LoadSpec(money.CachePath("", "GOLD"))
	if err != nil || gold.MoneyPerPriceUnit <= 0 {
		return p
	}
	for _, it := range pos.Positions {
		if !strings.EqualFold(it.GetEpic(), "GOLD") {
			continue
		}
		dist := it.Position.Level - it.Position.StopLevel
		if dist < 0 {
			dist = -dist
		}
		if dist <= 0 || it.Position.Size <= 0 {
			return p
		}
		p.RiskMoney = it.Position.Size * dist * gold.MoneyPerPriceUnit
		p.RiskKnown = true
		return p
	}
	return p
}

func writeUS100Shadow(v any) bool {
	path := filepath.Join("journals", "us100-shadow.jsonl")
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(v) == nil
}
