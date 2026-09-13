package shadow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"aurumflow/internal/absorption"
	"aurumflow/internal/book"
	"aurumflow/internal/bookfeatures"
	"aurumflow/internal/collector"
	"aurumflow/internal/exhaustion"
	"aurumflow/internal/md"
	"aurumflow/internal/microflow"
	"aurumflow/internal/ops"
	"aurumflow/internal/radar"
	"aurumflow/internal/research"
)

const FeatureVersion = "P5.4-L2-1"

// Runtime is intelligence-plane only. It must not embed an ExecutionProvider.
type Runtime struct {
	Provider   md.MarketDataProvider
	Instrument string
	VenueInst  string
	Relation   string
	Book       *book.Book
	Micro      *microflow.Engine
	Exh        *exhaustion.Engine
	BookFeat   *bookfeatures.Engine
	Radar      *radar.Engine
	Trades     collector.SegmentedWriter
	Depth      collector.SegmentedWriter
	Snaps      collector.SegmentedWriter
	Health     collector.SegmentedWriter
	Status     *ops.Server
	Checkpoint string
	SeenIDs    map[string]bool
	Metrics    SoakMetrics
}

type SoakMetrics struct {
	Trades, Deltas, Snapshots int64
	Gaps, Resyncs             int
	Events                    int64
	Drops                     int64
	P50Ms, P95Ms              float64
	PeakMemMB                 float64
	Panics                    int
	Started                   time.Time
}

func HasExecutionProvider(v any) bool {
	if v == nil {
		return false
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Type.String() == "execution.ExecutionProvider" {
			return true
		}
	}
	return false
}

func SignalID(instrument string, ts time.Time, dir int) string {
	return instrument + "|" + ts.UTC().Format(time.RFC3339Nano) + "|" + itoa(dir) + "|legacy"
}

func itoa(n int) string {
	if n < 0 {
		return "-" + itoa(-n)
	}
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func (r *Runtime) LoadCheckpoint() {
	r.SeenIDs = map[string]bool{}
	if r.Checkpoint == "" {
		return
	}
	b, err := os.ReadFile(r.Checkpoint)
	if err != nil {
		return
	}
	var ck struct {
		IDs []string `json:"prospective_signal_ids"`
	}
	if json.Unmarshal(b, &ck) == nil {
		for _, id := range ck.IDs {
			r.SeenIDs[id] = true
		}
	}
}

func (r *Runtime) SaveCheckpoint() {
	if r.Checkpoint == "" {
		return
	}
	ids := make([]string, 0, len(r.SeenIDs))
	for id := range r.SeenIDs {
		ids = append(ids, id)
	}
	raw, _ := json.MarshalIndent(map[string]any{
		"saved_at": time.Now().UTC(),
		"provider": r.Provider.Name(),
		"prospective_signal_ids": ids,
		"book_synced": r.Book != nil && r.Book.Synced,
		"note": "stale book is not live after restart",
	}, "", "  ")
	_ = os.MkdirAll(filepath.Dir(r.Checkpoint), 0o755)
	_ = os.WriteFile(r.Checkpoint, raw, 0o644)
}

func (r *Runtime) RecordProspective(now time.Time, dir int, score int, pressure float64, v1 exhaustion.Snapshot, mf microflow.Snapshot, depth bookfeatures.Depth, liq bookfeatures.Liquidity, abs absorption.Snapshot) error {
	id := SignalID(r.Instrument, now.Truncate(time.Minute), dir)
	if r.SeenIDs[id] {
		return nil
	}
	plr := bookfeatures.Response(dir, depth, liq)
	in := research.ProspectiveInput{
		SignalID: id, RecordedAt: time.Now().UTC(), Timestamp: now.UTC(),
		Instrument: r.Instrument, Asset: "BITCOIN", LegacyDirection: dir, LegacyScore: score,
		PressureScore: pressure, DirectionalPressure: v1.DirectionalPressure,
		V1Classification: v1.Classification, Features: v1.Features,
		FeatureVersion: FeatureVersion, SpecHash: exhaustion.V1SpecHash, OutcomeKnown: false,
		Microflow: &mf, BookProvider: r.Provider.Name(), BookInstrument: r.VenueInst,
		BookSynced: r.Book != nil && r.Book.Synced, AbsorptionStatus: abs.Status,
		Passive: &plr, Relation: r.Relation,
		V1FlowProvider: "binance_usdm_public", L2Provider: r.Provider.Name(),
		L2FlowProvider: r.Provider.Name(), L2ProxyQuality: plr.L2ProxyQuality,
		Spread: depth.Spread, Mid: depth.Mid, Microprice: depth.Microprice,
		Imb1: depth.Imb1, Imb5: depth.Imb5, Imb10: depth.Imb10, Imb20: depth.Imb20,
	}
	if err := research.AppendInput(research.ProspectiveDir, in); err != nil {
		return err
	}
	_ = research.AppendInput(research.L2MechanismDir, in)
	if r.SeenIDs == nil {
		r.SeenIDs = map[string]bool{}
	}
	r.SeenIDs[id] = true
	r.SaveCheckpoint()
	return nil
}

func (r *Runtime) Close() {
	_ = r.Trades.Close()
	_ = r.Depth.Close()
	_ = r.Snaps.Close()
	_ = r.Health.Close()
	r.SaveCheckpoint()
}
