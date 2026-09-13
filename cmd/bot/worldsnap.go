package main

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"sync"

	"aurumflow/internal/eligibility"
	"aurumflow/internal/globalsources"
	"aurumflow/internal/logger"
	"aurumflow/internal/microcap"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/ops"
	"aurumflow/internal/orchestrator"
	"aurumflow/internal/scanner"
	"aurumflow/internal/worldstate"
)

var (
	slowMu   sync.Mutex
	slowHold worldstate.WorldState
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
	in.BTCMicro = in.Micro.Healthy()
	ws := worldstate.At(now, in)
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
	srv.SetWorld(ws)
	srv.SetSources(map[string]any{
		"sources":     reg.All(),
		"catalog":     globalsources.Catalog(),
		"calendar":    globalsources.Calendar(now, globalsources.LastActuals()),
		"eligibility": elig.All(),
		"truth":       truth,
		"valid":       ws.Valid,
	})
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
	st := srv.Get()
	ws.Markets = cloneMarkets(ws.Markets)
	if btc, ok := ws.Markets["BTC"]; ok {
		cap := microcap.BTCFromRuntime(true, st.L2QuotesOK || st.BookSynced, st.LastPressure != 0 || st.L2QuotesOK, st.LastV1Class != "", st.Absorption != 0, healthFromStatus(st))
		btc.MicroAvailable = cap.Healthy()
		ws.Markets["BTC"] = btc
	}
	ws = worldstate.Finalize(ws)
	elig := eligibility.New()
	ranks := opportunity.Rank(ws)
	ws = scanner.AttachOpportunity(ws, ranks)
	_ = elig
	srv.SetWorld(ws)
}

func cloneMarkets(in map[string]worldstate.MarketState) map[string]worldstate.MarketState {
	out := make(map[string]worldstate.MarketState, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
