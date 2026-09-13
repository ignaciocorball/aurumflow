package globalsources

import (
	"time"

	"aurumflow/internal/worlddomain"
)

type CoverageRow struct {
	Source   string `json:"source"`
	Have     int    `json:"have"`
	Want     int    `json:"want"`
	Unit     string `json:"unit"`
	Label    string `json:"label"`
}

func OfficialCoverage(obs []worlddomain.ContextObservation, now time.Time) []CoverageRow {
	start := now.UTC().AddDate(-1, 0, 0)
	specs := []struct {
		src, metric, unit string
		want              int
		step              time.Duration
	}{
		{"FED_H41", "FED_ASSETS", "weeks", 52, 7 * 24 * time.Hour},
		{"ICI", "EQUITY", "weeks", 52, 7 * 24 * time.Hour},
		{"TIC", "NET_LT_SECURITIES", "months", 12, 0},
		{"JPX", "", "weeks", 52, 7 * 24 * time.Hour},
		{"CBOE", "TOTAL_PC", "days", 252, 24 * time.Hour},
		{"ECB", "POLICY_RATE", "weeks", 52, 7 * 24 * time.Hour},
		{"BIS", "GLOBAL_CREDIT_USD", "quarters", 4, 0},
		{"CFTC", "GOLD_MM_NET", "weeks", 52, 7 * 24 * time.Hour},
	}
	var out []CoverageRow
	for _, s := range specs {
		seen := map[string]bool{}
		for _, o := range obs {
			if o.Source != s.src || !o.Present {
				continue
			}
			if s.metric != "" && o.Metric != s.metric {
				continue
			}
			if o.ObservedAt.Before(start) {
				continue
			}
			key := o.ObservedAt.Format("2006-01-02")
			if s.unit == "months" {
				key = o.ObservedAt.Format("2006-01")
			}
			if s.unit == "quarters" {
				key = o.ObservedAt.Format("2006Q") + quarter(o.ObservedAt)
			}
			seen[key] = true
		}
		out = append(out, CoverageRow{
			Source: s.src, Have: len(seen), Want: s.want, Unit: s.unit,
			Label: itoa(len(seen)) + "/" + itoa(s.want) + " " + s.unit,
		})
	}
	return out
}

func quarter(t time.Time) string {
	return itoa(int((t.Month()-1)/3 + 1))
}
