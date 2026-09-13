package research

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"aurumflow/internal/quality"
)

type ExhaustionStudy struct {
	Role           string        `json:"role"`
	SpecID         string        `json:"spec_id"`
	SpecHash       string        `json:"spec_hash"`
	From           time.Time     `json:"from"`
	To             time.Time     `json:"to"`
	Days           int           `json:"days"`
	TradeEvents    int           `json:"trade_events"`
	Runtime        time.Duration `json:"runtime"`
	EventsPerSec   float64       `json:"events_per_sec"`
	MemMB          float64       `json:"mem_mb"`
	Quality        quality.Report `json:"quality"`
	Integrity      HoldoutIntegrity `json:"integrity"`
	Legacy         BucketStats   `json:"legacy_baseline"`
	Neutral        BucketStats   `json:"flow_neutral"`
	Continuation   BucketStats   `json:"flow_continuation"`
	Exhaustion     BucketStats   `json:"flow_exhaustion"`
	ExhLong        BucketStats   `json:"exhaustion_long"`
	ExhShort       BucketStats   `json:"exhaustion_short"`
	ContLong       BucketStats   `json:"continuation_long"`
	ContShort      BucketStats   `json:"continuation_short"`
	LegacyLong     BucketStats   `json:"legacy_long"`
	LegacyShort    BucketStats   `json:"legacy_short"`
	H5             HorizonBlock  `json:"h5"`
	H15            HorizonBlock  `json:"h15"`
	H30            HorizonBlock  `json:"h30"`
	H1h            HorizonBlock  `json:"h1h"`
	TimeToMFE      float64       `json:"exhaustion_time_to_mfe_sec"`
	TimeToMAE      float64       `json:"exhaustion_time_to_mae_sec"`
	Subperiods     []PeriodRow   `json:"subperiods"`
	Magnitude      []BucketStats `json:"pressure_magnitude"`
	Delay          []DelayRow    `json:"delay_diagnostic"`
	Decision       string        `json:"decision"`
	Symmetry       string        `json:"symmetry"`
	DeltaMean      float64       `json:"delta_mean"`
	DeltaHit       float64       `json:"delta_hit"`
	DeltaRatio     float64       `json:"delta_mfe_mae"`
	InterpNote     string        `json:"interpretation_note"`
}

type PeriodRow struct {
	Name string      `json:"period"`
	From time.Time   `json:"from"`
	To   time.Time   `json:"to"`
	Exh  BucketStats `json:"exhaustion"`
	Base BucketStats `json:"legacy"`
	Sign int         `json:"effect_sign"`
}

type DelayRow struct {
	Offset string      `json:"offset"`
	Label  string      `json:"label"`
	Exh    BucketStats `json:"exhaustion_using_that_pressure"`
}

type HoldoutIntegrity struct {
	Start       string   `json:"start"`
	End         string   `json:"end"`
	Days        int      `json:"days"`
	Files       int      `json:"archive_files"`
	Checksums   []string `json:"checksums,omitempty"`
	EventCount  int      `json:"event_count"`
	Duplicates  int      `json:"duplicates"`
	OutOfOrder  int      `json:"out_of_order"`
	Invalid     int      `json:"invalid_rows"`
	DatasetHash string   `json:"dataset_hash"`
}

func RunExhaustionFromZips(ctx context.Context, paths []string, symbol, role, specHash string, integrity HoldoutIntegrity) (ExhaustionStudy, error) {
	started := time.Now()
	m1, snaps, q, events, err := RunTape(ctx, paths, symbol)
	if err != nil {
		return ExhaustionStudy{}, err
	}
	fused, rp, prices := BuildFused(m1, snaps)
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	st := StudyExhaustion(fused, rp, prices, role, specHash, integrity, q, events, time.Since(started), float64(ms.Alloc)/1024/1024)
	return st, nil
}

func StudyExhaustion(fused []SignalRow, radar []RadarPoint, prices []PricePoint, role, specHash string, integrity HoldoutIntegrity, q quality.Report, events int, runtime time.Duration, mem float64) ExhaustionStudy {
	h := 15 * time.Minute
	st := ExhaustionStudy{
		Role: role, SpecID: SpecIDV1, SpecHash: specHash,
		TradeEvents: events, Runtime: runtime, MemMB: mem, Quality: q, Integrity: integrity,
	}
	if len(fused) > 0 {
		st.From, st.To = fused[0].Time, fused[len(fused)-1].Time
		if !st.From.IsZero() && !st.To.IsZero() {
			st.Days = int(st.To.Sub(st.From).Hours()/24) + 1
		}
	}
	if runtime > 0 {
		st.EventsPerSec = float64(events) / runtime.Seconds()
	}
	keep := func(name string) func(SignalRow) bool {
		return func(r SignalRow) bool { return r.FlowInterp == name }
	}
	st.Legacy = StudySignals(fused, prices, h, nil)
	st.Neutral = StudySignals(fused, prices, h, keep(FlowNeutral))
	st.Continuation = StudySignals(fused, prices, h, keep(FlowContinuation))
	st.Exhaustion = StudySignals(fused, prices, h, keep(FlowExhaustion))
	st.ExhLong = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.FlowInterp == FlowExhaustion && r.Direction > 0 })
	st.ExhShort = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.FlowInterp == FlowExhaustion && r.Direction < 0 })
	st.ContLong = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.FlowInterp == FlowContinuation && r.Direction > 0 })
	st.ContShort = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.FlowInterp == FlowContinuation && r.Direction < 0 })
	st.LegacyLong = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.Direction > 0 })
	st.LegacyShort = StudySignals(fused, prices, h, func(r SignalRow) bool { return r.Direction < 0 })
	st.H5 = horizonPair(fused, prices, 5*time.Minute)
	st.H15 = horizonPair(fused, prices, 15*time.Minute)
	st.H30 = horizonPair(fused, prices, 30*time.Minute)
	st.H1h = horizonPair(fused, prices, time.Hour)
	st.TimeToMFE, st.TimeToMAE = meanTimes(fused, prices, h, FlowExhaustion)
	bounds := SubperiodBounds(st.From, st.To.Add(time.Minute))
	names := []string{"EARLY", "MIDDLE", "LATE"}
	var signs []int
	for i, b := range bounds {
		ex := StudySignals(fused, prices, h, func(r SignalRow) bool {
			return r.FlowInterp == FlowExhaustion && InRange(r.Time, b[0], b[1])
		})
		lg := StudySignals(fused, prices, h, func(r SignalRow) bool { return InRange(r.Time, b[0], b[1]) })
		sg := SubperiodSign(ex.Mean, lg.Mean)
		if ex.N == 0 {
			sg = 0
		}
		signs = append(signs, sg)
		st.Subperiods = append(st.Subperiods, PeriodRow{Name: names[i], From: b[0], To: b[1], Exh: ex, Base: lg, Sign: sg})
	}
	mags := [][2]float64{{15, 20}, {20, 40}, {40, 60}, {60, 1e9}}
	magNames := []string{"15-20", "20-40", "40-60", "60+"}
	for i, b := range mags {
		st.Magnitude = append(st.Magnitude, namedStudy(fused, prices, h, magNames[i], func(r SignalRow) bool {
			if r.FlowInterp != FlowExhaustion {
				return false
			}
			a := abs(r.OriginalPressure)
			return a >= b[0] && a < b[1]
		}))
	}
	st.Delay = delayDiagnostic(fused, radar, prices, h)
	st.DeltaMean = st.Exhaustion.Mean - st.Legacy.Mean
	st.DeltaHit = st.Exhaustion.Hit - st.Legacy.Hit
	st.DeltaRatio = st.Exhaustion.Ratio - st.Legacy.Ratio
	st.Decision = DecideExhaustion(st.Exhaustion, st.Legacy, signs)
	if role != RoleHoldout {
		st.Decision = LabelPostHoc + "/" + st.Decision
	}
	st.Symmetry = SymmetryLabel(st.ExhLong.N, st.ExhShort.N, st.ExhLong.Mean-st.LegacyLong.Mean, st.ExhShort.Mean-st.LegacyShort.Mean)
	st.InterpNote = interpretExhaustion(fused)
	return st
}

func horizonPair(fused []SignalRow, prices []PricePoint, h time.Duration) HorizonBlock {
	return HorizonBlock{
		Horizon:      h.String(),
		Legacy:       StudySignals(fused, prices, h, nil),
		Aligned:      StudySignals(fused, prices, h, func(r SignalRow) bool { return r.FlowInterp == FlowContinuation }),
		Contradicted: StudySignals(fused, prices, h, func(r SignalRow) bool { return r.FlowInterp == FlowExhaustion }),
	}
}

func namedStudy(rows []SignalRow, prices []PricePoint, h time.Duration, name string, keep func(SignalRow) bool) BucketStats {
	s := StudySignals(rows, prices, h, keep)
	s.Name = name
	return s
}

func meanTimes(rows []SignalRow, prices []PricePoint, h time.Duration, interp string) (mfeSec, maeSec float64) {
	var n, smfe, smae float64
	for _, r := range rows {
		if r.FlowInterp != interp {
			continue
		}
		ls := Labels(prices, r.Time, priceAt(prices, r.Time), r.Direction, []time.Duration{h})
		if len(ls) == 0 {
			continue
		}
		smfe += ls[0].TimeToMFE.Seconds()
		smae += ls[0].TimeToMAE.Seconds()
		n++
	}
	if n == 0 {
		return 0, 0
	}
	return smfe / n, smae / n
}

func delayDiagnostic(fused []SignalRow, radar []RadarPoint, prices []PricePoint, h time.Duration) []DelayRow {
	offs := []struct {
		name string
		d    time.Duration
	}{{"t0", 0}, {"t0-1m", time.Minute}, {"t0-3m", 3 * time.Minute}, {"t0-5m", 5 * time.Minute}}
	var out []DelayRow
	for _, o := range offs {
		var tmp []SignalRow
		for _, s := range fused {
			rp := latestRadarAt(radar, s.Time.Add(-o.d))
			p := rp.Pressure
			if o.d == 0 {
				p = s.OriginalPressure
			}
			x := s
			x.OriginalPressure = p
			x.DirPressure = DirectionalPressure(s.Direction, p)
			x.FlowInterp = ClassifyFlow(s.Direction, p, DefaultAlignAbs)
			tmp = append(tmp, x)
		}
		out = append(out, DelayRow{
			Offset: o.name, Label: LabelPostHocDelay,
			Exh: StudySignals(tmp, prices, h, func(r SignalRow) bool { return r.FlowInterp == FlowExhaustion }),
		})
	}
	return out
}

func interpretExhaustion(rows []SignalRow) string {
	var n, lateSellLong, lateBuyShort int
	for _, r := range rows {
		if r.FlowInterp != FlowExhaustion {
			continue
		}
		n++
		if r.Direction > 0 && r.OriginalPressure < 0 {
			lateSellLong++
		}
		if r.Direction < 0 && r.OriginalPressure > 0 {
			lateBuyShort++
		}
	}
	if n == 0 {
		return "no exhaustion events"
	}
	return fmt.Sprintf("exhaustion_n=%d late_aggressive_sell_into_long=%d late_aggressive_buy_into_short=%d (no participant identity)", n, lateSellLong, lateBuyShort)
}

func WriteExhaustionReports(dir string, st ExhaustionStudy) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	name := "FLOW_EXHAUSTION_V1_HOLDOUT"
	if st.Role != RoleHoldout {
		name = "FLOW_EXHAUSTION_V1_DISCOVERY"
	}
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), raw, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".md"), []byte(renderExhaustion(st)), 0o644)
}

func renderExhaustion(st ExhaustionStudy) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\nRole: %s\nSpec: %s\nHash: %s\nRange: %s → %s (%d days)\nEvents: %d\nRuntime: %s\nEvents/sec: %.0f\nMem MB: %.1f\nDecision: %s\n\n",
		st.SpecID, st.Role, st.SpecID, st.SpecHash, st.From.UTC().Format(time.RFC3339), st.To.UTC().Format(time.RFC3339),
		st.Days, st.TradeEvents, st.Runtime, st.EventsPerSec, st.MemMB, st.Decision)
	fmt.Fprintf(&b, "## Legacy baseline\n%s\n\n", fmtStats("legacy", st.Legacy))
	fmt.Fprintf(&b, "## FLOW_NEUTRAL\n%s\n\n", fmtStats("neutral", st.Neutral))
	fmt.Fprintf(&b, "## FLOW_CONTINUATION_CONFIRM\n%s\n\n", fmtStats("continuation", st.Continuation))
	fmt.Fprintf(&b, "## FLOW_EXHAUSTION_CONFIRM\n%s\nDelta mean=%.5f hit=%.3f mfe/mae=%.2f\nTimeToMFE=%.0fs TimeToMAE=%.0fs\n%s\n\n",
		fmtStats("exhaustion", st.Exhaustion), st.DeltaMean, st.DeltaHit, st.DeltaRatio, st.TimeToMFE, st.TimeToMAE, st.InterpNote)
	fmt.Fprintf(&b, "## LONG/SHORT\n%s\n%s\n%s\n%s\nSymmetry: %s\n\n",
		fmtStats("exh_long", st.ExhLong), fmtStats("exh_short", st.ExhShort),
		fmtStats("leg_long", st.LegacyLong), fmtStats("leg_short", st.LegacyShort), st.Symmetry)
	fmt.Fprintf(&b, "## Subperiods\n")
	for _, p := range st.Subperiods {
		fmt.Fprintf(&b, "- %s %s sign=%d\n  %s\n  %s\n", p.Name, p.From.Format("2006-01-02"), p.Sign, fmtStats("exh", p.Exh), fmtStats("legacy", p.Base))
	}
	fmt.Fprintf(&b, "\n## Magnitude (exhaustion only)\n")
	for _, m := range st.Magnitude {
		fmt.Fprintf(&b, "- %s\n", fmtStats(m.Name, m))
	}
	fmt.Fprintf(&b, "\n## Delay diagnostic (%s)\n", LabelPostHocDelay)
	for _, d := range st.Delay {
		fmt.Fprintf(&b, "- %s %s\n", d.Offset, fmtStats("exh", d.Exh))
	}
	fmt.Fprintf(&b, "\n## Secondary horizons\n5m %s\n15m %s\n30m %s\n1h %s\n",
		fmtStats("exh5", st.H5.Contradicted), fmtStats("exh15", st.H15.Contradicted),
		fmtStats("exh30", st.H30.Contradicted), fmtStats("exh1h", st.H1h.Contradicted))
	return b.String()
}
