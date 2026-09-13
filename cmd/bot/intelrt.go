package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"aurumflow/config"
	"aurumflow/internal/featready"
	"aurumflow/internal/instrument"
	"aurumflow/internal/intelrt"
	"aurumflow/internal/livesurface"
	"aurumflow/internal/logger"
	"aurumflow/internal/market"
	"aurumflow/internal/mktwarmup"
	"aurumflow/internal/ops"
	"aurumflow/pkg/models"
)

var (
	intelOn   bool
	histMu    sync.Mutex
	histByMkt = map[string]mktwarmup.History{}
	featByMkt = map[string]featready.Snapshot{}
	prevStatus = map[string]string{}
	mapCache  []instrument.Mapping
	lastFrame livesurface.Frame
	lastFrameOK bool
	lastQuoteAt time.Time
	goldPeer  map[string]any
	soak      intelSoak
)

type intelSoak struct {
	mu             sync.Mutex
	Started        time.Time
	QuoteUpdates   int
	StaleFrames    int
	HistoryFetches int
	WorldUpdates   int
	OppChanges     int
	Proposals      int
	LegacyEvals    int
	Errors         int
	MarketsOpen    int
}

func runIntelligenceRuntime(ctx context.Context, soakMin int, statusAddr, l2Provider string) {
	intelOn = true
	if statusAddr == "" || statusAddr == "127.0.0.1:8765" {
		statusAddr = intelrt.Port
	}
	dur := time.Duration(soakMin) * time.Minute
	if soakMin <= 1 {
		dur = 24 * time.Hour
		logger.Info("intelligence-runtime PREOPEN_READY · soak unbounded until signal (24h cap) · port=%s · no ExecutionProvider · MULTI_MARKET_DEMO_EXECUTION=OFF", statusAddr)
	} else {
		logger.Info("intelligence-runtime soak=%dm port=%s · no ExecutionProvider · MULTI_MARKET_DEMO_EXECUTION=OFF", soakMin, statusAddr)
	}
	soak.Started = time.Now().UTC()
	go warmupLoop(ctx)
	go pollGoldDemo(ctx)
	runShadowRuntime(ctx, dur, statusAddr, l2Provider)
}

func warmupLoop(ctx context.Context) {
	tick := time.NewTicker(2 * time.Minute)
	defer tick.Stop()
	runWarmupOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			runWarmupOnce(ctx)
		}
	}
}

func runWarmupOnce(ctx context.Context) {
	if !demoEnvConfigured() {
		return
	}
	maps, err := instrument.LoadMappings("data/world/normalized/capital/mappings.json")
	if err != nil || len(maps) == 0 {
		return
	}
	histMu.Lock()
	mapCache = maps
	histMu.Unlock()
	sess, err := bootstrapDemoSession(ctx, config.ExecutionDisabled)
	if err != nil {
		bumpErr()
		return
	}
	now := time.Now().UTC()
	for _, m := range maps {
		if m.Epic == "" || m.Unresolved {
			continue
		}
		h, ferr := fetchHistory(ctx, sess.client, m.Canonical, m.Epic, now)
		histMu.Lock()
		histByMkt[m.Canonical] = h
		histMu.Unlock()
		soak.mu.Lock()
		soak.HistoryFetches++
		soak.mu.Unlock()
		if ferr != nil && h.Status != mktwarmup.StatusUnavailable {
			bumpErr()
		}
	}
}

func fetchHistory(ctx context.Context, c *market.Client, canonical, epic string, now time.Time) (mktwarmup.History, error) {
	m5c, err5 := c.GetPrices(ctx, epic, market.ResolutionMinute5, 80, time.Time{}, time.Time{})
	m15c, _ := c.GetPrices(ctx, epic, market.ResolutionMinute15, 80, time.Time{}, time.Time{})
	h1c, err1 := c.GetPrices(ctx, epic, market.ResolutionHour, 40, time.Time{}, time.Time{})
	h4c, err4 := c.GetPrices(ctx, epic, market.ResolutionHour4, 20, time.Time{}, time.Time{})
	if err5 != nil && err1 != nil && err4 != nil {
		return mktwarmup.ClosedUnavailable(canonical), err5
	}
	toLS := func(cs []models.Candle) []livesurface.Candle {
		var out []livesurface.Candle
		for _, x := range cs {
			out = append(out, livesurface.Candle{Time: x.Time, Close: x.Close, High: x.High, Low: x.Low})
		}
		return out
	}
	m5, h1, h4 := toLS(m5c), toLS(h1c), toLS(h4c)
	if len(m5)+len(h1)+len(h4) == 0 {
		return mktwarmup.ClosedUnavailable(canonical), nil
	}
	h := mktwarmup.FromBars(canonical, m5, h1, h4, now)
	h.M15Bars = toLS(m15c)
	h.M15 = len(h.M15Bars)
	return h, nil
}

func historyOf(market string) mktwarmup.History {
	histMu.Lock()
	defer histMu.Unlock()
	return histByMkt[market]
}

func cachedMaps() []instrument.Mapping {
	histMu.Lock()
	defer histMu.Unlock()
	return append([]instrument.Mapping{}, mapCache...)
}

func applyIntelMicro(st ops.Status) bool {
	if !intelOn {
		return st.L2QuotesOK && st.BookSynced
	}
	trades := st.Trades > 0 || st.EventRate > 0
	l2 := st.L2QuotesOK
	radar := st.LastPressure != 0 || st.RadarState != ""
	exh := st.LastV1Class != ""
	abs := st.AbsorptionStatus != "" || st.Absorption != 0
	return intelrt.BTCMicroHealthy(trades, l2, st.BookSynced, radar, exh, abs)
}

func bumpErr() {
	soak.mu.Lock()
	soak.Errors++
	soak.mu.Unlock()
}

func bumpWorld() {
	soak.mu.Lock()
	soak.WorldUpdates++
	soak.mu.Unlock()
}

func bumpQuotes(n, stale, open int) {
	soak.mu.Lock()
	soak.QuoteUpdates += n
	soak.StaleFrames += stale
	soak.MarketsOpen = open
	soak.mu.Unlock()
}

func bumpOpp() {
	soak.mu.Lock()
	soak.OppChanges++
	soak.mu.Unlock()
}

func bumpProposal() {
	soak.mu.Lock()
	soak.Proposals++
	soak.mu.Unlock()
}

func bumpLegacy() {
	soak.mu.Lock()
	soak.LegacyEvals++
	soak.mu.Unlock()
}

func pollGoldDemo(ctx context.Context) {
	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()
	client := &http.Client{Timeout: 2 * time.Second}
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:8765/api/gold", nil)
			if err != nil {
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			raw, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var m map[string]any
			if json.Unmarshal(raw, &m) != nil {
				continue
			}
			histMu.Lock()
			goldPeer = m
			histMu.Unlock()
		}
	}
}

func goldPeerCopy() map[string]any {
	histMu.Lock()
	defer histMu.Unlock()
	if goldPeer == nil {
		return nil
	}
	out := map[string]any{}
	for k, v := range goldPeer {
		out[k] = v
	}
	return out
}
