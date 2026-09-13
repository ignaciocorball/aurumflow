package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/demoexec"
	"aurumflow/internal/demoprep"
	"aurumflow/internal/eligibility"
	"aurumflow/internal/instrument"
	"aurumflow/internal/livesurface"
	"aurumflow/internal/logger"
	"aurumflow/internal/money"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/orchestrator"
	"aurumflow/internal/portfoliorisk"
	"aurumflow/internal/researchopp"
	"aurumflow/internal/scanner"
	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

func runGlobalShadow(ctx context.Context) {
	now := time.Now().UTC()
	in, res := worldstate.LoadOfficialInput(ctx, now)
	ws := worldstate.At(now, in)
	var maps []instrument.Mapping
	if demoEnvConfigured() {
		maps = discoverCapitalMarkets(ctx)
		if frame, ok := collectCapitalFrame(ctx, maps, now); ok {
			ws = worldstate.ApplyLiveFrame(ws, frame)
		}
	} else {
		logger.Info("global-shadow: no DEMO credentials — analysis-only, no broker discovery")
	}
	elig := eligibility.New()
	for _, m := range maps {
		if m.Unresolved || m.Epic == "" {
			elig.Set(eligibility.Entry{Canonical: m.Canonical, Status: worlddomain.EligAnalysis, Reason: "ANALYSIS_ONLY"})
			continue
		}
		elig.Set(eligibility.Advance(eligibility.Entry{Canonical: m.Canonical, Epic: m.Epic, DemoHost: true}, money.MonetaryInstrumentSpec{Epic: m.Epic}, false, false))
	}
	ws = elig.ApplyToWorld(ws)
	ranks := opportunity.Rank(ws)
	ws = scanner.AttachOpportunity(ws, ranks)
	orch := orchestrator.New()
	risk := portfoliorisk.New()
	demo := demoexec.New()
	var proposals []orchestrator.DecisionProposal
	watch, setup, cand, blocked := 0, 0, 0, 0
	prev := map[string]*researchopp.Signal{}
	for _, r := range ranks {
		e := elig.Get(r.Market)
		pr := risk.Evaluate(nil, portfoliorisk.Position{Canonical: r.Market, Region: string(r.State.Region), Asset: string(r.State.AssetClass), RiskKnown: false})
		p := orch.ProposeWith(ws, r, "", "", e.Status, pr.Blocks, orchestrator.Extra{WorldHash: ws.Hash})
		if p.WorldHash == "" {
			logger.Error("global-shadow world_hash empty")
		}
		proposals = append(proposals, p)
		switch p.Decision {
		case "SETUP":
			setup++
		case "DEMO_CANDIDATE":
			cand++
		case "BLOCKED":
			blocked++
		default:
			watch++
		}
		sig := researchopp.Stamp(researchopp.Signal{
			ID: r.Market + "-" + now.Format("20060102150405"), T0: now, WorldHash: ws.Hash,
			Proposal: jsonBytes(p), Ranking: jsonBytes(r), MarketState: jsonBytes(r.State),
			Tier: string(r.Tier), Eligibility: string(e.Status),
		})
		if researchopp.ShouldRecord(prev[r.Market], sig) {
			_ = researchopp.Record("journals", sig)
			cp := sig
			prev[r.Market] = &cp
		}
		acc := demo.Accept(p, false)
		if acc.Accepted {
			logger.Error("global-shadow must never accept execution")
		}
	}
	rep := map[string]any{
		"as_of": ws.AsOf, "valid": ws.Valid, "hash": ws.Hash, "fetch": res,
		"markets_analyzed": len(proposals), "watch": watch, "setup": setup,
		"demo_candidate": cand, "blocked": blocked, "proposals": proposals,
		"broker_mutation": orch.CanMutateBroker(),
		"multi_market_demo": demo.Enabled,
		"risk_unit": portfoliorisk.RiskUnit,
	}
	_ = worldstate.WriteJSON(filepath.Join("research/reports", "MULTI_MARKET_SHADOW.json"), rep)
	_ = worldstate.WriteMD(filepath.Join("research/reports", "MULTI_MARKET_SHADOW.md"), shadowMD(rep, proposals))
	writeLiveOpportunity(ws, ranks, proposals, elig)
	logger.Info("global-shadow valid=%s hash=%s analyzed=%d watch=%d setup=%d candidate=%d blocked=%d mutate=%v multi_demo=%v",
		ws.Valid, ws.Hash, len(proposals), watch, setup, cand, blocked, orch.CanMutateBroker(), demo.Enabled)
}

func discoverCapitalMarkets(ctx context.Context) []instrument.Mapping {
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		logger.Warn("global-shadow discovery: %v", err)
		return nil
	}
	var hits []instrument.SearchHit
	for _, term := range instrument.SearchTerms() {
		ms, err := sess.client.SearchMarkets(ctx, term)
		if err != nil {
			continue
		}
		for _, m := range ms {
			hits = append(hits, instrument.SearchHit{
				Epic: m.Epic, Name: m.InstrumentName, Status: m.MarketStatus,
				InstrumentType: m.InstrumentType,
			})
			logger.Info("DISCOVER %s epic=%s type=%s status=%s", term, m.Epic, m.InstrumentType, m.MarketStatus)
		}
	}
	maps := instrument.MapHits(hits)
	_ = instrument.PersistMappings("data/world/normalized/capital/mappings.json", maps)
	return maps
}

func collectCapitalFrame(ctx context.Context, maps []instrument.Mapping, now time.Time) (livesurface.Frame, bool) {
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		return livesurface.Frame{}, false
	}
	quotes := map[string]livesurface.Quote{}
	hist := map[string][]livesurface.Candle{}
	recv := time.Now().UTC()
	for _, m := range maps {
		if m.Epic == "" || m.Unresolved {
			continue
		}
		d, err := sess.client.GetMarketDetails(ctx, m.Epic)
		if err != nil {
			continue
		}
		event := now
		q := livesurface.NewQuote(m.Canonical, m.Epic, d.Snapshot.Bid, d.Snapshot.Offer, d.Snapshot.MarketStatus, event, recv, 2*time.Minute)
		quotes[m.Canonical] = q
		if h := historyOf(m.Canonical); len(h.M5Bars) > 0 {
			hist[m.Canonical] = h.M5Bars
		} else if cs, err := sess.client.GetPrices(ctx, m.Epic, "MINUTE_5", 80, time.Time{}, time.Time{}); err == nil {
			var candles []livesurface.Candle
			for _, c := range cs {
				candles = append(candles, livesurface.Candle{Time: c.Time, Close: c.Close, High: c.High, Low: c.Low})
			}
			hist[m.Canonical] = candles
		}
	}
	if len(quotes) == 0 {
		return livesurface.Frame{}, false
	}
	return livesurface.Assemble(now, quotes, hist), true
}

func runPrepareMarket(ctx context.Context, canonical string, calibrate bool) {
	canonical = strings.ToUpper(strings.TrimSpace(canonical))
	if canonical == "" {
		logger.Error("prepare-market requires --prepare-market <canonical>")
		os.Exit(1)
	}
	epic := canonical
	if maps, err := instrument.LoadMappings("data/world/normalized/capital/mappings.json"); err == nil {
		for _, m := range maps {
			if strings.EqualFold(m.Canonical, canonical) && m.Epic != "" {
				epic = m.Epic
				break
			}
		}
	}
	if calibrate {
		logger.Info("CALIBRATE MARKET canonical=%s epic=%s DEMO-only explicit", canonical, epic)
	} else {
		logger.Info("PREPARE MARKET canonical=%s epic=%s READ-ONLY", canonical, epic)
	}
	runPrepareInstrument(ctx, epic, calibrate)
}

func runPrepareDemoUniverse(ctx context.Context) {
	maps := []instrument.Mapping{}
	if demoEnvConfigured() {
		maps = discoverCapitalMarkets(ctx)
	} else if loaded, err := instrument.LoadMappings("data/world/normalized/capital/mappings.json"); err == nil {
		maps = loaded
		logger.Info("prepare-demo-universe: using cached mappings (no DEMO session)")
	}
	rows := demoprep.UniverseRows(maps)
	if demoEnvConfigured() {
		if sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled); err == nil {
			for i, r := range rows {
				if r.Epic == "" {
					continue
				}
				d, err := sess.client.GetMarketDetails(ctx, r.Epic)
				if err != nil {
					continue
				}
				spec := money.FromMarketDetails(d)
				rows[i].Tradeable = strings.EqualFold(d.Snapshot.MarketStatus, "TRADEABLE")
				rows[i].MarketStatus = d.Snapshot.MarketStatus
				rows[i].MinSize = spec.MinDealSize
				rows[i].Increment = spec.SizeIncrement
				rows[i].Currency = spec.Currency
				rows[i].MonetaryMeta = spec.ValidationStatus
				rows[i].MoneyConfidence = spec.ValidationStatus
				rows[i].SpecValid = spec.DealingComplete() && spec.ValidationStatus != money.Unverified
				if spec.MoneyPerPriceUnit <= 0 {
					rows[i].MoneyConfidence = "UNKNOWN"
				}
				rows[i].StopRules = "UNKNOWN"
				if d.DealingRules.MinStopOrProfitDistance.Value > 0 {
					rows[i].StopRules = "BROKER_MIN_STOP_PRESENT"
				}
				rows[i].Eligibility = eligibility.Advance(eligibility.Entry{Canonical: r.Market, Epic: r.Epic, DemoHost: true, Tradeable: rows[i].Tradeable}, spec, false, false).Status
				rows[i].ReadyToCalibrate = rows[i].SpecValid && rows[i].Tradeable
			}
		}
	}
	now := time.Now().UTC()
	in, _ := worldstate.LoadOfficialInput(ctx, now)
	ws := worldstate.At(now, in)
	ranks := opportunity.Rank(ws)
	q := demoprep.Queue(rows, ranks)
	ready := demoprep.ReadyNames(q)
	logger.Info("PREPARE DEMO UNIVERSE READ-ONLY · no orders · MULTI_MARKET_DEMO_EXECUTION=OFF")
	for _, r := range rows {
		logger.Info("MATRIX %s discovered=%v tradeable=%v spec=%v money=%s calibrated=%v eligible=%v status=%s",
			r.Market, r.Discovered, r.Tradeable, r.SpecValid, r.MoneyConfidence, r.RuntimeCalibrated, r.DemoEligible, r.MarketStatus)
	}
	logger.Info("READY_TO_CALIBRATE")
	if len(ready) == 0 {
		logger.Info("(none — gates closed or markets not TRADEABLE)")
	}
	for _, n := range ready {
		logger.Info("%s", n)
	}
	writeReadiness(rows, ranks)
	_ = worldstate.WriteJSON(filepath.Join("research/reports", "DEMO_PREPARE_UNIVERSE.json"), map[string]any{
		"rows": rows, "queue": q, "ready": ready, "mutate": false,
	})
}

func writeLiveOpportunity(ws worldstate.WorldState, ranks []opportunity.Ranked, ps []orchestrator.DecisionProposal, elig *eligibility.Registry) {
	type row struct {
		Rank, Market, Attention, Coverage, Tier, Setup, Eligibility string
	}
	var rows []row
	var b strings.Builder
	b.WriteString("# LIVE OPPORTUNITY SURFACE\n\nATTENTION_SCORE_V1 frozen. world_hash=" + ws.Hash + "\n\n")
	b.WriteString("| Rank | Market | Attention | Coverage | Tier | Setup | Eligibility |\n|---|---|---|---|---|---|---|\n")
	for _, r := range ranks {
		e := elig.Get(r.Market)
		rows = append(rows, row{
			Rank: itoaBot(r.Rank), Market: r.Market, Attention: fnum(r.Score),
			Coverage: fnum(r.Coverage) + "%", Tier: string(r.Tier), Setup: string(r.State.Setup),
			Eligibility: string(e.Status),
		})
		b.WriteString("| " + itoaBot(r.Rank) + " | " + r.Market + " | " + fnum(r.Score) + " | " + fnum(r.Coverage) + " | " + string(r.Tier) + " | " + string(r.State.Setup) + " | " + string(e.Status) + " |\n")
	}
	_ = ps
	_ = worldstate.WriteJSON(filepath.Join("research/reports", "LIVE_OPPORTUNITY_SURFACE.json"), map[string]any{
		"as_of": ws.AsOf, "hash": ws.Hash, "valid": ws.Valid, "ranks": ranks, "proposals": ps,
	})
	_ = worldstate.WriteMD(filepath.Join("research/reports", "LIVE_OPPORTUNITY_SURFACE.md"), b.String())
}

func writeReadiness(rows []demoprep.Row, ranks []opportunity.Ranked) {
	cov := map[string]float64{}
	for _, r := range ranks {
		cov[r.Market] = r.Coverage
	}
	var b strings.Builder
	b.WriteString("# MULTI-MARKET READINESS\n\nGates only. No return optimization. MULTI_MARKET_DEMO_EXECUTION=OFF.\n\n")
	b.WriteString("| Market | Data | Strategy | Monetary | Execution | Risk | Operations | Total |\n|---|---|---|---|---|---|---|---|\n")
	var out []demoprep.Readiness
	for _, r := range rows {
		rd := demoprep.ScoreReadiness(r, cov[r.Market], eligibility.Entry{Canonical: r.Market, Status: r.Eligibility})
		out = append(out, rd)
		b.WriteString("| " + rd.Market + " | " + itoaBot(rd.Data) + " | " + itoaBot(rd.Strategy) + " | " + itoaBot(rd.Monetary) + " | " + itoaBot(rd.Execution) + " | " + itoaBot(rd.Risk) + " | " + itoaBot(rd.Operations) + " | " + itoaBot(rd.Total) + " |\n")
	}
	_ = worldstate.WriteJSON(filepath.Join("research/reports", "MULTI_MARKET_READINESS.json"), out)
	_ = worldstate.WriteMD(filepath.Join("research/reports", "MULTI_MARKET_READINESS.md"), b.String())
}

func jsonBytes(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func shadowMD(rep map[string]any, ps []orchestrator.DecisionProposal) string {
	var b strings.Builder
	b.WriteString("# MULTI-MARKET SHADOW\n\nNo orders. MULTI_MARKET_DEMO_EXECUTION=OFF.\n\n")
	for _, p := range ps {
		b.WriteString("- " + p.Market + " attention=" + fnum(p.AttentionScore) + " coverage=" + fnum(p.Coverage) + " elig=" + string(p.Eligibility) + " hash=" + p.WorldHash + " → " + p.Decision + "\n")
	}
	_ = rep
	return b.String()
}

func fnum(v float64) string {
	return strings.TrimRight(strings.TrimRight(
		func() string {
			s := ""
			n := int(v*10 + 0.5)
			neg := n < 0
			if n < 0 {
				n = -n
			}
			s = itoaBot(n/10) + "." + itoaBot(n%10)
			if neg {
				s = "-" + s
			}
			return s
		}(), "0"), ".")
}
