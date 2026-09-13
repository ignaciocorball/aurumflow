package research

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aurumflow/internal/exhaustion"
	"aurumflow/internal/radar"
	"aurumflow/pkg/models"
)

type MechRow struct {
	Time   time.Time
	Class  string
	Dir    int
	Score  int
	Hour   int
	Vol    float64
	Snap   exhaustion.Snapshot
}

type Dist struct {
	Name   string  `json:"name"`
	N      int     `json:"n"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	P25    float64 `json:"p25"`
	P75    float64 `json:"p75"`
}

type MechBlock struct {
	Role         string            `json:"role"`
	Range        string            `json:"range"`
	LegacyN      int               `json:"legacy_n"`
	ExhaustionN  int               `json:"exhaustion_n"`
	ContinuationN int              `json:"continuation_n"`
	NeutralN     int               `json:"neutral_n"`
	Exhaustion   map[string]Dist   `json:"exhaustion"`
	Continuation map[string]Dist   `json:"continuation"`
	Neutral      map[string]Dist   `json:"neutral"`
	Match        MatchReport       `json:"matched_controls"`
	Central      string            `json:"central_question"`
	Paths        [][]exhaustion.PathPoint `json:"paths,omitempty"`
	PrePost      PrePost           `json:"pre_post"`
}

type MatchReport struct {
	Method string  `json:"method"`
	ExhN   int     `json:"exhaustion_n"`
	CtlN   int     `json:"control_n"`
	Effects map[string]float64 `json:"effects"`
}

type PrePost struct {
	Note string `json:"note"`
	PreAggMean float64 `json:"pre_aggression_mean"`
	PostDispLegacy float64 `json:"post_disp_vs_legacy_mean"`
}

type MechanismReport struct {
	Discovery MechBlock `json:"discovery"`
	Holdout   MechBlock `json:"external_holdout"`
	Consistency map[string]string `json:"cross_dataset_consistency"`
	Overall   string `json:"overall_central"`
	Runtime   string `json:"runtime"`
	Events    int    `json:"events"`
	EventsPerSec float64 `json:"events_per_sec"`
	MemMB     float64 `json:"mem_mb"`
}

func FeaturesFromTape(fused []SignalRow, m1 []models.Candle, snaps []radar.PressureSnapshot) []MechRow {
	byMin := map[int64]radar.PressureSnapshot{}
	for _, s := range snaps {
		byMin[s.Timestamp.Unix()] = s
	}
	eng := exhaustion.NewEngine("BTCUSDT")
	out := make([]MechRow, 0, len(fused))
	fi := 0
	for i, c := range m1 {
		if s, ok := byMin[c.Time.Unix()]; ok {
			if s.AggBuy > 0 {
				eng.OnTrade(c.Time.Add(-2*time.Second), c.Close, s.AggBuy, false)
			}
			if s.AggSell > 0 {
				eng.OnTrade(c.Time.Add(-time.Second), c.Close, s.AggSell, true)
			}
		} else if c.Volume > 0 {
			eng.OnTrade(c.Time.Add(-time.Second), c.Close, c.Volume, false)
		}
		for fi < len(fused) && !fused[fi].Time.After(c.Time) {
			s := fused[fi]
			atr := 0.0
			if i > 0 {
				atr = math.Abs(c.Close - m1[i-1].Close)
			}
			// Observe at signal time using only trades already ingested (this minute and earlier).
			snap := eng.Observe(s.Time, s.Direction, float64(s.Score), s.OriginalPressure, atr)
			out = append(out, MechRow{
				Time: s.Time, Class: s.FlowInterp, Dir: s.Direction, Score: s.Score,
				Hour: s.Time.UTC().Hour(), Vol: snap.Features.RealizedVol, Snap: snap,
			})
			fi++
		}
	}
	return out
}

func AnalyzeMechanism(role, rng string, fused []SignalRow, m1 []models.Candle, snaps []radar.PressureSnapshot) MechBlock {
	rows := FeaturesFromTape(fused, m1, snaps)
	b := MechBlock{Role: role, Range: rng, LegacyN: len(fused)}
	var exh, cont, neu []MechRow
	var vols []float64
	for _, r := range rows {
		vols = append(vols, r.Vol)
		switch r.Class {
		case exhaustion.ClassExhaustion:
			exh = append(exh, r)
		case exhaustion.ClassContinuation:
			cont = append(cont, r)
		default:
			neu = append(neu, r)
		}
	}
	b.ExhaustionN, b.ContinuationN, b.NeutralN = len(exh), len(cont), len(neu)
	b.Exhaustion = groupDists(exh)
	b.Continuation = groupDists(cont)
	b.Neutral = groupDists(neu)
	b.Match = matchStudy(rows, vols)
	b.Central = centralAnswer(b.Exhaustion, b.Neutral, b.Match)
	mins := minutesFrom(m1, snaps)
	var preAgg, postDisp float64
	npp := 0
	for _, r := range exh {
		p := exhaustion.BuildPath(mins, r.Time)
		if len(p) > 0 {
			b.Paths = append(b.Paths, p)
		}
		pre, post := pathPrePost(p)
		preAgg += pre
		postDisp += post
		npp++
	}
	if npp > 0 {
		b.PrePost = PrePost{
			Note: "PRE = mean |imbalance| t-15..t0; POST = mean price_norm * legacy at +15 (descriptive)",
			PreAggMean: preAgg / float64(npp), PostDispLegacy: postDisp / float64(npp),
		}
	}
	if len(b.Paths) > 40 {
		b.Paths = b.Paths[:40]
	}
	return b
}

func groupDists(rows []MechRow) map[string]Dist {
	pick := func(name string, f func(MechRow) float64) Dist {
		var xs []float64
		for _, r := range rows {
			xs = append(xs, f(r))
		}
		return summarizeDist(name, xs)
	}
	return map[string]Dist{
		"flow_magnitude": pick("flow_magnitude", func(r MechRow) float64 { return abs(r.Snap.Features.FlowMagNorm) }),
		"flow_imbalance": pick("flow_imbalance", func(r MechRow) float64 {
			w := r.Snap.Features.Windows["5m0s"]
			return w.ImbalanceRatio
		}),
		"cvd": pick("cvd", func(r MechRow) float64 { return r.Snap.Features.CVD }),
		"cvd_slope": pick("cvd_slope", func(r MechRow) float64 {
			return r.Snap.Features.Windows["5m0s"].CVDSlope
		}),
		"trade_velocity": pick("trade_velocity", func(r MechRow) float64 {
			return r.Snap.Features.Windows["5m0s"].TradeVelocity
		}),
		"price_displacement": pick("price_displacement", func(r MechRow) float64 {
			return r.Snap.Features.Prices["5m0s"].DispVsFlow
		}),
		"flow_efficiency": pick("flow_efficiency", func(r MechRow) float64 { return r.Snap.Features.FlowEffNorm }),
		"impact_failure": pick("impact_failure", func(r MechRow) float64 { return r.Snap.Features.ImpactFailure }),
	}
}

func summarizeDist(name string, xs []float64) Dist {
	d := Dist{Name: name, N: len(xs)}
	if len(xs) == 0 {
		return d
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	d.Mean = sum / float64(len(xs))
	cp := append([]float64(nil), xs...)
	for i := 1; i < len(cp); i++ {
		j := i
		for j > 0 && cp[j] < cp[j-1] {
			cp[j], cp[j-1] = cp[j-1], cp[j]
			j--
		}
	}
	d.Median = cp[len(cp)/2]
	d.P25 = cp[len(cp)/4]
	d.P75 = cp[(3*len(cp))/4]
	return d
}

func matchStudy(rows []MechRow, vols []float64) MatchReport {
	mr := make([]exhaustion.MatchRow, len(rows))
	for i, r := range rows {
		mr[i] = exhaustion.MatchRow{Time: r.Time, Dir: r.Dir, Hour: r.Hour, Score: r.Score, Vol: r.Vol, Class: r.Class, Idx: i}
	}
	pairs := exhaustion.MatchControls(mr, vols)
	rep := MatchReport{Method: "stratified nearest unused control (dir, UTC hour, vol tertile, score bucket); no future outcomes", ExhN: len(pairs), CtlN: len(pairs), Effects: map[string]float64{}}
	if len(pairs) == 0 {
		return rep
	}
	keys := []string{"flow_magnitude", "flow_efficiency", "impact_failure", "cvd", "trade_velocity"}
	get := func(r MechRow, k string) float64 {
		switch k {
		case "flow_magnitude":
			return abs(r.Snap.Features.FlowMagNorm)
		case "flow_efficiency":
			return r.Snap.Features.FlowEffNorm
		case "impact_failure":
			return r.Snap.Features.ImpactFailure
		case "cvd":
			return r.Snap.Features.CVD
		default:
			return r.Snap.Features.Windows["5m0s"].TradeVelocity
		}
	}
	for _, k := range keys {
		s := 0.0
		for _, p := range pairs {
			s += get(rows[p[0]], k) - get(rows[p[1]], k)
		}
		rep.Effects[k] = s / float64(len(pairs))
	}
	return rep
}

func centralAnswer(exh, neu map[string]Dist, m MatchReport) string {
	if exh["flow_magnitude"].N < 20 {
		return "INCONCLUSIVE"
	}
	stronger := exh["flow_magnitude"].Mean > neu["flow_magnitude"].Mean
	weaker := exh["price_displacement"].Mean < neu["price_displacement"].Mean || exh["impact_failure"].Mean > neu["impact_failure"].Mean
	if m.CtlN >= 15 {
		if m.Effects["flow_magnitude"] > 0 && m.Effects["impact_failure"] > 0 {
			stronger, weaker = true, true
		}
	}
	if stronger && weaker {
		return "YES"
	}
	if !stronger && !weaker {
		return "NO"
	}
	return "INCONCLUSIVE"
}

func minutesFrom(m1 []models.Candle, snaps []radar.PressureSnapshot) []exhaustion.MinuteObs {
	sm := map[int64]radar.PressureSnapshot{}
	for _, s := range snaps {
		sm[s.Timestamp.Unix()] = s
	}
	out := make([]exhaustion.MinuteObs, 0, len(m1))
	for _, c := range m1 {
		o := exhaustion.MinuteObs{T: c.Time, Close: c.Close}
		if s, ok := sm[c.Time.Unix()]; ok {
			o.CVD = s.CVD
			tot := s.AggBuy + s.AggSell
			if tot > 0 {
				o.Imbalance = (s.AggBuy - s.AggSell) / tot
			}
			o.Velocity = s.TradeVel
		}
		out = append(out, o)
	}
	return out
}

func pathPrePost(p []exhaustion.PathPoint) (preAbsImb, postPrice float64) {
	var npre, npost int
	for _, x := range p {
		if x.OffsetMin < 0 {
			preAbsImb += abs(x.Imbalance)
			npre++
		}
		if x.OffsetMin == 15 {
			postPrice = x.Price
			npost++
		}
	}
	if npre > 0 {
		preAbsImb /= float64(npre)
	}
	_ = npost
	return
}

func Consistency(a, b Dist) string {
	if a.N < 10 || b.N < 10 {
		return "MIXED"
	}
	sa := sign(a.Mean)
	sb := sign(b.Mean)
	if sa == sb {
		return "CONSISTENT"
	}
	return "INCONSISTENT"
}

func sign(v float64) int {
	if v > 0 {
		return 1
	}
	if v < 0 {
		return -1
	}
	return 0
}

func WriteMechanismReports(dir string, rep MechanismReport) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "FLOW_EXHAUSTION_MECHANISM.json"), raw, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "FLOW_EXHAUSTION_MECHANISM.md"), []byte(renderMech(rep)), 0o644)
}

func renderMech(r MechanismReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# FLOW EXHAUSTION MECHANISM\n\nEXPLORATORY_MECHANISM_RESEARCH only. Not a new validation number.\n\n")
	fmt.Fprintf(&b, "Runtime %s events=%d ev/s=%.0f mem=%.1fMB\n\n", r.Runtime, r.Events, r.EventsPerSec, r.MemMB)
	writeBlock(&b, "DISCOVERY", r.Discovery)
	writeBlock(&b, "EXTERNAL_HOLDOUT", r.Holdout)
	fmt.Fprintf(&b, "## Consistency\n")
	for k, v := range r.Consistency {
		fmt.Fprintf(&b, "- %s: %s\n", k, v)
	}
	fmt.Fprintf(&b, "\nOVERALL central question: %s\n", r.Overall)
	return b.String()
}

func writeBlock(b *strings.Builder, title string, blk MechBlock) {
	fmt.Fprintf(b, "## %s (%s)\nLegacy=%d Exh=%d Cont=%d Neu=%d Central=%s\n", title, blk.Range, blk.LegacyN, blk.ExhaustionN, blk.ContinuationN, blk.NeutralN, blk.Central)
	for _, g := range []struct {
		n string
		m map[string]Dist
	}{{"EXHAUSTION", blk.Exhaustion}, {"CONTINUATION", blk.Continuation}, {"NEUTRAL", blk.Neutral}} {
		fmt.Fprintf(b, "### %s\n", g.n)
		for _, k := range []string{"flow_magnitude", "flow_imbalance", "cvd", "cvd_slope", "trade_velocity", "price_displacement", "flow_efficiency", "impact_failure"} {
			d := g.m[k]
			fmt.Fprintf(b, "- %s n=%d mean=%.5f median=%.5f p25=%.5f p75=%.5f\n", k, d.N, d.Mean, d.Median, d.P25, d.P75)
		}
	}
	fmt.Fprintf(b, "Matched %s exh=%d ctl=%d effects=%v\nPrePost %+v\n\n", blk.Match.Method, blk.Match.ExhN, blk.Match.CtlN, blk.Match.Effects, blk.PrePost)
}
