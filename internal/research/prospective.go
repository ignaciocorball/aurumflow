package research

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ProspectiveRecord struct {
	RecordedAt       time.Time `json:"recorded_at"`
	SignalTime       time.Time `json:"signal_time"`
	SpecID           string    `json:"spec_id"`
	SpecHash         string    `json:"spec_hash"`
	LegacyDirection  int       `json:"legacy_direction"`
	OriginalPressure float64   `json:"original_pressure"`
	DirPressure      float64   `json:"directional_pressure"`
	FlowInterp       string    `json:"flow_interpretation"`
	LegacyScore      int       `json:"legacy_score,omitempty"`
	LegacyState      string    `json:"legacy_state,omitempty"`
	OutcomeKnown     bool      `json:"outcome_known"`
}

func AppendProspective(path string, rec ProspectiveRecord) error {
	if rec.OutcomeKnown {
		return fmt.Errorf("prospective input must be written before outcomes are known")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if existsProspective(path, rec.SignalTime, rec.SpecID) {
		return fmt.Errorf("immutable: record already exists for %s %s", rec.SpecID, rec.SignalTime.UTC().Format(time.RFC3339Nano))
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(rec)
}

func existsProspective(path string, at time.Time, spec string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var rec ProspectiveRecord
		if json.Unmarshal(sc.Bytes(), &rec) != nil {
			continue
		}
		if rec.SpecID == spec && rec.SignalTime.Equal(at) {
			return true
		}
	}
	return false
}
