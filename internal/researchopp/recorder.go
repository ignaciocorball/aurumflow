package researchopp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const Horizons = "15m,1h,4h,1d"

type Signal struct {
	ID           string          `json:"id"`
	T0           time.Time       `json:"t0"`
	WorldHash    string          `json:"world_hash"`
	Ranking      json.RawMessage `json:"ranking"`
	MarketState  json.RawMessage `json:"market_state"`
	Proposal     json.RawMessage `json:"proposal"`
	Outcomes     map[string]any  `json:"outcomes,omitempty"`
	Horizons     []string        `json:"horizons"`
}

func Path(dir string) string {
	if dir == "" {
		dir = "journals"
	}
	return filepath.Join(dir, "opportunity-research.jsonl")
}

func Record(dir string, s Signal) error {
	if s.T0.IsZero() {
		s.T0 = time.Now().UTC()
	}
	if len(s.Horizons) == 0 {
		s.Horizons = []string{"15m", "1h", "4h", "1d"}
	}
	if err := os.MkdirAll(filepath.Dir(Path(dir)), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(Path(dir), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = f.Write(append(raw, '\n'))
	return err
}
