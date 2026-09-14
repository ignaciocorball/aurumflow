package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"aurumflow/config"
	"aurumflow/internal/capsched"
	"aurumflow/internal/capitalhist"
	"aurumflow/internal/catalog"
	"aurumflow/internal/chronosplit"
	"aurumflow/internal/costmodel"
	"aurumflow/internal/disloc"
	"aurumflow/internal/execacct"
	"aurumflow/internal/instrument"
	"aurumflow/internal/legacynorm"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/mktval"
	"aurumflow/internal/money"
	"aurumflow/internal/research"
	"aurumflow/internal/shadowport"
	"aurumflow/pkg/models"
)

type fetched struct {
	id              string
	m5, m15, h1, h4 []models.Candle
	spread          float64
	broker          string
	monetary        string
	err             error
}

func runP89Fabric(days int) {
	ctx := context.Background()
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		logger.Error("P8.9 DEMO session: %v", err)
		os.Exit(1)
	}
	if sess.Identity.MayTrade == false {
		logger.Error("P8.9 BLOCK: explicit DEMO account not verified — no broker mutation")
	}
	if config.IsLiveCapitalHost(sess.cfg.API.BaseURL) {
		logger.Error("LIVE host")
		os.Exit(1)
	}
	maps, _ := instrument.LoadMappings("data/world/normalized/capital/mappings.json")
	byID := map[string]instrument.Mapping{}
	for _, m := range maps {
		byID[m.Canonical] = m
	}
	root := "data/research/p89"
	_ = os.MkdirAll(root, 0o755)
	reg, _ := catalog.Load(filepath.Join(root, "DATASET_REGISTRY.json"))
	sched := capsched.New(2, 300*time.Millisecond)
	hist := &capitalhist.HistoricalProvider{Client: sess.client, Dir: root}
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -days)
	// fetched is package-level
	outCh := make(chan fetched, len(mktval.Universe))
	var wg sync.WaitGroup
	for _, id := range mktval.Universe {
		m := byID[id]
		if m.Epic == "" {
			outCh <- fetched{id: id, err: fmt.Errorf("unresolved"), broker: "UNRESOLVED", monetary: "UNKNOWN"}
			continue
		}
		wg.Add(1)
		go func(id string, m instrument.Mapping) {
			defer wg.Done()
			f := fetched{id: id, broker: "UNKNOWN", monetary: "UNVERIFIED"}
			_ = sched.Do(ctx, func(ctx context.Context) error {
				md, err := sess.client.GetMarketDetails(ctx, m.Epic)
				if err != nil {
					return err
				}
				if md != nil && md.Snapshot.Offer > md.Snapshot.Bid {
					f.spread = md.Snapshot.Offer - md.Snapshot.Bid
				}
				spec := money.FromMarketDetails(md)
				f.broker = "READ"
				if spec.DealingComplete() {
					f.broker = "SPEC_READY"
				}
				if cached, err := money.LoadSpec(money.CachePath("", id)); err == nil && cached.ValidationStatus == money.RuntimeValidated {
					f.monetary = money.RuntimeValidated
				} else {
					f.monetary = spec.ValidationStatus
				}
				return nil
			})
			loadOrFetch := func(res string) []models.Candle {
				path := filepath.Join(root, "capital", id, res+".json")
				if raw, err := os.ReadFile(path); err == nil {
					var cs []models.Candle
					if json.Unmarshal(raw, &cs) == nil && len(cs) > 50 {
						return capitalhist.DedupSort(cs)
					}
				}
				var cs []models.Candle
				_ = sched.Do(ctx, func(ctx context.Context) error {
					got, err := hist.FetchWindow(ctx, m.Epic, res, from, to, 500)
					cs = got
					return err
				})
				if len(cs) > 0 {
					p, _ := hist.Save(id, res, cs)
					d := capitalhist.CatalogEntry(id, res, cs, p)
					d.Checksum = hashCandles(cs)
					reg.Upsert(d)
				}
				return cs
			}
			f.m5 = loadOrFetch(market.ResolutionMinute5)
			f.m15 = loadOrFetch(market.ResolutionMinute15)
			f.h1 = loadOrFetch(market.ResolutionHour)
			f.h4 = loadOrFetch(market.ResolutionHour4)
			outCh <- f
		}(id, m)
	}
	go func() { wg.Wait(); close(outCh) }()
	var rows []mktval.Row
	var port shadowport.Portfolio
	var highSal int
	var salSum float64
	salN := 0
	for f := range outCh {
		row := evaluateFetched(f)
		rows = append(rows, row)
		if row.Status == mktval.ShadowValidated {
			logger.Info("P8.9 SHADOW_VALIDATED %s holdout=%d exp=%.3f", row.Market, row.HoldoutN, row.Expectancy)
		}
	}
	_ = reg.Save(filepath.Join(root, "DATASET_REGISTRY.json"))
	order := map[string]int{}
	for i, id := range mktval.Universe {
		order[id] = i
	}
	for i := 1; i < len(rows); i++ {
		j := i
		for j > 0 && order[rows[j].Market] < order[rows[j-1].Market] {
			rows[j], rows[j-1] = rows[j-1], rows[j]
			j--
		}
	}
	_ = os.WriteFile("research/reports/MARKET_VALIDATION_MATRIX.md", []byte(mktval.MatrixMarkdown(rows)), 0o644)
	_ = os.WriteFile("research/reports/MARKET_PROMOTION_RANKING.md", []byte(mktval.RankingMarkdown(rows)), 0o644)
	raw, _ := json.MarshalIndent(rows, "", "  ")
	_ = os.WriteFile("research/reports/MARKET_VALIDATION_MATRIX.json", raw, 0o644)
	st := sched.Stats()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	logger.Info("P8.9 fabric done markets=%d conc=%d req=%d 429=%d err=%d p50=%.0f p95=%.0f rpm=%.1f mem=%.1fMB",
		len(rows), sched.Concurrency(), st.Requests, st.Limited, st.Errors, st.P50Ms, st.P95Ms, st.PerMin, float64(ms.Alloc)/1024/1024)
	_ = port
	_ = disloc.Classify(disloc.Input{HighSalience: highSal, MedianSal: salSum / float64(max1(salN))})
	_ = execacct.Mask(sess.Identity.AccountID)
}

func evaluateFetched(f fetched) mktval.Row {
	r := mktval.Row{Market: f.id, BrokerSpec: f.broker, Monetary: f.monetary, Shadow: "SHADOW", DemoElig: "NO", Trust: "NONE", Data: "MISSING", History: "NONE", Mechanics: "UNKNOWN", NormCompat: "PENDING"}
	if f.err != nil {
		r.Status = mktval.ResearchRejected
		r.Data = "UNRESOLVED"
		return r
	}
	r.M5, r.M15, r.H1, r.H4 = len(f.m5), len(f.m15), len(f.h1), len(f.h4)
	if len(f.m15) > 0 {
		r.From = f.m15[0].Time.UTC().Format("2006-01-02")
		r.To = f.m15[len(f.m15)-1].Time.UTC().Format("2006-01-02")
		r.DatasetHash = hashCandles(f.m15)[:16]
		r.Data = "OK"
		r.History = fmt.Sprintf("%d M15", len(f.m15))
	}
	r.Spread = f.spread
	r.MedianATR = legacynorm.MedianATR(f.m15)
	r.LegacyCompat = legacynorm.Compat(f.id, r.MedianATR)
	r.NormCompat = "DIMENSIONLESS"
	r.Mechanics = fmt.Sprintf("atr=%.3f spread=%.4f", r.MedianATR, r.Spread)
	if len(f.m15) < 80 {
		r.Status = mktval.DataReady
		return r
	}
	split := chronosplit.Of(f.m15[0].Time, f.m15[len(f.m15)-1].Time)
	cur := research.ScanLegacy(f.m15, f.h1, f.h4, legacynorm.CurrentLegacyConfig())
	norm := legacynorm.ScanNormalized(f.m15, f.h1, f.h4)
	r.Signals = len(norm)
	zero := mktval.Simulate(norm, f.m15, f.spread, costmodel.Zero, split)
	tr := mktval.Simulate(norm, f.m15, f.spread, costmodel.Normal, split)
	_ = cur
	var disc, vali, hold []mktval.Trade
	for _, t := range tr {
		switch t.Bucket {
		case "DISCOVERY":
			disc = append(disc, t)
		case "VALIDATION":
			vali = append(vali, t)
		case "HOLDOUT":
			hold = append(hold, t)
		}
	}
	r.DiscoveryN, r.ValidationN, r.HoldoutN = len(disc), len(vali), len(hold)
	hs := mktval.Summarize(hold)
	if hs.N == 0 {
		hs = mktval.Summarize(tr)
	}
	zs := mktval.Summarize(zero)
	r.Expectancy, r.Hit, r.ProfitFactor, r.MaxDD = hs.Expectancy, hs.Hit, hs.PF, hs.MaxDD
	r.MFE, r.MAE, r.Long, r.Short = hs.MFE, hs.MAE, hs.Long, hs.Short
	r.PositiveSubs = hs.PositiveThirds
	r.CostDrag = zs.Expectancy - hs.Expectancy
	r.AttentionNote = "V1 frozen; completeness"
	r.SalienceNote = "V0 frozen; unusual now"
	brokerOK := f.broker == "SPEC_READY" || f.broker == "READ"
	r.Status = mktval.DecideStatus(r, hs, len(f.m15) >= 80, brokerOK)
	if r.Status == mktval.ShadowValidated && f.monetary == money.RuntimeValidated && brokerOK {
		r.DemoElig = "DEMO_CALIBRATION_READY"
	}
	return r
}

func hashCandles(cs []models.Candle) string {
	h := sha256.New()
	for _, c := range cs {
		fmt.Fprintf(h, "%s %.8f %.8f %.8f %.8f\n", c.Time.UTC().Format(time.RFC3339Nano), c.Open, c.High, c.Low, c.Close)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
