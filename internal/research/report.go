package research

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ReportFile struct {
	GeneratedAt string    `json:"generated_at"`
	GitCommit   string    `json:"git_commit"`
	Result      BTCResult `json:"result"`
	H1          string    `json:"h1_radar_info"`
	H2          string    `json:"h2_pressure"`
	H3          string    `json:"h3_fusion"`
	Filter      string    `json:"filter"`
}

func WriteReports(dir string, res BTCResult, commit string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	h1, h2, h3, filt := Hypotheses(res)
	doc := ReportFile{GeneratedAt: time.Now().UTC().Format(time.RFC3339), GitCommit: commit, Result: res, H1: h1, H2: h2, H3: h3, Filter: filt}
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "FREE_RESEARCH_BASELINE.json"), raw, 0o644); err != nil {
		return err
	}
	md := renderMD(doc)
	return os.WriteFile(filepath.Join(dir, "FREE_RESEARCH_BASELINE.md"), []byte(md), 0o644)
}

func Hypotheses(res BTCResult) (h1, h2, h3, filter string) {
	filter = "INCONCLUSIVE"
	if res.FilterExcl.N >= 5 && res.FilterAll.N >= 5 {
		if res.FilterExcl.Hit > res.FilterAll.Hit+0.02 || res.FilterExcl.Mean > res.FilterAll.Mean {
			filter = "YES"
		} else if res.FilterExcl.Hit < res.FilterAll.Hit-0.02 && res.FilterExcl.Mean < res.FilterAll.Mean {
			filter = "NO"
		}
	}
	h1 = "INCONCLUSIVE"
	if res.LegacyAligned.N >= 5 && res.LegacyContradicted.N >= 5 {
		if res.LegacyAligned.Mean > res.LegacyContradicted.Mean && res.LegacyAligned.Hit > res.LegacyContradicted.Hit {
			h1 = "SUPPORTED"
		} else {
			h1 = "NOT SUPPORTED"
		}
	}
	h2 = "INCONCLUSIVE"
	if len(res.Pressure) >= 3 {
		var ns []int
		var means []float64
		for _, b := range res.Pressure {
			ns = append(ns, b.N)
			means = append(means, b.Mean)
		}
		if ns[0]+ns[len(ns)-1] >= 6 && means[len(means)-1] > means[0] {
			h2 = "SUPPORTED"
		} else if ns[0]+ns[len(ns)-1] >= 6 {
			h2 = "NOT SUPPORTED"
		}
	}
	h3 = "INCONCLUSIVE"
	if res.FusionOOS.N >= 5 && res.LegacyOOS.N >= 5 {
		if res.FusionOOS.Hit > res.LegacyOOS.Hit && res.FusionOOS.Mean > res.LegacyOOS.Mean {
			h3 = "SUPPORTED"
		} else {
			h3 = "NOT SUPPORTED"
		}
	}
	return
}

func renderMD(d ReportFile) string {
	r := d.Result
	var b strings.Builder
	fmt.Fprintf(&b, "# FREE RESEARCH BASELINE\n\n")
	fmt.Fprintf(&b, "Generated: %s\nCommit: %s\n\n", d.GeneratedAt, d.GitCommit)
	fmt.Fprintf(&b, "Asset: Binance USD-M BTCUSDT (NOT Capital BTCUSD)\n")
	fmt.Fprintf(&b, "Range: %s → %s (%d days)\n", r.From.UTC().Format(time.RFC3339), r.To.UTC().Format(time.RFC3339), r.Days)
	fmt.Fprintf(&b, "Trade events: %d\nRadar snapshots: %d\nRuntime: %s\nEvents/sec: %.0f\nMem MB: %.1f\nHorizon: %s\n\n", r.TradeEvents, r.RadarSnaps, r.Runtime, r.EventsPerSec, r.MemMB, r.Horizon)
	fmt.Fprintf(&b, "## A Legacy\nSignals: %d long=%d short=%d\n%s\n%s\n%s\n\n", r.LegacySignals, r.LegacyLong, r.LegacyShort, fmtStats("all", r.LegacyAll), fmtStats("long", r.LegacyLongStudy), fmtStats("short", r.LegacyShortStudy))
	fmt.Fprintf(&b, "## B Radar\nExpansion snaps: %d\n%s\n\n", r.RadarExpansion, fmtStats("radar_dir", r.RadarAll))
	fmt.Fprintf(&b, "## C Fusion\nAligned: %d\nContradicted: %d\n%s\n%s\n\n", r.Aligned, r.Contradicted, fmtStats("legacy_aligned", r.LegacyAligned), fmtStats("legacy_contradicted", r.LegacyContradicted))
	fmt.Fprintf(&b, "## Horizons (1m candles; %s)\n", r.Note10s30s)
	for _, hz := range r.Horizons {
		fmt.Fprintf(&b, "### %s\n%s\n%s\n%s\n%s\n", hz.Horizon, fmtStats("legacy", hz.Legacy), fmtStats("aligned", hz.Aligned), fmtStats("contradicted", hz.Contradicted), fmtStats("radar", hz.Radar))
	}
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "## Radar filter\nAll: %s\nExcl contradictions: %s\nImproves? %s\n\n", fmtStats("all", r.FilterAll), fmtStats("excl", r.FilterExcl), d.Filter)
	fmt.Fprintf(&b, "## Pressure\n")
	for _, x := range r.Pressure {
		fmt.Fprintf(&b, "- %s\n", fmtStats(x.Name, x))
	}
	fmt.Fprintf(&b, "\n## Confidence\n")
	for _, x := range r.Confidence {
		fmt.Fprintf(&b, "- %s\n", fmtStats(x.Name, x))
	}
	fmt.Fprintf(&b, "\n## States\n")
	for _, x := range r.States {
		fmt.Fprintf(&b, "- %s\n", fmtStats(x.Name, x))
	}
	fmt.Fprintf(&b, "\n## Walk-forward\nTrain %s → %s\nValid %s → %s\nOOS %s → %s\nLegacy OOS %s\nRadar OOS %s\nFusion aligned OOS %s\n\n",
		r.Split.TrainFrom.Format("2006-01-02"), r.Split.TrainTo.Format("2006-01-02"),
		r.Split.ValidFrom.Format("2006-01-02"), r.Split.ValidTo.Format("2006-01-02"),
		r.Split.OOSFrom.Format("2006-01-02"), r.Split.OOSTo.Format("2006-01-02"),
		fmtStats("legacy_oos", r.LegacyOOS), fmtStats("radar_oos", r.RadarOOS), fmtStats("fusion_oos", r.FusionOOS))
	fmt.Fprintf(&b, "## Hypotheses\nH1 Radar adds information: %s\nH2 Pressure magnitude orders outcomes: %s\nH3 Fusion improves Legacy OOS: %s\n", d.H1, d.H2, d.H3)
	fmt.Fprintf(&b, "\nCosts: ESTIMATED_COST not applied (event-study forward returns, not executable PnL).\n")
	return b.String()
}

func fmtStats(name string, s BucketStats) string {
	return fmt.Sprintf("%s n=%d hit=%.3f mean=%.5f ci=[%.5f,%.5f] median=%.5f mfe=%.5f mae=%.5f mfe/mae=%.2f", name, s.N, s.Hit, s.Mean, s.MeanLo, s.MeanHi, s.Median, s.MFE, s.MAE, s.Ratio)
}
