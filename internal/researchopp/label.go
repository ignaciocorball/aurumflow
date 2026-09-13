package researchopp

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

const (
	OutcomeValid   = "VALID"
	OutcomePartial = "PARTIAL_SESSION"
	OutcomeUnavail = "UNAVAILABLE"
)

type PricePoint struct {
	T time.Time
	P float64
}

type Outcome struct {
	Horizon string  `json:"horizon"`
	Quality string  `json:"quality"`
	Return  float64 `json:"return,omitempty"`
}

func LabelQuality(t0 time.Time, horizon time.Duration, path []PricePoint, closedGap bool) string {
	if len(path) == 0 || closedGap {
		return OutcomeUnavail
	}
	end := t0.Add(horizon)
	var last time.Time
	for _, p := range path {
		if p.P <= 0 {
			continue
		}
		if !p.T.Before(t0) && !p.T.After(end) {
			last = p.T
		}
	}
	if last.IsZero() {
		return OutcomeUnavail
	}
	if end.Sub(last) > horizon/4 {
		return OutcomePartial
	}
	return OutcomeValid
}

func ForwardReturn(t0 time.Time, horizon time.Duration, entry float64, path []PricePoint) (float64, bool) {
	if entry <= 0 {
		return 0, false
	}
	end := t0.Add(horizon)
	var px float64
	ok := false
	for _, p := range path {
		if p.P <= 0 || p.T.Before(t0) {
			continue
		}
		if p.T.After(end) {
			break
		}
		px = p.P
		ok = true
	}
	if !ok {
		return 0, false
	}
	return (px - entry) / entry, true
}

func LabelRecord(path string, id string, now time.Time, prices map[string][]PricePoint) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	var out []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		var s Signal
		if json.Unmarshal([]byte(line), &s) != nil {
			out = append(out, line)
			continue
		}
		if s.ID != id {
			out = append(out, line)
			continue
		}
		if s.Outcomes == nil {
			s.Outcomes = map[string]any{}
		}
		px := prices[s.ID]
		if len(px) == 0 {
			// also try market prefix
			if i := strings.IndexByte(s.ID, '-'); i > 0 {
				px = prices[s.ID[:i]]
			}
		}
		for _, h := range []struct {
			name string
			d    time.Duration
		}{{"15m", 15 * time.Minute}, {"1h", time.Hour}, {"4h", 4 * time.Hour}, {"1d", 24 * time.Hour}} {
			q := LabelQuality(s.T0, h.d, px, false)
			rec := Outcome{Horizon: h.name, Quality: q}
			if q == OutcomeValid {
				if r, ok := ForwardReturn(s.T0, h.d, midFromState(s), px); ok {
					rec.Return = r
				}
			}
			s.Outcomes[h.name] = rec
		}
		b, _ := json.Marshal(s)
		out = append(out, string(b))
	}
	return os.WriteFile(path, []byte(strings.Join(out, "\n")+"\n"), 0o644)
}

func midFromState(s Signal) float64 {
	var st struct{ Mid float64 }
	_ = json.Unmarshal(s.MarketState, &st)
	return st.Mid
}
