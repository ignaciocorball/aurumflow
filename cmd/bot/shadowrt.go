package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"aurumflow/internal/absorption"
	"aurumflow/internal/binanceusdm"
	"aurumflow/internal/candles"
	"aurumflow/internal/book"
	"aurumflow/internal/bookfeatures"
	"aurumflow/internal/collector"
	"aurumflow/internal/exhaustion"
	"aurumflow/internal/l2proxy"
	"aurumflow/internal/logger"
	"aurumflow/internal/md"
	"aurumflow/internal/microflow"
	"aurumflow/internal/money"
	"aurumflow/internal/okxswap"
	"aurumflow/internal/ops"
	"aurumflow/internal/radar"
	"aurumflow/internal/rawbuf"
	"aurumflow/internal/research"
	"aurumflow/internal/shadow"
	"aurumflow/pkg/models"
)

func runShadowRuntime(ctx context.Context, dur time.Duration, statusAddr string, providerName string) {
	if dur <= 0 {
		dur = time.Hour
	}
	runCtx, cancel := context.WithTimeout(ctx, dur)
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	var prov md.MarketDataProvider
	var bk *book.Book
	instrument := "BITCOIN"
	venueInst := "BTCUSDT"
	relation := "CORRELATED_PROXY"
	v1Flow := "binance_usdm_public"
	l2Flow := "binance_usdm_public"
	if providerName == "" || providerName == "auto" {
		pr := binanceusdm.ProbeIncrementalDepth(runCtx, 12*time.Second)
		ox := okxswap.Diagnose(runCtx)
		if pr.Status == "OPERATIONAL" && pr.Deltas > 0 {
			providerName = "binance"
			logger.Info("auto L2 binance incremental deltas=%d selected=binance", pr.Deltas)
		} else if ox.Status == "OPERATIONAL" || ox.Connect == "ok" {
			providerName = "okx"
			logger.Info("auto L2 binance=%s okx=%s selected=okx", pr.Status, ox.Status)
		} else {
			providerName = "okx"
			logger.Info("auto L2 fallback okx (binance=%s)", pr.Status)
		}
	}
	switch providerName {
	case "okx", "okx_swap", "okx_swap_public":
		ad := okxswap.New()
		prov, bk = ad, ad.Book
		instrument = "BITCOIN"
		venueInst = okxswap.InstID
		relation = okxswap.Relation
		l2Flow = okxswap.Name
	default:
		ad := binanceusdm.New()
		prov, bk = ad, ad.Book
		venueInst = "BTCUSDT"
		relation = "CORRELATED_PROXY"
		l2Flow = "binance_usdm_public"
	}

	rt := &shadow.Runtime{
		Provider: prov, Instrument: instrument, VenueInst: venueInst, Relation: relation,
		Book: bk, Micro: microflow.New(), Exh: exhaustion.NewEngine(venueInst),
		BookFeat: bookfeatures.New(), Radar: radar.NewEngine(venueInst),
		Checkpoint: "data/live/shadow-checkpoint.json",
	}
	rt.Radar.Book = bk
	root := "data/live"
	rt.Trades = *collector.NewSegmented(root, prov.Name(), venueInst, "trades")
	rt.Depth = *collector.NewSegmented(root, prov.Name(), venueInst, "depth")
	rt.Snaps = *collector.NewSegmented(root, prov.Name(), venueInst, "snapshots")
	rt.Health = *collector.NewSegmented(root, prov.Name(), venueInst, "health")
	rt.LoadCheckpoint()
	defer rt.Close()

	st := ops.NewStatus()
	st.RadarMode = radar.ModeShadow
	st.RadarProvider = prov.Name()
	st.BookCapability = "BOOK_CAPABILITY_LIMITED"
	st.FeedFreshness = "PUBLIC_WSS_OR_REST"
	srv := ops.NewServer(statusAddr, st)
	go func() { _ = srv.ListenAndServe() }()
	refreshWorld(srv, false)

	bus := md.NewBus(8192)
	go func() { _ = prov.Run(runCtx, bus) }()
	if l2Flow != v1Flow {
		bn := binanceusdm.New()
		go func() { _ = bn.Run(runCtx, bus) }()
		logger.Info("V1FlowProvider=%s L2Provider=%s L2FlowProvider=%s relation=%s", v1Flow, prov.Name(), l2Flow, relation)
	}
	buf := rawbuf.New()
	basis := l2proxy.NewBasis(300)
	var lastBinanceMid, lastOKXMid float64

	var lats []float64
	var latSum float64
	started := time.Now()
	var pending []candles.Trade
	diskTick := 0
	var m1 []models.Candle
	seenLegacy := map[string]bool{}
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	logger.Info("shadow-runtime provider=%s instrument=%s relation=%s duration=%s NO execution provider", prov.Name(), venueInst, relation, dur)
	if shadow.HasExecutionProvider(rt) {
		logger.Error("compile invariant violated")
		os.Exit(1)
	}

	for {
		select {
		case <-runCtx.Done():
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			rt.Metrics.PeakMemMB = float64(ms.Alloc) / 1024 / 1024
			if n := len(lats); n > 0 {
				rt.Metrics.P50Ms = pct(lats, 0.5)
				rt.Metrics.P95Ms = pct(lats, 0.95)
			}
			raw, _ := json.MarshalIndent(rt.Metrics, "", "  ")
			_ = os.MkdirAll("journals", 0o755)
			_ = os.WriteFile("journals/shadow-runtime-metrics.json", raw, 0o644)
			logger.Info("shadow-runtime stop trades=%d deltas=%d snaps=%d gaps=%d resyncs=%d p50=%.1f p95=%.1f mem=%.1f",
				rt.Metrics.Trades, rt.Metrics.Deltas, rt.Metrics.Snapshots, rt.Metrics.Gaps, rt.Metrics.Resyncs, rt.Metrics.P50Ms, rt.Metrics.P95Ms, rt.Metrics.PeakMemMB)
			return
		case ev := <-bus.C():
			rt.Metrics.Events++
			buf.Add(ev)
			lat := ev.ReceiveTime.Sub(ev.EventTime).Seconds() * 1000
			if lat > 0 && lat < 60_000 {
				latSum += lat
				lats = append(lats, lat)
			}
			switch ev.Kind {
			case md.KindTrade:
				rt.Metrics.Trades++
				if tr, ok := ev.Payload.(md.Trade); ok {
					fromV1 := ev.Provider == v1Flow || ev.Venue == "binance_usdm"
					if fromV1 {
						rt.Micro.OnTrade(ev.EventTime, tr.Price, tr.Qty, tr.BuyerMaker)
						rt.Exh.OnTrade(ev.EventTime, tr.Price, tr.Qty, tr.BuyerMaker)
						rt.Radar.OnTrade(tr.Price, tr.Qty, tr.BuyerMaker)
						pending = append(pending, candles.Trade{T: ev.EventTime, Price: tr.Price, Qty: tr.Qty})
					}
					_ = rt.Trades.WriteEvent(ev.EventTime, ev.Seq, ev)
				}
			case md.KindBookDelta:
				rt.Metrics.Deltas++
				_ = rt.Depth.WriteEvent(ev.EventTime, ev.Seq, ev)
				if dlt, ok := ev.Payload.(md.BookDelta); ok && ev.Provider == prov.Name() {
					bl := make([]book.Level, len(dlt.Bids))
					al := make([]book.Level, len(dlt.Asks))
					for i, l := range dlt.Bids {
						bl[i] = book.Level{Price: l.Price, Qty: l.Qty}
					}
					for i, l := range dlt.Asks {
						al[i] = book.Level{Price: l.Price, Qty: l.Qty}
					}
					_, _ = rt.BookFeat.OnDelta(bk, ev.EventTime, bl, al)
				}
			case md.KindBookSnapshot:
				rt.Metrics.Snapshots++
				_ = rt.Snaps.WriteEvent(ev.EventTime, ev.Seq, ev)
				if ev.Provider == prov.Name() {
					_, _ = rt.BookFeat.OnSnapshot(bk, ev.EventTime)
				}
			case md.KindProviderHealth:
				_ = rt.Health.WriteEvent(ev.EventTime, ev.Seq, ev)
			}
		case now := <-tick.C:
			now = now.UTC()
			if len(pending) > 1 {
				built := candles.Synthesize(pending[:len(pending)-1], time.Minute)
				if len(built) > 0 {
					m1 = append(m1, built...)
					pending = pending[len(pending)-1:]
				}
			}
			newLegacy := 0
			if len(m1) > 80 {
				m15 := research.Resample(m1, 15*time.Minute)
				h1 := research.Resample(m1, time.Hour)
				h4 := research.Resample(m1, 4*time.Hour)
				for _, sig := range research.ScanLegacy(m15, h1, h4, research.CryptoResearchConfig()) {
					id := shadow.SignalID(venueInst, sig.Time, sig.Direction)
					if seenLegacy[id] || rt.SeenIDs[id] {
						continue
					}
					v1s := rt.Exh.Observe(sig.Time, sig.Direction, float64(sig.Score), rt.Radar.Last.PressureScore, 0)
					dd := bookfeatures.Observe(bk, now)
					ll := bookfeatures.Liquidity{}
					ab := absorption.Observe(v1s, dd.Available, 0, bookfeatures.Response(sig.Direction, dd, ll))
					_ = rt.RecordProspective(sig.Time, sig.Direction, sig.Score, v1s.PressureScore, v1s, rt.Micro.Snapshot(sig.Time), dd, ll, ab)
					seenLegacy[id] = true
					newLegacy = sig.Direction
				}
			}
			mf := rt.Micro.Snapshot(now)
			synced, gaps, resyncs := false, 0, 0
			if bk != nil {
				synced, _, gaps, resyncs = bk.Meta()
			}
			rs := rt.Radar.Snapshot(now, synced, 1)
			v1 := rt.Exh.Observe(now, 0, 0, rs.PressureScore, 0)
			d, liq := rt.BookFeat.OnBook(bk, now)
			if d.Mid > 0 {
				if l2Flow == okxswap.Name {
					lastOKXMid = d.Mid
				} else {
					lastBinanceMid = d.Mid
				}
			}
			bz, bzZ := 0.0, 0.0
			if lastBinanceMid > 0 && lastOKXMid > 0 {
				bz, bzZ = basis.Observe(lastBinanceMid, lastOKXMid)
			}
			pq := l2proxy.Classify(l2proxy.Input{
				BookSynced: synced, ProviderOK: true,
				BookAgeMs: func() int64 {
					if bk == nil {
						return 0
					}
					return bk.Age().Milliseconds()
				}(),
				ReceiveLatencyP95: rt.Metrics.P95Ms, AbsBasisZ: absf(bzZ),
			})
			plr := bookfeatures.Response(0, d, liq)
			plr.L2ProxyQuality = pq.Quality
			ab := absorption.Observe(v1, d.Available, 0, plr)
			cur := srv.Get()
			cur.UptimeSeconds = ops.AgeSeconds(started)
			cur.APIEnvironment = "DEMO"
			cur.CapitalEnv = "DEMO"
			cur.ExecutionMode = "SHADOW"
			cur.LivePossible = "IMPOSSIBLE / FAIL-CLOSED"
			cur.RadarMode = radar.ModeShadow
			cur.RadarShadow = "SHADOW"
			cur.ExhaustionShadow = "SHADOW"
			cur.AbsorptionShadow = "SHADOW"
			cur.RadarProvider = prov.Name()
			cur.LastPressure = rs.PressureScore
			cur.Pressure = rs.PressureScore
			cur.DirectionalP = v1.DirectionalPressure
			cur.LastV1Class = v1.Classification
			cur.ImpactFailure = v1.Features.ImpactFailure
			cur.FlowEfficiency = v1.Features.FlowEffNorm
			if bk != nil {
				cur.BookAgeMs = bk.Age().Milliseconds()
			}
			cur.BookSynced = synced
			cur.BookGaps = gaps
			cur.Resyncs = resyncs
			cur.V1FlowProvider = v1Flow
			cur.L2Provider = prov.Name()
			cur.PrimaryL2 = prov.Name()
			cur.L2Instrument = venueInst
			cur.L2Relation = relation
			cur.L2ProxyQuality = pq.Quality
			cur.Microprice = d.Microprice
			cur.L2Spread = d.Spread
			cur.L2Mid = d.Mid
			cur.L2QuotesOK = d.Available
			cur.Imb1, cur.Imb5, cur.Imb10, cur.Imb20 = d.Imb1, d.Imb5, d.Imb10, d.Imb20
			cur.CVD = rs.CVD
			if w, ok := mf.Windows["30s"]; ok {
				cur.AggBuy, cur.AggSell = w.BuyQty, w.SellQty
			}
			if bk != nil {
				bids, asks := bk.CopyTop(15)
				cur.TopBids = levelsFromMap(bids, true)
				cur.TopAsks = levelsFromMap(asks, false)
			}
			cur.BidRepl, cur.AskRepl = liq.BidReplenishment, liq.AskReplenishment
			cur.BidDepl, cur.AskDepl = liq.BidDepletion, liq.AskDepletion
			cur.BidPersist, cur.AskPersist = liq.BidPersistence, liq.AskPersistence
			cur.AbsorptionWhy = ab.Why
			cur.AbsorptionEvidence = evidenceLine(ab)
			cur.BTCPrice = d.Mid
			cur.Basis, cur.BasisZ = bz, bzZ
			cur.Deltas = rt.Metrics.Deltas
			cur.Snapshots = rt.Metrics.Snapshots
			cur.LastUpdate = now.Format(time.RFC3339)
			cur.EventFreshness = "PUBLIC_WSS"
			cur.BookFreshness = pq.Quality
			if !cur.BookSynced {
				cur.Alerts = appendUnique(cur.Alerts, "book_unsynced")
			}
			if v1.Classification == exhaustion.ClassExhaustion {
				cur.Alerts = appendUnique(cur.Alerts, "FLOW_EXHAUSTION")
			}
			if ab.Status == absorption.Supportive || ab.Status == absorption.StronglySupportive {
				cur.Alerts = appendUnique(cur.Alerts, "SUPPORTIVE_ABSORPTION")
			}
			cur.Vel1s = mf.Windows["1s"].TradesPerSec
			cur.Vel5s = mf.Windows["5s"].TradesPerSec
			cur.Vel30s = mf.Windows["30s"].TradesPerSec
			cur.AbsorptionStatus = ab.Status
			cur.SupportingRepl = plr.SupportingReplenishment
			cur.OpposingDepl = plr.OpposingDepletion
			cur.SupportingPers = plr.SupportingPersistence
			cur.PrimaryL2 = prov.Name()
			pst := research.ReadProspectiveStatus(research.ProspectiveDir)
			cur.ProspectiveTotal = pst.Signals
			cur.ProspectiveExh = pst.Exhaustion
			cur.ProspectiveCont = countClass(research.ProspectiveDir, exhaustion.ClassContinuation)
			cur.ProspectiveNeu = countClass(research.ProspectiveDir, exhaustion.ClassNeutral)
			cur.Mature15m = pst.Mature15m
			cur.Mature1h = pst.Mature1h
			if n := len(lats); n > 0 {
				cur.LatencyP50 = pct(lats, 0.5)
				cur.LatencyP95 = pct(lats, 0.95)
				cur.LatencyP99 = pct(lats, 0.99)
				rt.Metrics.P50Ms, rt.Metrics.P95Ms = cur.LatencyP50, cur.LatencyP95
			}
			if d.Available {
				cur.BookCapability = "BOOK_SYNCED"
			} else {
				cur.BookCapability = "BOOK_CAPABILITY_LIMITED"
			}
			if up := ops.AgeSeconds(started); up > 0 {
				cur.EventRate = float64(rt.Metrics.Events) / float64(up)
			}
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)
			cur.PeakMemMB = float64(ms.Alloc) / 1024 / 1024
			cur.Trades = rt.Metrics.Trades
			_, drop := bus.Stats()
			cur.Drops = drop
			if newLegacy != 0 {
				cur.LastLegacyDir = newLegacy
				if newLegacy > 0 {
					cur.LastStrategy = "LONG"
				} else {
					cur.LastStrategy = "SHORT"
				}
			}
			if g, ok := ops.ReadGoldObservatory(""); ok {
				cur.MarketStatus = g.MarketStatus
				if g.Validation != "" {
					cur.GoldValidation = g.Validation
				}
				if g.QuotesOK {
					cur.GoldBid, cur.GoldAsk, cur.GoldSpread = g.Bid, g.Ask, g.Spread
				}
				if g.DemoBalanceOK {
					cur.DemoBalance, cur.DemoBalanceOK = g.DemoBalance, true
				}
				if g.PositionsOK {
					cur.PositionsKnown = true
					cur.OpenPositions = g.OpenPositions
				}
			} else if spec, err := money.LoadSpec(money.CachePath("", "GOLD")); err == nil {
				cur.GoldValidation = spec.ValidationStatus
			}
			diskTick++
			if diskTick%15 == 1 {
				cur.DiskMB = dirSizeMB("data/live")
			}
			srv.Set(cur)
		}
	}
}

func countClass(dir, class string) int {
	ins, _ := research.ListInputs(dir)
	n := 0
	for _, in := range ins {
		if in.V1Classification == class {
			n++
		}
	}
	return n
}

func pct(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	cp := append([]float64(nil), xs...)
	for i := 1; i < len(cp); i++ {
		j := i
		for j > 0 && cp[j] < cp[j-1] {
			cp[j], cp[j-1] = cp[j-1], cp[j]
			j--
		}
	}
	idx := int(float64(len(cp)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cp) {
		idx = len(cp) - 1
	}
	return cp[idx]
}

func absf(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func appendUnique(xs []string, v string) []string {
	for _, x := range xs {
		if x == v {
			return xs
		}
	}
	return append(xs, v)
}

func levelsFromMap(m map[float64]float64, highFirst bool) []ops.BookLevel {
	type kv struct{ p, q float64 }
	xs := make([]kv, 0, len(m))
	for p, q := range m {
		xs = append(xs, kv{p, q})
	}
	for i := 1; i < len(xs); i++ {
		j := i
		for j > 0 {
			less := xs[j].p < xs[j-1].p
			if highFirst {
				less = xs[j].p > xs[j-1].p
			}
			if !less {
				break
			}
			xs[j], xs[j-1] = xs[j-1], xs[j]
			j--
		}
	}
	out := make([]ops.BookLevel, len(xs))
	for i, x := range xs {
		out[i] = ops.BookLevel{Price: x.p, Qty: x.q}
	}
	return out
}

func dirSizeMB(root string) float64 {
	var n int64
	_ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			n += info.Size()
		}
		return nil
	})
	return float64(n) / 1024 / 1024
}

func evidenceLine(ab absorption.Snapshot) string {
	out := ""
	for _, e := range ab.Evidence {
		if out != "" {
			out += " | "
		}
		out += e.Name + "=" + e.Level
	}
	return out
}
