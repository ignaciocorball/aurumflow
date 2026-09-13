package researchopp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"aurumflow/internal/livesurface"
	"aurumflow/internal/opportunity"
	"aurumflow/internal/worldstate"
)

const Horizons = "15m,1h,4h,1d"

// Dedup: write at most every 5 minutes unless Opportunity tier or eligibility changes.
const DedupWindow = 5 * time.Minute

type Signal struct {
	ID                    string          `json:"id"`
	T0                    time.Time       `json:"t0"`
	WorldHash             string          `json:"world_hash"`
	Ranking               json.RawMessage `json:"ranking"`
	MarketState           json.RawMessage `json:"market_state"`
	Proposal              json.RawMessage `json:"proposal"`
	Outcomes              map[string]any  `json:"outcomes,omitempty"`
	Horizons              []string        `json:"horizons"`
	AttentionSpecVersion  string          `json:"attention_spec_version"`
	WorldStateVersion     string          `json:"worldstate_version"`
	MarketFeaturesVersion string          `json:"market_features_version"`
	GitCommit             string          `json:"git_commit"`
	Tier                  string          `json:"tier,omitempty"`
	Eligibility           string          `json:"eligibility,omitempty"`
}

func Path(dir string) string {
	if dir == "" {
		dir = "journals"
	}
	return filepath.Join(dir, "opportunity-research.jsonl")
}

func GitCommit() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" {
				if len(s.Value) > 12 {
					return s.Value[:12]
				}
				return s.Value
			}
		}
	}
	return "unknown"
}

func Stamp(s Signal) Signal {
	if s.T0.IsZero() {
		s.T0 = time.Now().UTC()
	}
	if len(s.Horizons) == 0 {
		s.Horizons = []string{"15m", "1h", "4h", "1d"}
	}
	if s.AttentionSpecVersion == "" {
		s.AttentionSpecVersion = opportunity.Spec
	}
	if s.WorldStateVersion == "" {
		s.WorldStateVersion = worldstate.Version
	}
	if s.MarketFeaturesVersion == "" {
		s.MarketFeaturesVersion = livesurface.Version
	}
	if s.GitCommit == "" {
		s.GitCommit = GitCommit()
	}
	return s
}

func ShouldRecord(prev *Signal, next Signal) bool {
	if next.WorldHash == "" {
		return false
	}
	if prev == nil {
		return true
	}
	if prev.WorldHash != next.WorldHash {
		return true
	}
	if prev.Tier != next.Tier || prev.Eligibility != next.Eligibility {
		return true
	}
	return next.T0.Sub(prev.T0) >= DedupWindow
}

func Record(dir string, s Signal) error {
	s = Stamp(s)
	if s.WorldHash == "" {
		return errEmptyHash
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

var errEmptyHash = errWorldHash{}

type errWorldHash struct{}

func (errWorldHash) Error() string { return "world_hash required" }
