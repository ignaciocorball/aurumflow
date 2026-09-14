package main

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"sync"

	"aurumflow/internal/eligibility"
	"aurumflow/internal/featready"
	"aurumflow/internal/globalsources"
	"aurumflow/internal/legacycompat"
	"aurumflow/internal/livesurface"
	"aurumflow/internal/logger"
	"aurumflow/internal/microcap"
	"aurumflow/internal/mktwarmup"
	"aurumflow/internal/dataint"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/ops"
	"aurumflow/internal/salience"
	"aurumflow/internal/orchestrator"
	"aurumflow/internal/quotehealth"
	"aurumflow/internal/research"
	"aurumflow/internal/researchopp"
	"aurumflow/internal/scanner"
	"aurumflow/internal/sessions"
	"aurumflow/internal/worldstate"
	"aurumflow/pkg/models"
)

var (
	slowMu   sync.Mutex
	slowHold worldstate.WorldState
	prevOpp  = map[string]*researchopp.Signal{}
	pricePath = map[string][]researchopp.PricePoint{}
	labeledOpp = map[string]map[string]bool{}
	lastLabelAt time.Time
)

func runWorldSnapshot(fixture bool) {
	now := time.Now().UTC()
	ws, ranks, n, err := writeWorldReports("research/reports", fixture, now)
	if err != nil {
		logger.Error("world snapshot: %v", err)
		return
	}
	logger.Info("world snapshot as_of=%s valid=%s ranks=%d hist_points=%d liq=%s risk=%s fixture=%v",
		ws.AsOf.Format(time.RFC3339), ws.Valid, len(ranks), n, ws.Liquidity.Class, ws.Risk, fixture)
}

func writeWorldReports(reportDir string, fixture bool, now time.Time) (worldstate.WorldState, []opportunity.Ranked, int, error) {
	in, _ := productionOrFixture(context.Background(), now, fixture)
	ws := worldstate.At(now, in)
	ranks := opportunity.Rank(ws)
	ws = scanner.AttachOpportunity(ws, ranks)
	start := now.AddDate(-1, 0, 0)
	points := worldstate.DailyAsOf(start, now, 7*24*time.Hour)
	hist := worldstate.Replay(in, points)
	views := rankViews(ranks)
	if fixture {
		if err := worldstate.WriteJSON(filepath.Join(reportDir, "CURRENT_WORLD_STATE.json"), ws); err != nil {
			return ws, ranks, 0, err
		}
		if err := worldstate.WriteMD(filepath.Join(reportDir, "CURRENT_WORLD_STATE.md"), worldstate.CurrentMarkdown(ws, views)); err != nil {
			return ws, ranks, 0, err
		}
	} else {
		if err := worldstate.WriteJSON(filepath.Join(reportDir, "LIVE_WORLD_STATE.json"), ws); err != nil {
			return ws, ranks, 0, err
		}
		if err := worldstate.WriteMD(filepath.Join(reportDir, "LIVE_WORLD_STATE.md"), liveMarkdown(ws, views)); err != nil {
			return ws, ranks, 0, err
		}
		if err := worldstate.WriteJSON(filepath.Join(reportDir, "CURRENT_WORLD_STATE.json"), ws); err != nil {
			return ws, ranks, 0, err
		}
		if err := worldstate.WriteMD(filepath.Join(reportDir, "CURRENT_WORLD_STATE.md"), liveMarkdown(ws, views)); err != nil {
			return ws, ranks, 0, err
		}
	}
	if err := worldstate.WriteJSON(filepath.Join(reportDir, "OPPORTUNITY_SURFACE.json"), ranks); err != nil {
		return ws, ranks, 0, err
	}
	if err := worldstate.WriteMD(filepath.Join(reportDir, "OPPORTUNITY_SURFACE.md"), worldstate.OpportunityMarkdown(views)); err != nil {
		return ws, ranks, 0, err
	}
	covRows := globalsources.OfficialCoverage(in.Observations, now)
	cov := coverageLabel(covRows, fixture)
	if err := worldstate.WriteJSON(filepath.Join(reportDir, "GLOBAL_CAPITAL_12M.json"), map[string]any{
		"snapshots": len(hist), "coverage": cov, "rows": covRows, "latest": ws,
	}); err != nil {
		return ws, ranks, 0, err
	}
	if err := worldstate.WriteMD(filepath.Join(reportDir, "GLOBAL_CAPITAL_12M.md"), worldstate.Capital12Markdown(ws, len(hist), cov)); err != nil {
		return ws, ranks, 0, err
	}
	if !fixture {
		_ = worldstate.WriteJSON(filepath.Join(reportDir, "OFFICIAL_12M_COVERAGE.json"), covRows)
		_ = worldstate.WriteMD(filepath.Join(reportDir, "OFFICIAL_12M_COVERAGE.md"), coverageMD(covRows))
	}
	return ws, ranks, len(hist), nil
}

func productionOrFixture(ctx context.Context, now time.Time, fixture bool) (worldstate.Input, []globalsources.FetchResult) {
	if fixture {
		return worldstate.LoadResearchInput("internal/globalsources/testdata", "", now), nil
	}
	return worldstate.LoadOfficialInput(ctx, now)
}

func rankViews(ranks []opportunity.Ranked) []worldstate.RankView {
	out := make([]worldstate.RankView, len(ranks))
	for i, r := range ranks {
		out[i] = worldstate.RankView{Rank: r.Rank, Market: r.Market, Score: r.Score, Coverage: r.Coverage, Tier: string(r.Tier), Why: r.Why, Risks: r.Risks, Freshness: string(r.Freshness), State: r.State}
	}
	return out
}

func liveMarkdown(ws worldstate.WorldState, ranks []worldstate.RankView) string {
	var b strings.Builder
	b.WriteString(worldstate.CurrentMarkdown(ws, ranks))
	b.WriteString("\n## Source audit\n")
	b.WriteString("Valid: " + ws.Valid + "\n")
	b.WriteString("Origins: " + strings.Join(ws.Origins, ",") + "\n")
	b.WriteString("Rejected fixtures: " + itoaBot(ws.RejectedFix) + "\n")
	return b.String()
}

func coverageLabel(rows []globalsources.CoverageRow, fixture bool) string {
	if fixture {
		return "FIXTURE — not official coverage"
	}
	parts := make([]string, 0, len(rows))
	for _, r := range rows {
		parts = append(parts, r.Source+" "+r.Label)
	}
	if len(parts) == 0 {
		return "official cache empty"
	}
	return strings.Join(parts, "; ")
}

func coverageMD(rows []globalsources.CoverageRow) string {
	var b strings.Builder
	b.WriteString("# OFFICIAL 12M COVERAGE\n\nActual observation counts only. Missing days are not fabricated.\n\n")
	for _, r := range rows {
		b.WriteString("- " + r.Source + ": " + r.Label + "\n")
	}
	return b.String()
}

func itoaBot(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func publishWorld(srv *ops.Server, fixture bool) {
	now := time.Now().UTC()
	ctx := context.Background()
	var in worldstate.Input
	if fixture {
		in = worldstate.LoadResearchInput("internal/globalsources/testdata", "", now)
	} else {
		in = worldstate.LoadCachedOfficialInput(ctx)
		if len(in.Observations) == 0 {
			in, _ = worldstate.LoadOfficialInput(ctx, now)
		}
	}
	st := srv.Get()
	in.Micro = microcap.BTCFromRuntime(true, st.L2QuotesOK || st.BookSynced, st.LastPressure != 0 || st.L2QuotesOK, st.LastV1Class != "", st.Absorption != 0 || strings.Contains(strings.ToUpper(st.LastV1Class), "ABSORB"), healthFromStatus(st))
	in.BTCMicro = applyIntelMicro(st)
	ws := worldstate.At(now, in)
	ws.SlowAt = now
	if lastFrameOK {
		ws = worldstate.ApplyLiveFrame(ws, lastFrame)
	}
	elig := eligibility.New()
	ws = elig.ApplyToWorld(ws)
	ranks := opportunity.Rank(ws)
	orch := orchestrator.New()
	for i, r := range ranks {
		e := elig.Get(r.Market)
		p := orch.ProposeFull(ws, r, "", "", e.Status, nil)
		stt := r.State
		stt.Attention = r.Score
		stt.Coverage = r.Coverage
		stt.Proposal = p.Decision
		stt.Eligibility = e.Status
		stt.EligReason = e.Reason
		ranks[i].State = stt
	}
	ws = scanner.AttachOpportunity(ws, ranks)
	slowMu.Lock()
	slowHold = ws
	slowMu.Unlock()
	reg := globalsources.NewRegistry()
	truth := "UNKNOWN"
	if !fixture && ws.Valid == "CURRENT_WORLD_STATE_VALID" {
		truth = "LIVE_OFFICIAL"
		for _, o := range ws.Origins {
			if o == "CACHE_OFFICIAL" {
				truth = "CACHED_OFFICIAL"
			}
		}
	} else if fixture {
		truth = "FIXTURE"
	}
	if peer := goldPeerCopy(); peer != nil {
		srv.SetSources(map[string]any{
			"sources":     reg.All(),
			"catalog":     globalsources.Catalog(),
			"calendar":    globalsources.Calendar(now, globalsources.LastActuals()),
			"eligibility": elig.All(),
			"truth":       truth,
			"valid":       ws.Valid,
			"gold_peer":   peer,
			"risk_unit":   "account_currency_risk",
			"risk_cap":    "Aggregate risk cap: $300 DEMO",
		})
	} else {
		srv.SetSources(map[string]any{
			"sources":     reg.All(),
			"catalog":     globalsources.Catalog(),
			"calendar":    globalsources.Calendar(now, globalsources.LastActuals()),
			"eligibility": elig.All(),
			"truth":       truth,
			"valid":       ws.Valid,
			"risk_unit":   "account_currency_risk",
			"risk_cap":    "Aggregate risk cap: $300 DEMO",
		})
	}
	srv.SetWorld(ws)
	bumpWorld()
}

func refreshWorld(srv *ops.Server, fixture bool) {
	if !fixture {
		go func() {
			_, _ = worldstate.LoadOfficialInput(context.Background(), time.Now().UTC())
			publishWorld(srv, false)
		}()
	}
	publishWorld(srv, fixture)
	go func() {
		cacheTick := time.NewTicker(60 * time.Second)
		liveTick := time.NewTicker(6 * time.Hour)
		defer cacheTick.Stop()
		defer liveTick.Stop()
		liveFrame := time.NewTicker(5 * time.Second)
		defer liveFrame.Stop()
		for {
			select {
			case <-cacheTick.C:
				publishWorld(srv, fixture)
			case <-liveTick.C:
				if !fixture {
					_, _ = worldstate.LoadOfficialInput(context.Background(), time.Now().UTC())
				}
				publishWorld(srv, fixture)
			case <-liveFrame.C:
				materializeLive(srv)
			}
		}
	}()
}

func healthFromStatus(st ops.Status) string {
	if st.L2QuotesOK && st.BookSynced {
		return "HEALTHY"
	}
	if st.L2QuotesOK {
		return "OK"
	}
	return "UNKNOWN"
}

func materializeLive(srv *ops.Server) {
	slowMu.Lock()
	ws := slowHold
	slowMu.Unlock()
	if ws.AsOf.IsZero() {
		return
	}
	now := time.Now().UTC()
	if intelOn && demoEnvConfigured() && now.Sub(lastQuoteAt) >= 30*time.Second {
		maps := cachedMaps()
		if frame, ok := collectCapitalFrame(context.Background(), maps, now); ok {
			lastFrame = frame
			lastFrameOK = true
			lastQuoteAt = now
			stale, open, n := 0, 0, 0
			for _, q := range frame.Quotes {
				n++
				if q.Stale {
					stale++
				}
				if q.Live {
					open++
				}
				if tr, ok := sessions.BrokerTransition(q.Market, prevStatus[q.Market], q.MarketStatus, now); ok {
					logger.Info("MARKET TRANSITION %s %s → %s", tr.Market, tr.From, tr.To)
					prevStatus[q.Market] = q.MarketStatus
				} else if prevStatus[q.Market] == "" {
					prevStatus[q.Market] = q.MarketStatus
				}
			}
			bumpQuotes(n, stale, open)
		} else if !lastQuoteAt.IsZero() && now.Sub(lastQuoteAt) > 3*time.Minute {
			lastFrameOK = false
			bumpErr()
		}
	}
	if lastFrameOK {
		ws = worldstate.ApplyLiveFrame(ws, lastFrame)
	}
	st := srv.Get()
	ws.Markets = cloneMarkets(ws.Markets)
	if btc, ok := ws.Markets["BTC"]; ok {
		btc.MicroAvailable = applyIntelMicro(st)
		if btc.MicroAvailable {
			ws.MicroAt = now
		}
		ws.Markets["BTC"] = btc
	}
	goldLegacy := ""
	if !intelOn {
		ws = worldstate.Finalize(ws)
		ranks := opportunity.Rank(ws)
		ws = scanner.AttachOpportunity(ws, ranks)
		srv.SetWorld(ws)
		return
	}
	for id, stt := range ws.Markets {
		h := historyOf(id)
		q := lastFrame.Quotes[id]
		f := lastFrame.Features[id]
		fr := featready.Assess(id, q, f, h)
		histMu.Lock()
		featByMkt[id] = fr
		histMu.Unlock()
		stt.HistoryStatus = h.Status
		stt.FeatPrice, stt.FeatMomentum, stt.FeatVol, stt.FeatCross, stt.FeatLegacy = fr.Price, fr.Momentum, fr.Volatility, fr.CrossAsset, fr.Legacy
		stt.TapeState = quotehealth.TapeState(q, h.Status)
		stt.QuoteHealth = string(quotehealth.Of(q, 1, 30*time.Second, 0).Status)
		stt.LegacyCompat = legacycompat.Of(id).Status
		ws.Markets[id] = stt
		if id == "GOLD" && h.Status == mktwarmup.StatusReady {
			goldLegacy = evalGoldLegacy(h)
		}
	}
	ws = worldstate.Finalize(ws)
	elig := eligibility.New()
	ws = elig.ApplyToWorld(ws)
	ranks := opportunity.Rank(ws)
	orch := orchestrator.New()
	for i, r := range ranks {
		h := historyOf(r.Market)
		ready := h.Status == mktwarmup.StatusReady
		legacy := ""
		if r.Market == "GOLD" {
			legacy = goldLegacy
		}
		extra := orchestrator.Extra{WorldHash: ws.Hash, HistoryKnown: true, HistoryPresent: ready && r.Market == "GOLD"}
		if r.Market != "GOLD" {
			extra.HistoryPresent = false
			extra.HistoryKnown = true
		}
		p := orch.ProposeWith(ws, r, legacy, "", r.State.Eligibility, []string{"unknown monetary risk blocks new order"}, extra)
		stt := r.State
		stt.Proposal = p.Decision
		stt.Setup = p.Setup
		stt.Attention = r.Score
		stt.Coverage = r.Coverage
		sal := salienceFor(r.Market, stt, ranks)
		stt.Salience = sal.Score
		stt.SalienceReason = sal.Reason
		ranks[i].State = stt
		integ := dataint.Classify(st.Drops+st.PersistDrops, st.BookGaps, st.Resyncs, st.BookSynced)
		if st.IntegrityStatus != "" {
			integ.Status = st.IntegrityStatus
			integ.Reason = st.IntegrityReason
		}
		sig := researchopp.Stamp(researchopp.Signal{
			ID: r.Market + "-" + now.Format("20060102150405"), T0: now, WorldHash: ws.Hash,
			Proposal: jsonBytes(p), Ranking: jsonBytes(r), MarketState: jsonBytes(stt),
			Tier: string(r.Tier), Eligibility: string(r.State.Eligibility),
			Session: string(ws.Session), MarketStatus: stt.MarketStatus, HistoryReady: stt.HistoryStatus,
			Attention: r.Score, Coverage: r.Coverage, Legacy: string(p.Setup), Setup: string(p.Setup),
			Micro: map[bool]string{true: "AVAILABLE", false: "UNAVAILABLE"}[stt.MicroAvailable],
			ProposalLabel: p.Decision, DataQuality: string(stt.DataQuality),
			Salience: sal.Score, Components: r.Components,
			IntegrityStatus: integ.Status, IntegrityReason: integ.Reason,
		})
		if researchopp.ShouldRecord(prevOpp[r.Market], sig) {
			_ = researchopp.Record("journals", sig)
			cp := sig
			prevOpp[r.Market] = &cp
			bumpOpp()
		}
		bumpProposal()
	}
	if lastFrameOK {
		for id, q := range lastFrame.Quotes {
			if q.Mid <= 0 {
				continue
			}
			pricePath[id] = append(pricePath[id], researchopp.PricePoint{T: now, P: q.Mid})
			if len(pricePath[id]) > 4000 {
				pricePath[id] = pricePath[id][len(pricePath[id])-2000:]
			}
		}
	}
	if lastLabelAt.IsZero() || now.Sub(lastLabelAt) >= 30*time.Second {
		_, _, _ = researchopp.LabelMature("journals", now, pricePath, labeledOpp)
		lastLabelAt = now
	}
	ws = scanner.AttachOpportunity(ws, ranks)
	_ = elig
	srv.SetWorld(ws)
}

func evalGoldLegacy(h mktwarmup.History) string {
	bars := h.M15Bars
	if len(bars) < 21 {
		bars = h.M5Bars
	}
	if len(bars) < 21 || len(h.H1Bars) < 5 || len(h.H4Bars) < 3 {
		return ""
	}
	toM := func(xs []livesurface.Candle) []models.Candle {
		var out []models.Candle
		for _, c := range xs {
			out = append(out, models.Candle{Time: c.Time, Open: c.Close, High: c.High, Low: c.Low, Close: c.Close})
		}
		return out
	}
	bumpLegacy()
	cfg := research.CryptoResearchConfig()
	cfg.Crypto = false
	cfg.TradingSessions = []string{"LONDON", "NY"}
	sigs := research.ScanLegacy(toM(bars), toM(h.H1Bars), toM(h.H4Bars), cfg)
	if len(sigs) == 0 {
		return ""
	}
	last := sigs[len(sigs)-1]
	if last.Direction > 0 {
		return "LONG"
	}
	if last.Direction < 0 {
		return "SHORT"
	}
	return ""
}

func salienceFor(market string, st worldstate.MarketState, ranks []opportunity.Ranked) salience.Result {
	h := historyOf(market)
	var bars []salience.Bar
	src := h.M15Bars
	if len(src) < 8 {
		src = h.M5Bars
	}
	for _, c := range src {
		bars = append(bars, salience.Bar{Close: c.Close, High: c.High, Low: c.Low})
	}
	var xs []float64
	for _, r := range ranks {
		if r.State.MarketStatus != "TRADEABLE" {
			continue
		}
		peer := historyOf(r.Market)
		srcp := peer.M15Bars
		if len(srcp) < 8 {
			srcp = peer.M5Bars
		}
		if len(srcp) < 8 {
			continue
		}
		look := 12
		if len(srcp) < look+1 {
			look = len(srcp) - 1
		}
		a, b := srcp[len(srcp)-1].Close, srcp[len(srcp)-1-look].Close
		if b > 0 {
			ret := a - b
			if ret < 0 {
				ret = -ret
			}
			xs = append(xs, ret/b)
		}
	}
	return salience.Score(salience.Input{Market: market, Bars: bars, DQPenalty: salience.DQPenalty(string(st.DataQuality)), CrossAbsRet: xs})
}

func cloneMarkets(in map[string]worldstate.MarketState) map[string]worldstate.MarketState {
	out := make(map[string]worldstate.MarketState, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
