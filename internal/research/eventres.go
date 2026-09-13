package research

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aurumflow/internal/binancehist"
	"aurumflow/internal/exhaustion"
	"aurumflow/internal/microflow"
)

type VelDist struct {
	Window string  `json:"window"`
	N      int     `json:"n"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	P25    float64 `json:"p25"`
	P75    float64 `json:"p75"`
}

type EventResBlock struct {
	Role         string             `json:"role"`
	Range        string             `json:"range"`
	Exhaustion   map[string]VelDist `json:"exhaustion"`
	Continuation map[string]VelDist `json:"continuation"`
	Neutral      map[string]VelDist `json:"neutral"`
	MatchTPS     map[string]float64 `json:"matched_trades_per_sec"`
	Central      string             `json:"short_window_velocity_conclusion"`
	Note         string             `json:"note"`
}

type EventResReport struct {
	Kind       string        `json:"kind"`
	Discovery  EventResBlock `json:"discovery"`
	Holdout    EventResBlock `json:"external_holdout"`
	Runtime    string        `json:"runtime"`
	Events     int           `json:"events"`
	EventsPerS float64       `json:"events_per_sec"`
	MemMB      float64       `json:"mem_mb"`
}

type eventRow struct {
	Class string
	Snap  microflow.Snapshot
	Dir   int
	Score int
	Hour  int
	Vol   float64
}

func AnalyzeEventResolution(role, rng string, fused []SignalRow, paths []string) (EventResBlock, int, error) {
	b := EventResBlock{
		Role: role, Range: rng,
		Note: "EXPLORATORY_MECHANISM_ONLY. True aggTrade timestamps. Not a performance validation.",
	}
	if len(fused) == 0 {
		return b, 0, nil
	}
	eng := microflow.New()
	si := 0
	events := 0
	var rows []eventRow
	for _, p := range paths {
		err := binancehist.IterTrades(p, func(tr binancehist.AggTrade) error {
			events++
			for si < len(fused) && !tr.Time.Before(fused[si].Time) {
				s := fused[si]
				snap := eng.Snapshot(s.Time)
				rows = append(rows, eventRow{Class: s.FlowInterp, Snap: snap, Dir: s.Direction, Score: s.Score, Hour: s.Time.UTC().Hour(), Vol: snap.Windows["5m"].QtyVelocity})
				si++
			}
			if si < len(fused) {
				eng.OnTrade(tr.Time, tr.Price, tr.Qty, tr.BuyerMaker)
			}
			return nil
		})
		if err != nil {
			return b, events, err
		}
	}
	for si < len(fused) {
		s := fused[si]
		rows = append(rows, eventRow{Class: s.FlowInterp, Snap: eng.Snapshot(s.Time), Dir: s.Direction, Score: s.Score, Hour: s.Time.UTC().Hour()})
		si++
	}
	var exh, cont, neu []eventRow
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
	b.Exhaustion = velGroupRows(exh)
	b.Continuation = velGroupRows(cont)
	b.Neutral = velGroupRows(neu)
	b.MatchTPS = matchVel(rows, vols)
	b.Central = velAnswer(b.Exhaustion, b.Continuation, b.Neutral, b.MatchTPS)
	return b, events, nil
}

func velGroupRows(rows []eventRow) map[string]VelDist {
	out := map[string]VelDist{}
	for _, w := range []string{"1s", "5s", "15s", "30s", "1m", "5m"} {
		var xs []float64
		for _, r := range rows {
			xs = append(xs, r.tps(w))
		}
		out[w] = summarizeVel(w, xs)
	}
	return out
}

func (r eventRow) tps(w string) float64 {
	return r.Snap.Windows[w].TradesPerSec
}

func summarizeVel(name string, xs []float64) VelDist {
	d := VelDist{Window: name, N: len(xs)}
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

func matchVel(rows []eventRow, vols []float64) map[string]float64 {
	mr := make([]exhaustion.MatchRow, len(rows))
	for i, r := range rows {
		mr[i] = exhaustion.MatchRow{Time: r.Snap.At, Dir: r.Dir, Hour: r.Hour, Score: r.Score, Vol: r.Vol, Class: r.Class, Idx: i}
	}
	pairs := exhaustion.MatchControls(mr, vols)
	out := map[string]float64{}
	if len(pairs) == 0 {
		return out
	}
	for _, w := range []string{"1s", "5s", "30s", "1m"} {
		s := 0.0
		for _, p := range pairs {
			s += rows[p[0]].tps(w) - rows[p[1]].tps(w)
		}
		out[w] = s / float64(len(pairs))
	}
	return out
}

func velAnswer(exh, cont, neu map[string]VelDist, match map[string]float64) string {
	if exh["1s"].N < 15 {
		return "INCONCLUSIVE"
	}
	higherNeu := exh["1s"].Mean > neu["1s"].Mean && exh["5s"].Mean > neu["5s"].Mean
	higherCont := exh["1s"].Mean > cont["1s"].Mean
	if match["1s"] > 0 && higherNeu {
		if higherCont {
			return "YES"
		}
		return "YES"
	}
	if !higherNeu && match["1s"] <= 0 {
		return "NO"
	}
	return "INCONCLUSIVE"
}

func WriteEventResReports(dir string, rep EventResReport) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	rep.Kind = "EXPLORATORY_MECHANISM_ONLY"
	raw, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "P5_3_EVENT_RESOLUTION_MECHANISM.json"), raw, 0o644); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# P5.3 Event-resolution mechanism\n\nEXPLORATORY_MECHANISM_ONLY. V1 membership unchanged.\n\n")
	fmt.Fprintf(&b, "Runtime %s events=%d ev/s=%.0f mem=%.1fMB\n\n", rep.Runtime, rep.Events, rep.EventsPerS, rep.MemMB)
	writeVel(&b, "DISCOVERY", rep.Discovery)
	writeVel(&b, "EXTERNAL_HOLDOUT", rep.Holdout)
	return os.WriteFile(filepath.Join(dir, "P5_3_EVENT_RESOLUTION_MECHANISM.md"), []byte(b.String()), 0o644)
}

func writeVel(b *strings.Builder, title string, blk EventResBlock) {
	fmt.Fprintf(b, "## %s (%s) conclusion=%s\n", title, blk.Range, blk.Central)
	for _, g := range []struct {
		n string
		m map[string]VelDist
	}{{"EXHAUSTION", blk.Exhaustion}, {"CONTINUATION", blk.Continuation}, {"NEUTRAL", blk.Neutral}} {
		fmt.Fprintf(b, "### %s\n", g.n)
		for _, w := range []string{"1s", "5s", "15s", "30s", "1m", "5m"} {
			d := g.m[w]
			fmt.Fprintf(b, "- %s n=%d mean=%.4f median=%.4f p25=%.4f p75=%.4f\n", w, d.N, d.Mean, d.Median, d.P25, d.P75)
		}
	}
	fmt.Fprintf(b, "Matched tps effects %v\n\n", blk.MatchTPS)
}