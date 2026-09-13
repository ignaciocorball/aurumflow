package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"aurumflow/internal/binanceusdm"
	"aurumflow/internal/logger"
	"aurumflow/internal/md"
	"aurumflow/internal/radar"
)

type soakMetrics struct {
	DurationSeconds int64   `json:"duration_seconds"`
	EventsReceived  int64   `json:"events_received"`
	EventsDropped   int64   `json:"events_dropped"`
	Trades          int64   `json:"trades"`
	BookDeltas      int64   `json:"book_deltas"`
	BookResyncs     int     `json:"book_resyncs"`
	Reconnects      int     `json:"reconnects"`
	BookSynced      bool    `json:"book_synced"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	P95LatencyMs    float64 `json:"p95_latency_ms"`
	FeatureUpdates  int64   `json:"feature_updates"`
	CVD             float64 `json:"cvd"`
	Pressure        float64 `json:"pressure"`
	RadarState      string  `json:"radar_state"`
	Confidence      float64 `json:"confidence"`
	Panicked        bool    `json:"panicked"`
	RadarMutations  int     `json:"radar_mutations"`
}

func runShadowSoak(ctx context.Context, dur time.Duration) {
	if dur <= 0 {
		dur = time.Minute
	}
	logger.Info("SHADOW soak starting duration=%s provider=binance_usdm_public symbol=BTCUSDT mutations=0", dur)
	bus := md.NewBus(4096)
	ad := binanceusdm.New()
	eng := radar.NewEngine("BTCUSDT")
	eng.Book = ad.Book
	eng.Flow = ad.Flow

	runCtx, cancel := context.WithTimeout(ctx, dur)
	defer cancel()
	go func() {
		if err := ad.Run(runCtx, bus); err != nil && runCtx.Err() == nil {
			logger.Warn("binance adapter: %v", err)
		}
	}()

	var trades, deltas, features int64
	var latSum float64
	var lats []float64
	started := time.Now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	radarPath := filepath.Join("journals", "radar.jsonl")
	_ = os.MkdirAll("journals", 0o755)
	rf, _ := os.OpenFile(radarPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if rf != nil {
		defer rf.Close()
	}

	for {
		select {
		case <-runCtx.Done():
			sent, drop := bus.Stats()
			m := soakMetrics{
				DurationSeconds: int64(time.Since(started).Seconds()),
				EventsReceived:  sent,
				EventsDropped:   drop,
				Trades:          trades,
				BookDeltas:      deltas,
				BookResyncs:     ad.Book.Resyncs,
				Reconnects:      ad.Reconnects,
				BookSynced:      ad.Book.Synced,
				FeatureUpdates:  features,
				CVD:             ad.Flow.CVD(),
				Pressure:        eng.Last.PressureScore,
				RadarState:      eng.Last.State,
				Confidence:      eng.Last.Confidence,
				RadarMutations:  0,
			}
			if n := len(lats); n > 0 {
				m.AvgLatencyMs = latSum / float64(n)
				m.P95LatencyMs = percentile(lats, 0.95)
			}
			out := filepath.Join("journals", "soak-metrics.json")
			raw, _ := json.MarshalIndent(m, "", "  ")
			_ = os.WriteFile(out, raw, 0o644)
			logger.Info("SHADOW soak done events=%d drop=%d trades=%d deltas=%d resyncs=%d reconnects=%d synced=%v avg_lat_ms=%.1f p95_ms=%.1f cvd=%.4f pressure=%.1f state=%s mutations=%d",
				m.EventsReceived, m.EventsDropped, m.Trades, m.BookDeltas, m.BookResyncs, m.Reconnects, m.BookSynced, m.AvgLatencyMs, m.P95LatencyMs, m.CVD, m.Pressure, m.RadarState, m.RadarMutations)
			if !radar.MayMutateBroker(radar.ModeShadow) {
				logger.Info("Radar SHADOW safety: no execution mutations")
			}
			return
		case ev := <-bus.C():
			lat := ev.ReceiveTime.Sub(ev.EventTime).Seconds() * 1000
			if lat > 0 && lat < 60_000 {
				latSum += lat
				lats = append(lats, lat)
			}
			switch ev.Kind {
			case md.KindTrade:
				trades++
				if tr, ok := ev.Payload.(md.Trade); ok {
					eng.OnTrade(tr.Price, tr.Qty, tr.BuyerMaker)
				}
			case md.KindBookDelta, md.KindBookSnapshot:
				deltas++
			}
		case <-ticker.C:
			snap := eng.Snapshot(time.Now().UTC(), ad.Book.Synced, 1)
			features++
			if rf != nil {
				row, _ := json.Marshal(map[string]any{
					"event": "radar_state", "ts": time.Now().UTC().Format(time.RFC3339),
					"state": snap.State, "pressure": snap.PressureScore, "confidence": snap.Confidence,
					"book_synced": snap.BookSynced, "cvd": ad.Flow.CVD(),
				})
				_, _ = rf.Write(append(row, '\n'))
			}
		}
	}
}

func percentile(xs []float64, p float64) float64 {
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
