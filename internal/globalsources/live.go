package globalsources

import (
	"context"
	"os"
	"strings"
	"time"

	"aurumflow/internal/cftc"
	"aurumflow/internal/worlddomain"
)

type FetchResult struct {
	Provider string
	OK       bool
	Origin   worlddomain.Origin
	Rows     int
	Latest   time.Time
	Error    string
	URL      string
	Hash     string
}

type FedClient interface {
	SeriesCSV(ctx context.Context, seriesID string) ([]byte, string, error)
}

type fredClient struct{ cache *HTTPCache }

func (f fredClient) SeriesCSV(ctx context.Context, seriesID string) ([]byte, string, error) {
	url := "https://fred.stlouisfed.org/graph/fredgraph.csv?id=" + seriesID
	raw, via, err := f.cache.GetFresh(ctx, url, "fred-"+seriesID+".csv")
	return raw, via + " " + url, err
}

func originFromVia(via string) worlddomain.Origin {
	if strings.Contains(via, "network") {
		return worlddomain.OriginLive
	}
	return worlddomain.OriginCache
}

func FetchOfficial(ctx context.Context) ([]worlddomain.ContextObservation, []FetchResult) {
	now := time.Now().UTC()
	cache := NewHTTPCache(RawRoot)
	cache.MinGap = 1500 * time.Millisecond
	fed := fredClient{cache: cache}
	var all []worlddomain.ContextObservation
	var res []FetchResult
	add := func(xs []worlddomain.ContextObservation, r FetchResult) {
		res = append(res, r)
		all = append(all, xs...)
	}
	add(fetchFed(ctx, fed, now))
	add(fetchTIC(ctx, cache, now))
	add(fetchICI(ctx, cache, now))
	add(fetchCboe(ctx, cache, now))
	add(fetchECB(ctx, cache, now))
	add(fetchBIS(ctx, cache, now))
	add(fetchJPX(ctx, cache, now))
	add(fetchCFTC(ctx, now))
	add(fetchWGC(ctx, cache, now))
	add(fetchIShares(ctx, cache, now))
	if os.Getenv("EIA_API_KEY") != "" {
		add(fetchEIA(ctx, cache, now))
	} else {
		res = append(res, FetchResult{Provider: "EIA", Error: "PENDING_FREE_KEY"})
	}
	res = append(res, FetchResult{Provider: "HKEX", Error: HKEXStatus()})
	return all, res
}

func fetchFed(ctx context.Context, fed FedClient, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	h := NewHTTPCache(RawRoot)
	h.Client.Timeout = 90 * time.Second
	url := "https://www.federalreserve.gov/releases/h41/current/h41.htm"
	raw, via, err := h.GetFresh(ctx, url, "fed-h41.htm")
	if err == nil && len(raw) > 0 {
		_, hash, _ := WriteRaw("fed", "h41.htm", raw)
		obs := StampOrigin(ParseFedH41HTML(raw, now), originFromVia(via), hash)
		if len(obs) > 0 {
			latest := latestObs(obs)
			_ = WriteNormalized("FED_H41", obs, CacheMeta{
				Provider: "FED_H41", OfficialSource: url, DatasetID: "H.4.1",
				RetrievedAt: now, LatestObservation: formatDay(latest), LatestPublication: formatDay(latest),
				Rows: len(obs), Hash: hash, Origin: string(originFromVia(via)), Quality: "SLOW_CONTEXT", Freshness: "WEEKLY", ParserVersion: "h41-html-v1",
			})
			return obs, FetchResult{Provider: "FED_H41", OK: true, Origin: originFromVia(via), Rows: len(obs), Latest: latest, URL: url, Hash: hash}
		}
	}
	ids := []string{"WALCL", "WRESBAL", "WTREGEN", "RRPONTSYD"}
	var obs []worlddomain.ContextObservation
	var hash string
	via = ""
	urls := ""
	for _, id := range ids {
		b, v, ferr := fed.SeriesCSV(ctx, id)
		if ferr != nil && len(b) == 0 {
			return loadCache("FED_H41", FetchResult{Provider: "FED_H41", Error: ferr.Error()})
		}
		via = v
		urls += v + " "
		part := ParseFredSeries(id, b, now)
		obs = append(obs, part...)
		hash = HashBytes(b)
		_, _, _ = WriteRaw("fed", "fred-"+id+".csv", b)
	}
	origin := originFromVia(via)
	obs = StampOrigin(obs, origin, hash)
	latest := latestObs(obs)
	_ = WriteNormalized("FED_H41", obs, CacheMeta{
		Provider: "FED_H41", OfficialSource: "https://fred.stlouisfed.org/series/WALCL", DatasetID: "H.4.1/FRED",
		RetrievedAt: now, LatestObservation: latest.Format("2006-01-02"), LatestPublication: latest.Format("2006-01-02"),
		Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "SLOW_CONTEXT", Freshness: "WEEKLY", ParserVersion: "v1",
	})
	return obs, FetchResult{Provider: "FED_H41", OK: len(obs) > 0, Origin: origin, Rows: len(obs), Latest: latest, URL: strings.TrimSpace(urls), Hash: hash}
}

func fetchTIC(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	type item struct{ url, name string }
	items := []item{
		{"https://ticdata.treasury.gov/resource-center/data-chart-center/tic/Documents/slt_table1.txt", "slt_table1.txt"},
		{"https://ticdata.treasury.gov/resource-center/data-chart-center/tic/Documents/slt_table5.txt", "slt_table5.txt"},
	}
	var obs []worlddomain.ContextObservation
	var hash, via, first string
	for _, it := range items {
		raw, v, err := h.GetFresh(ctx, it.url, "tic-"+it.name)
		if err != nil && len(raw) == 0 {
			continue
		}
		if first == "" {
			first, via = it.url, v
		}
		_, hash, _ = WriteRaw("tic", it.name, raw)
		obs = append(obs, StampOrigin(ParseTICAny(raw, now), originFromVia(v), HashBytes(raw))...)
	}
	if len(obs) == 0 {
		return loadCache("TIC", FetchResult{Provider: "TIC", Error: "no official SLT tables", URL: items[0].url})
	}
	origin := originFromVia(via)
	var keep []worlddomain.ContextObservation
	for _, o := range obs {
		if !o.AvailableAt.After(now) && !o.PublishedAt.After(now) {
			keep = append(keep, o)
		}
	}
	obs = keep
	latest := latestObs(obs)
	next := time.Date(now.Year(), now.Month(), 16, 0, 0, 0, 0, time.UTC)
	if !next.After(now) {
		next = next.AddDate(0, 1, 0)
	}
	_ = WriteNormalized("TIC", obs, CacheMeta{
		Provider: "TIC", OfficialSource: first, DatasetID: "slt_table1+slt_table5", RetrievedAt: now,
		LatestObservation: latest.Format("2006-01"), LatestPublication: latest.Format("2006-01-02"),
		Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "SLOW_CONTEXT", Freshness: "MONTHLY", ParserVersion: "slt-v1",
		NextExpected: next.Format("2006-01-02"),
	})
	return obs, FetchResult{Provider: "TIC", OK: len(obs) > 0, Origin: origin, Rows: len(obs), Latest: latest, URL: first, Hash: hash}
}

func fetchICI(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	url := "https://www.ici.org/research/stats/weekly_trends"
	raw, via, err := h.GetFresh(ctx, url, "ici-weekly.html")
	if err != nil && len(raw) == 0 {
		return loadCache("ICI", FetchResult{Provider: "ICI", Error: err.Error(), URL: url})
	}
	_, hash, _ := WriteRaw("ici", "weekly.html", raw)
	origin := originFromVia(via)
	obs := StampOrigin(ParseICIHTML(raw, now), origin, hash)
	latest := latestObs(obs)
	_ = WriteNormalized("ICI", obs, CacheMeta{
		Provider: "ICI", OfficialSource: url, DatasetID: "weekly_trends", RetrievedAt: now,
		LatestObservation: formatDay(latest), LatestPublication: formatDay(latest),
		Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "SLOW_CONTEXT", Freshness: "WEEKLY", ParserVersion: "v1",
	})
	return obs, FetchResult{Provider: "ICI", OK: len(obs) > 0, Origin: origin, Rows: len(obs), Latest: latest, URL: url, Hash: hash}
}

func fetchCboe(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	type item struct{ url, name, metric string }
	items := []item{
		{"https://cdn.cboe.com/resources/options/volume_and_call_put_ratios/totalpc.csv", "totalpc.csv", "TOTAL_PC"},
		{"https://cdn.cboe.com/resources/options/volume_and_call_put_ratios/equitypc.csv", "equitypc.csv", "EQUITY_PC"},
		{"https://cdn.cboe.com/resources/options/volume_and_call_put_ratios/indexpc.csv", "indexpc.csv", "INDEX_PC"},
	}
	var obs []worlddomain.ContextObservation
	var hash, via, firstURL string
	for _, it := range items {
		raw, v, err := h.GetFresh(ctx, it.url, "cboe-"+it.name)
		if err != nil && len(raw) == 0 {
			continue
		}
		if firstURL == "" {
			firstURL = it.url
			via = v
		}
		_, hash, _ = WriteRaw("cboe", it.name, raw)
		obs = append(obs, StampOrigin(ParseCboeOfficial(raw, it.metric, now), originFromVia(v), HashBytes(raw))...)
	}
	if len(obs) == 0 {
		return loadCache("CBOE", FetchResult{Provider: "CBOE", Error: "no official put/call files", URL: items[0].url})
	}
	origin := originFromVia(via)
	latest := latestObs(obs)
	if latest.Before(now.AddDate(-1, 0, 0)) {
		_ = WriteNormalized("CBOE", obs, CacheMeta{
			Provider: "CBOE", OfficialSource: firstURL, DatasetID: "put_call_ratios", RetrievedAt: now,
			LatestObservation: formatDay(latest), LatestPublication: formatDay(latest),
			Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "HISTORICAL_FILE_NOT_CURRENT", Freshness: "DAILY", ParserVersion: "v1",
		})
		return nil, FetchResult{Provider: "CBOE", OK: false, Origin: origin, Rows: 0, Latest: latest, URL: firstURL, Hash: hash, Error: "public historical file ends " + formatDay(latest) + "; current daily not machine-readable"}
	}
	_ = WriteNormalized("CBOE", obs, CacheMeta{
		Provider: "CBOE", OfficialSource: firstURL, DatasetID: "put_call_ratios", RetrievedAt: now,
		LatestObservation: formatDay(latest), LatestPublication: formatDay(latest),
		Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "SLOW_CONTEXT", Freshness: "DAILY", ParserVersion: "v1",
	})
	return obs, FetchResult{Provider: "CBOE", OK: true, Origin: origin, Rows: len(obs), Latest: latest, URL: firstURL, Hash: hash}
}

func fetchECB(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	url := "https://data-api.ecb.europa.eu/service/data/FM/B.U2.EUR.4F.KR.MRR_FR.LEV?format=csvdata"
	raw, via, err := h.GetFresh(ctx, url, "ecb-mrr.csv")
	if err != nil && len(raw) == 0 {
		return loadCache("ECB", FetchResult{Provider: "ECB", Error: err.Error(), URL: url})
	}
	_, hash, _ := WriteRaw("ecb", "mrr.csv", raw)
	origin := originFromVia(via)
	obs := StampOrigin(ParseECBSDMX(raw, now), origin, hash)
	latest := latestObs(obs)
	_ = WriteNormalized("ECB", obs, CacheMeta{
		Provider: "ECB", OfficialSource: url, DatasetID: "FM/B.U2.EUR.4F.KR.MRR_FR.LEV", RetrievedAt: now,
		LatestObservation: formatDay(latest), LatestPublication: formatDay(latest),
		Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "SLOW_CONTEXT", Freshness: "WEEKLY", ParserVersion: "v1",
	})
	return obs, FetchResult{Provider: "ECB", OK: len(obs) > 0, Origin: origin, Rows: len(obs), Latest: latest, URL: url, Hash: hash}
}

func fetchBIS(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	url := "https://stats.bis.org/api/v2/data/dataflow/BIS/WS_GLI/1.0/.USD?format=csv"
	raw, via, err := h.GetFresh(ctx, url, "bis-gli.csv")
	if err != nil && len(raw) == 0 {
		return loadCache("BIS", FetchResult{Provider: "BIS", Error: err.Error(), URL: url})
	}
	_, hash, _ := WriteRaw("bis", "gli.csv", raw)
	origin := originFromVia(via)
	obs := StampOrigin(ParseBISOfficial(raw, now), origin, hash)
	latest := latestObs(obs)
	_ = WriteNormalized("BIS", obs, CacheMeta{
		Provider: "BIS", OfficialSource: url, DatasetID: "WS_GLI", RetrievedAt: now,
		LatestObservation: formatDay(latest), LatestPublication: formatDay(latest),
		Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "SLOW_CONTEXT", Freshness: "QUARTERLY", ParserVersion: "v1",
	})
	return obs, FetchResult{Provider: "BIS", OK: len(obs) > 0, Origin: origin, Rows: len(obs), Latest: latest, URL: url, Hash: hash}
}

func fetchJPX(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	url := "https://www.jpx.co.jp/english/markets/statistics-equities/investor-type/index.html"
	raw, via, err := h.GetFresh(ctx, url, "jpx-index.html")
	if err != nil && len(raw) == 0 {
		return loadCache("JPX", FetchResult{Provider: "JPX", Error: err.Error(), URL: url})
	}
	_, hash, _ := WriteRaw("jpx", "index.html", raw)
	origin := originFromVia(via)
	obs := StampOrigin(ParseJPXAny(raw, now), origin, hash)
	if len(obs) == 0 {
		obs = StampOrigin(ParseJPXIndexWeeks(raw, now), origin, hash)
	}
	latest := latestObs(obs)
	_ = WriteNormalized("JPX", obs, CacheMeta{
		Provider: "JPX", OfficialSource: url, DatasetID: "investor-type", RetrievedAt: now,
		LatestObservation: formatDay(latest), LatestPublication: formatDay(latest),
		Rows: len(obs), Hash: hash, Origin: string(origin), Quality: "SLOW_CONTEXT", Freshness: "WEEKLY", ParserVersion: "v1-2026",
	})
	return obs, FetchResult{Provider: "JPX", OK: len(obs) > 0, Origin: origin, Rows: len(obs), Latest: latest, URL: url, Hash: hash}
}

func fetchCFTC(ctx context.Context, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	rows, err := cftc.FetchCurrent(ctx)
	if err != nil {
		return loadCache("CFTC", FetchResult{Provider: "CFTC", Error: err.Error(), URL: cftc.DisaggURL})
	}
	var obs []worlddomain.ContextObservation
	for _, r := range rows {
		id, ok := cftc.CanonicalMarket(r.Market)
		if !ok {
			continue
		}
		o := Obs("CFTC", id+"_MM_NET", r.MMNet, true, r.AsOf, r.Available, worlddomain.FreqWeekly, worlddomain.RegionGlobal, worlddomain.AssetPrecious, cftc.DisaggURL)
		o.Origin = worlddomain.OriginLive
		o.DatasetID = "f_disagg"
		obs = append(obs, o)
	}
	hash := HashBytes([]byte("cftc-live"))
	_ = WriteNormalized("CFTC", obs, CacheMeta{
		Provider: "CFTC", OfficialSource: cftc.DisaggURL, DatasetID: "f_disagg", RetrievedAt: now,
		LatestObservation: formatDay(latestObs(obs)), Rows: len(obs), Hash: hash,
		Origin: string(worlddomain.OriginLive), Quality: "SLOW_CONTEXT", Freshness: "WEEKLY", ParserVersion: "reuse",
	})
	return obs, FetchResult{Provider: "CFTC", OK: len(obs) > 0, Origin: worlddomain.OriginLive, Rows: len(obs), Latest: latestObs(obs), URL: cftc.DisaggURL, Hash: hash}
}

func fetchWGC(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	url := "https://www.gold.org/"
	raw, via, err := h.GetFresh(ctx, url, "wgc.html")
	if err != nil && len(raw) == 0 {
		return loadCache("WGC", FetchResult{Provider: "WGC", Error: "CACHE_MANUAL_OR_PUBLIC", URL: url})
	}
	_, hash, _ := WriteRaw("wgc", "index.html", raw)
	obs := StampOrigin(ParseWGC(raw, now), originFromVia(via), hash)
	if len(obs) == 0 {
		return nil, FetchResult{Provider: "WGC", Error: "CACHE_MANUAL_OR_PUBLIC", URL: url}
	}
	return obs, FetchResult{Provider: "WGC", OK: true, Origin: originFromVia(via), Rows: len(obs), Latest: latestObs(obs), URL: url, Hash: hash}
}

func fetchIShares(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	// Official product pages exist; holdings CSV identity is verified per ticker before use.
	return nil, FetchResult{Provider: "ISHARES", Error: "PENDING_VERIFIED_FUND_IDENTITY"}
}

func fetchEIA(ctx context.Context, h *HTTPCache, now time.Time) ([]worlddomain.ContextObservation, FetchResult) {
	url := "https://www.eia.gov/dnav/pet/xls/PET_STOC_WSTK_DCU_NUS_W.xls"
	raw, via, err := h.GetFresh(ctx, url, "eia-stocks.xls")
	if err != nil && len(raw) == 0 {
		return nil, FetchResult{Provider: "EIA", Error: err.Error(), URL: url}
	}
	_, hash, _ := WriteRaw("eia", "PET_STOC_WSTK_DCU_NUS_W.xls", raw)
	obs := StampOrigin(ParseEIA(raw, now), originFromVia(via), hash)
	return obs, FetchResult{Provider: "EIA", OK: len(obs) > 0, Origin: originFromVia(via), Rows: len(obs), Latest: latestObs(obs), URL: url, Hash: hash}
}

func loadCache(provider string, r FetchResult) ([]worlddomain.ContextObservation, FetchResult) {
	xs, meta, err := LoadNormalized(provider)
	if err != nil {
		return nil, r
	}
	r.OK = len(xs) > 0
	r.Origin = worlddomain.OriginCache
	r.Rows = len(xs)
	r.Latest = ParseDate(meta.LatestObservation)
	if r.Error != "" {
		r.Error = r.Error + "; using CACHE_OFFICIAL"
	}
	return xs, r
}

func latestObs(xs []worlddomain.ContextObservation) time.Time {
	var t time.Time
	for _, o := range xs {
		if o.ObservedAt.After(t) {
			t = o.ObservedAt
		}
	}
	return t
}

func formatDay(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func errString(err error) string {
	if err == nil {
		return "empty body"
	}
	return err.Error()
}
