package researchopp

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"time"
)

type OutcomeRecord struct {
	ObservationID string             `json:"observation_id"`
	LabeledAt     time.Time          `json:"labeled_at"`
	Horizons      map[string]Outcome `json:"horizons"`
	AbsReturn     map[string]float64 `json:"abs_return,omitempty"`
	MFE           map[string]float64 `json:"mfe,omitempty"`
	MAE           map[string]float64 `json:"mae,omitempty"`
}

func OutcomesPath(dir string) string {
	if dir == "" {
		dir = "journals"
	}
	return filepath.Join(dir, "opportunity-outcomes.jsonl")
}

func AppendOutcomeOnly(dir string, rec OutcomeRecord) error {
	if rec.ObservationID == "" {
		return errEmptyID{}
	}
	if err := os.MkdirAll(filepath.Dir(OutcomesPath(dir)), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(OutcomesPath(dir), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	raw, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = f.Write(append(raw, '\n'))
	return err
}

type errEmptyID struct{}

func (errEmptyID) Error() string { return "observation_id required" }

func MarketFromID(id string) string {
	for i := 0; i < len(id); i++ {
		if id[i] == '-' {
			return id[:i]
		}
	}
	return id
}

func BuildOutcome(id string, t0 time.Time, entry float64, now time.Time, path []PricePoint) OutcomeRecord {
	rec := OutcomeRecord{ObservationID: id, LabeledAt: now.UTC(), Horizons: map[string]Outcome{}, AbsReturn: map[string]float64{}, MFE: map[string]float64{}, MAE: map[string]float64{}}
	for _, h := range []struct {
		name string
		d    time.Duration
	}{{"15m", 15 * time.Minute}, {"1h", time.Hour}, {"4h", 4 * time.Hour}, {"1d", 24 * time.Hour}} {
		if now.Before(t0.Add(h.d)) {
			continue
		}
		q := LabelQuality(t0, h.d, path, false)
		o := Outcome{Horizon: h.name, Quality: q}
		if r, ok := ForwardReturn(t0, h.d, entry, path); ok {
			o.Return = r
			rec.AbsReturn[h.name] = math.Abs(r)
		}
		rec.MFE[h.name], rec.MAE[h.name] = excursion(t0, h.d, entry, path)
		rec.Horizons[h.name] = o
	}
	return rec
}

func excursion(t0 time.Time, horizon time.Duration, entry float64, path []PricePoint) (mfe, mae float64) {
	if entry <= 0 {
		return 0, 0
	}
	end := t0.Add(horizon)
	for _, p := range path {
		if p.P <= 0 || p.T.Before(t0) || p.T.After(end) {
			continue
		}
		r := (p.P - entry) / entry
		if r > mfe {
			mfe = r
		}
		if r < mae {
			mae = r
		}
	}
	return mfe, mae
}
