package research

import (
	"context"
	"runtime"
	"time"

	"aurumflow/internal/binancehist"
	"aurumflow/internal/candles"
	"aurumflow/internal/quality"
	"aurumflow/internal/radar"
	"aurumflow/pkg/models"
)

type BTCResult struct {
	From, To           time.Time
	Days               int
	TradeEvents        int
	RadarSnaps         int
	LegacySignals      int
	LegacyLong         int
	LegacyShort        int
	Aligned            int
	Contradicted       int
	RadarExpansion     int
	EventsPerSec       float64
	Runtime            time.Duration
	MemMB              float64
	LegacyAll          BucketStats
	LegacyAligned      BucketStats
	LegacyContradicted BucketStats
	RadarAll           BucketStats
	FusionAligned      BucketStats
	FilterAll          BucketStats
	FilterExcl         BucketStats
	Pressure           []BucketStats
	Confidence         []BucketStats
	States             []BucketStats
	Split              Split
	LegacyOOS          BucketStats
	RadarOOS           BucketStats
	FusionOOS          BucketStats
	LegacyLongStudy    BucketStats
	LegacyShortStudy   BucketStats
	Horizons           []HorizonBlock
	Quality            quality.Report
	Horizon            time.Duration
	Note10s30s         string
}

type HorizonBlock struct {
	Horizon      string      `json:"horizon"`
	Legacy       BucketStats `json:"legacy"`
	Aligned      BucketStats `json:"aligned"`
	Contradicted BucketStats `json:"contradicted"`
	Radar        BucketStats `json:"radar"`
}

func ProcessDayZip(zipPath string, symbol string, e *radar.Engine, q *quality.Report, m1 *[]models.Candle, snaps *[]radar.PressureSnapshot) error {
	var pending []candles.Trade
	var lastMin time.Time
	return binancehist.IterZipCSV(zipPath, func(rec []string) error {
		if binancehist.IsHeaderRow(rec) {
			return nil
		}
		tr, err := binancehist.ParseAggRow(rec)
		if err != nil {
			q.InvalidPrice++
			return nil
		}
		q.Rows++
		e.OnTrade(tr.Price, tr.Qty, tr.BuyerMaker)
		pending = append(pending, candles.Trade{T: tr.Time, Price: tr.Price, Qty: tr.Qty})
		min := tr.Time.Truncate(time.Minute)
		if lastMin.IsZero() {
			lastMin = min
		} else if min.After(lastMin) {
			old := pending[:len(pending)-1]
			*m1 = append(*m1, candles.Synthesize(old, time.Minute)...)
			pending = pending[len(pending)-1:]
			*snaps = append(*snaps, e.Snapshot(lastMin, true, 1))
			lastMin = min
		}
		return nil
	})
}

func FinalizeBTC(m1 []models.Candle, snaps []radar.PressureSnapshot, q quality.Report, started time.Time, events int) BTCResult {
	m15 := Resample(m1, 15*time.Minute)
	h1 := Resample(m1, time.Hour)
	h4 := Resample(m1, 4*time.Hour)
	legacy := ScanLegacy(m15, h1, h4, CryptoResearchConfig())
	rp := ToRadarPoints(snaps)
	fused := AlignFusion(legacy, rp, 15)
	prices := CandlePrices(m1)
	h := 15 * time.Minute
	res := BTCResult{Horizon: h, Quality: q, TradeEvents: events, RadarSnaps: len(snaps)}
	if len(m1) > 0 {
		res.From, res.To = m1[0].Time, m1[len(m1)-1].Time
		res.Days = int(res.To.Sub(res.From).Hours()/24) + 1
	}
	res.LegacySignals = len(legacy)
	for _, s := range legacy {
		if s.Direction > 0 {
			res.LegacyLong++
		} else {
			res.LegacyShort++
		}
	}
	for _, s := range fused {
		switch s.Class {
		case AlignedLong, AlignedShort:
			res.Aligned++
		case ContradictedLong, ContradictedShort:
			res.Contradicted++
		}
	}
	for _, s := range snaps {
		if s.State == radar.StateExpansion {
			res.RadarExpansion++
		}
	}
	radarRows := make([]SignalRow, 0, len(snaps))
	for _, s := range snaps {
		if s.Direction == 0 {
			continue
		}
		radarRows = append(radarRows, SignalRow{Time: s.Timestamp, Direction: s.Direction, Pressure: s.PressureScore, Confidence: s.Confidence, RadarState: s.State, Source: "radar"})
	}
	res.LegacyAll = StudySignals(fused, prices, h, nil)
	res.LegacyAligned = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.Class == AlignedLong || r.Class == AlignedShort })
	res.LegacyContradicted = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.Class == ContradictedLong || r.Class == ContradictedShort })
	res.RadarAll = StudySignals(radarRows, prices, h, nil)
	res.FusionAligned = res.LegacyAligned
	res.LegacyLongStudy = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.Direction > 0 })
	res.LegacyShortStudy = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.Direction < 0 })
	res.FilterAll, res.FilterExcl = FilterValue(fused, prices, h)
	res.Pressure = PressureBuckets(fused, prices, h)
	res.Confidence = ConfidenceBuckets(fused, prices, h)
	res.States = StateBuckets(fused, prices, h)
	res.Note10s30s = "RESOLUTION_LIMITED: labels use 1m synthesized candles; +10s/+30s are not independent of +1m"
	for _, hz := range []time.Duration{time.Minute, 5 * time.Minute, 15 * time.Minute, 30 * time.Minute, time.Hour} {
		res.Horizons = append(res.Horizons, HorizonBlock{
			Horizon:      hz.String(),
			Legacy:       StudySignals(fused, prices, hz, nil),
			Aligned:      StudySignals(fused, prices, hz, func(r SignalRow) bool { return r.Class == AlignedLong || r.Class == AlignedShort }),
			Contradicted: StudySignals(fused, prices, hz, func(r SignalRow) bool { return r.Class == ContradictedLong || r.Class == ContradictedShort }),
			Radar:        StudySignals(radarRows, prices, hz, nil),
		})
	}
	if !res.From.IsZero() {
		res.Split = ProportionalSplit(res.From, res.To)
		res.LegacyOOS = StudySignals(fused, prices, h, func(r SignalRow) bool { return InRange(r.Time, res.Split.OOSFrom, res.Split.OOSTo) })
		res.RadarOOS = StudySignals(radarRows, prices, h, func(r SignalRow) bool { return InRange(r.Time, res.Split.OOSFrom, res.Split.OOSTo) })
		res.FusionOOS = StudySignals(fused, prices, h, func(r SignalRow) bool {
			return InRange(r.Time, res.Split.OOSFrom, res.Split.OOSTo) && (r.Class == AlignedLong || r.Class == AlignedShort)
		})
	}
	res.Runtime = time.Since(started)
	if res.Runtime > 0 {
		res.EventsPerSec = float64(events) / res.Runtime.Seconds()
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	res.MemMB = float64(ms.Alloc) / 1024 / 1024
	return res
}

func RunTape(ctx context.Context, paths []string, symbol string) (m1 []models.Candle, snaps []radar.PressureSnapshot, q quality.Report, events int, err error) {
	e := radar.NewTradeFlowEngine(symbol)
	for _, p := range paths {
		if ctx.Err() != nil {
			break
		}
		before := q.Rows
		if err = ProcessDayZip(p, symbol, e, &q, &m1, &snaps); err != nil {
			return nil, nil, q, events, err
		}
		events += q.Rows - before
	}
	q.Rows = events
	return m1, snaps, q, events, nil
}

func BuildFused(m1 []models.Candle, snaps []radar.PressureSnapshot) ([]SignalRow, []RadarPoint, []PricePoint) {
	m15 := Resample(m1, 15*time.Minute)
	h1 := Resample(m1, time.Hour)
	h4 := Resample(m1, 4*time.Hour)
	legacy := ScanLegacy(m15, h1, h4, CryptoResearchConfig())
	rp := ToRadarPoints(snaps)
	fused := AlignFusion(legacy, rp, DefaultAlignAbs)
	for i := range fused {
		fused[i].PreReturn = preReturn(CandlePrices(m1), fused[i].Time, 5*time.Minute)
	}
	return fused, rp, CandlePrices(m1)
}

func preReturn(prices []PricePoint, at time.Time, lookback time.Duration) float64 {
	now := priceAt(prices, at)
	prev := priceAt(prices, at.Add(-lookback))
	if prev == 0 {
		return 0
	}
	return (now - prev) / prev
}

func RunFromZips(ctx context.Context, paths []string, symbol string) (BTCResult, error) {
	started := time.Now()
	m1, snaps, q, events, err := RunTape(ctx, paths, symbol)
	if err != nil {
		return BTCResult{}, err
	}
	return FinalizeBTC(m1, snaps, q, started, events), nil
}
