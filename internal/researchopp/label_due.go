package researchopp

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"
)

func LoadSignals(dir string) ([]Signal, error) {
	f, err := os.Open(Path(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []Signal
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var s Signal
		if json.Unmarshal([]byte(line), &s) != nil {
			continue
		}
		out = append(out, s)
	}
	return out, sc.Err()
}

func LoadLabeledHorizons(dir string) map[string]map[string]bool {
	out := map[string]map[string]bool{}
	f, err := os.Open(OutcomesPath(dir))
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var rec OutcomeRecord
		if json.Unmarshal([]byte(line), &rec) != nil {
			continue
		}
		if out[rec.ObservationID] == nil {
			out[rec.ObservationID] = map[string]bool{}
		}
		for h := range rec.Horizons {
			out[rec.ObservationID][h] = true
		}
	}
	return out
}

// LabelMature appends newly matured horizons to opportunity-outcomes.jsonl.
// It never rewrites t0 rows in opportunity-research.jsonl.
func LabelMature(dir string, now time.Time, prices map[string][]PricePoint, labeled map[string]map[string]bool) (mature15, mature1h int, err error) {
	if labeled == nil {
		labeled = LoadLabeledHorizons(dir)
	}
	sigs, err := LoadSignals(dir)
	if err != nil {
		return 0, 0, err
	}
	for _, s := range sigs {
		if s.ID == "" || s.WorldHash == "" || s.T0.IsZero() {
			continue
		}
		if s.IntegrityStatus == "" {
			continue
		}
		mkt := MarketFromID(s.ID)
		path := prices[mkt]
		if len(path) == 0 {
			path = prices[s.ID]
		}
		if len(path) == 0 {
			continue
		}
		entry := midFromState(s)
		rec := BuildOutcome(s.ID, s.T0, entry, now, path)
		if len(rec.Horizons) == 0 {
			continue
		}
		have := labeled[s.ID]
		fresh := OutcomeRecord{ObservationID: s.ID, LabeledAt: now.UTC(), Horizons: map[string]Outcome{}, AbsReturn: map[string]float64{}, MFE: map[string]float64{}, MAE: map[string]float64{}}
		added := false
		for h, o := range rec.Horizons {
			if have[h] {
				continue
			}
			fresh.Horizons[h] = o
			if v, ok := rec.AbsReturn[h]; ok {
				fresh.AbsReturn[h] = v
			}
			if v, ok := rec.MFE[h]; ok {
				fresh.MFE[h] = v
			}
			if v, ok := rec.MAE[h]; ok {
				fresh.MAE[h] = v
			}
			added = true
			if h == "15m" {
				mature15++
			}
			if h == "1h" {
				mature1h++
			}
		}
		if !added {
			continue
		}
		if err := AppendOutcomeOnly(dir, fresh); err != nil {
			return mature15, mature1h, err
		}
		if labeled[s.ID] == nil {
			labeled[s.ID] = map[string]bool{}
		}
		for h := range fresh.Horizons {
			labeled[s.ID][h] = true
		}
	}
	return mature15, mature1h, nil
}
