package main

import (
	"context"
	"os"
	"strings"
	"time"

	"aurumflow/config"
	"aurumflow/internal/binancehist"
	"aurumflow/internal/capitalhist"
	"aurumflow/internal/catalog"
	"aurumflow/internal/cftc"
	"aurumflow/internal/finra"
	"aurumflow/internal/logger"
	"aurumflow/internal/macro"
	"aurumflow/internal/market"
	"aurumflow/internal/optionsctx"
	"aurumflow/internal/secctx"
)

func runDownloadResearch(ctx context.Context, preset string, days int) {
	if days <= 0 {
		days = 30
	}
	catPath := "data/catalog.json"
	cat, _ := catalog.Load(catPath)
	to := time.Now().UTC().AddDate(0, 0, -1)
	from := to.AddDate(0, 0, -(days - 1))

	arch := binancehist.New("data", "BTCUSDT")
	var paths []string
	for _, day := range binancehist.DaysInclusive(from, to) {
		p, err := arch.DownloadDay(ctx, day)
		if err != nil {
			logger.Warn("binance %s: %v", day, err)
			continue
		}
		sum, ok, cErr := arch.VerifyChecksum(ctx, day, p)
		q, _ := arch.ScanDay(p)
		st := "ok"
		if cErr != nil {
			st = "checksum_unverified"
		} else if !ok {
			st = "checksum_mismatch"
		}
		cat.Upsert(catalog.Dataset{
			Provider: "binance_vision", Instrument: "BTCUSDT", Type: "aggTrades",
			From: day, To: day, Rows: q.Rows, Quality: "DIRECT", Gaps: q.Gaps,
			Checksum: sum, DownloadedAt: catalog.NowUTC(), SourceVersion: "data.binance.vision", Path: p, Status: st,
		})
		paths = append(paths, p)
		logger.Info("binance day=%s rows=%d checksum=%s status=%s", day, q.Rows, short(sum), st)
		_ = cat.Save(catPath)
	}
	logger.Info("binance days downloaded=%d requested=%d", len(paths), days)

	if demoEnvConfigured() {
		sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
		if err != nil {
			logger.Warn("capital hist: %v", err)
		} else {
			hp := &capitalhist.HistoricalProvider{Client: sess.client, Dir: "data"}
			found, missing := capitalhist.Discover(ctx, sess.client, capitalhist.CoreEpics)
			logger.Info("capital universe found=%d missing=%d %v", len(found), len(missing), missing)
			seen := map[string]bool{}
			for _, epic := range found {
				if seen[epic] {
					continue
				}
				seen[epic] = true
				if len(seen) > 8 {
					break
				}
				cs, err := hp.FetchWindow(ctx, epic, market.ResolutionHour, to.AddDate(0, 0, -30), to, 500)
				if err != nil {
					logger.Warn("capital %s: %v", epic, err)
					continue
				}
				path, _ := hp.Save(epic, market.ResolutionHour, cs)
				cat.Upsert(capitalhist.CatalogEntry(epic, market.ResolutionHour, cs, path))
				logger.Info("capital %s hours=%d", epic, len(cs))
			}
			_ = cat.Save(catPath)
		}
	} else {
		logger.Info("capital historical skipped (no AURUMFLOW_DEMO_*)")
	}

	sec := secctx.New("data/sec")
	tickers, err := sec.Tickers(ctx)
	if err != nil {
		logger.Warn("SEC tickers: %v", err)
	} else {
		g := &secctx.Graph{}
		for _, sym := range []string{"AAPL", "MSFT", "NVDA", "META", "AMZN"} {
			c, ok := tickers[sym]
			if !ok {
				logger.Info("SEC missing ticker %s", sym)
				continue
			}
			fils, err := sec.Submissions(ctx, c.CIK)
			if err != nil {
				logger.Warn("SEC %s: %v", sym, err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			g.AddCompanyFilings(c, fils)
			facts, fErr := sec.CompanyFacts(ctx, c.CIK)
			if fErr != nil {
				logger.Warn("SEC facts %s: %v", sym, fErr)
			} else {
				g.AddFacts(c, facts)
			}
			logger.Info("SEC %s cik=%s filings=%d facts=%d", sym, c.CIK, len(fils), len(facts))
			time.Sleep(200 * time.Millisecond)
		}
		_ = g.Save("research/context/institutional_graph.json")
		cat.Upsert(catalog.Dataset{Provider: "sec", Instrument: "megacap", Type: "submissions", Rows: len(g.Filings), Quality: "DELAYED", DownloadedAt: catalog.NowUTC(), SourceVersion: "data.sec.gov", Path: "research/context/institutional_graph.json", Status: "ok"})
	}

	if rows, err := cftc.FetchCurrent(ctx); err != nil {
		logger.Warn("CFTC: %v", err)
		cat.Upsert(catalog.Dataset{Provider: "cftc", Instrument: "GOLD", Type: "cot_disagg", Status: "unavailable", Quality: "SLOW_CONTEXT", DownloadedAt: catalog.NowUTC()})
	} else {
		logger.Info("CFTC GOLD rows=%d", len(rows))
		cat.Upsert(catalog.Dataset{Provider: "cftc", Instrument: "GOLD", Type: "cot_disagg", Rows: len(rows), Quality: "SLOW_CONTEXT", DownloadedAt: catalog.NowUTC(), SourceVersion: "cftc.gov/dea/newcot", Status: "ok"})
	}
	fp := finra.New()
	finraWeeks := 0
	for _, sym := range []string{"AAPL", "MSFT", "NVDA", "META", "AMZN"} {
		ws, err := fp.FetchSymbol(ctx, sym)
		if err != nil {
			logger.Warn("FINRA %s: %v", sym, err)
			continue
		}
		finraWeeks += len(ws)
		logger.Info("FINRA %s weeks=%d", sym, len(ws))
		time.Sleep(250 * time.Millisecond)
	}
	st := "ok"
	if finraWeeks == 0 {
		st = "unavailable_or_empty"
	}
	cat.Upsert(catalog.Dataset{Provider: "finra", Instrument: "megacap", Type: "otc_weekly", Rows: finraWeeks, Quality: "SLOW_CONTEXT", DownloadedAt: catalog.NowUTC(), SourceVersion: "api.finra.org", Status: st})

	m := macro.New()
	logger.Info("FRED status=%s", m.Status)
	opt := optionsctx.New()
	logger.Info("OPTIONS realtime=%s historical=%s", opt.Realtime, opt.Historical)
	_ = cat.Save(catPath)
	logger.Info("download preset=%s catalog=%s", preset, catPath)
}

func short(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func envTrue(k string) bool { return strings.TrimSpace(os.Getenv(k)) != "" }
