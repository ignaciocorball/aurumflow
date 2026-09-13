package research

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"aurumflow/internal/bookfeatures"
	"aurumflow/internal/exhaustion"
	"aurumflow/internal/microflow"
)

const ProspectiveDir = "research/prospective/FLOW_EXHAUSTION_V1"
const L2MechanismDir = "research/prospective/L2_MECHANISM"

type ProspectiveInput struct {
	SignalID            string    `json:"signal_id"`
	RecordedAt          time.Time `json:"recorded_at"`
	Timestamp           time.Time `json:"timestamp"`
	Instrument          string    `json:"instrument"`
	LegacyDirection     int       `json:"legacy_direction"`
	LegacyScore         int       `json:"legacy_score"`
	PressureScore       float64   `json:"pressure_score"`
	DirectionalPressure float64   `json:"directional_pressure"`
	V1Classification    string    `json:"v1_classification"`
	Features            exhaustion.FeatureSet `json:"features"`
	GitCommit           string    `json:"git_commit"`
	FeatureVersion      string    `json:"feature_version"`
	SpecHash            string    `json:"spec_hash"`
	OutcomeKnown        bool      `json:"outcome_known"`
	Microflow           *microflow.Snapshot `json:"microflow,omitempty"`
	BookProvider        string    `json:"book_provider,omitempty"`
	BookInstrument      string    `json:"book_instrument,omitempty"`
	BookSynced          bool      `json:"book_synced,omitempty"`
	BookAgeSec          float64   `json:"book_age_sec,omitempty"`
	AbsorptionStatus    string    `json:"absorption_status,omitempty"`
	Passive             *bookfeatures.PassiveLiquidityResponse `json:"passive_liquidity_response,omitempty"`
	Relation            string    `json:"venue_relation,omitempty"`
	EventWindowRef      string    `json:"event_window_ref,omitempty"`
	Asset               string    `json:"asset,omitempty"`
	V1FlowProvider      string    `json:"v1_flow_provider,omitempty"`
	L2Provider          string    `json:"l2_provider,omitempty"`
	L2FlowProvider      string    `json:"l2_flow_provider,omitempty"`
	L2ProxyQuality      string    `json:"l2_proxy_quality,omitempty"`
	Spread              float64   `json:"spread,omitempty"`
	Mid                 float64   `json:"mid,omitempty"`
	Microprice          float64   `json:"microprice,omitempty"`
	Imb1                float64   `json:"imbalance_top1,omitempty"`
	Imb5                float64   `json:"imbalance_top5,omitempty"`
	Imb10               float64   `json:"imbalance_top10,omitempty"`
	Imb20               float64   `json:"imbalance_top20,omitempty"`
	Session             string    `json:"session,omitempty"`
	Structure           string    `json:"structure,omitempty"`
	ReceiveLatencyMs    float64   `json:"receive_latency_ms,omitempty"`
}

type ProspectiveOutcome struct {
	SignalID         string             `json:"signal_id"`
	ObservedAt       time.Time          `json:"outcome_observed_at"`
	Horizons         map[string]Label   `json:"horizons"`
}

func InputPath(dir string) string   { return filepath.Join(dir, "signal_inputs.jsonl") }
func OutcomePath(dir string) string { return filepath.Join(dir, "signal_outcomes.jsonl") }

func AppendInput(dir string, rec ProspectiveInput) error {
	if rec.OutcomeKnown {
		return fmt.Errorf("input must have outcome_known=false")
	}
	if rec.SignalID == "" {
		return fmt.Errorf("signal_id required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if inputExists(InputPath(dir), rec.SignalID) {
		return fmt.Errorf("duplicate signal_id %s", rec.SignalID)
	}
	return appendJSONL(InputPath(dir), rec)
}

func AppendOutcome(dir string, rec ProspectiveOutcome) error {
	if rec.SignalID == "" {
		return fmt.Errorf("signal_id required")
	}
	if !inputExists(InputPath(dir), rec.SignalID) {
		return fmt.Errorf("unknown signal_id %s", rec.SignalID)
	}
	if outcomeExists(OutcomePath(dir), rec.SignalID) {
		return fmt.Errorf("outcome already recorded for %s", rec.SignalID)
	}
	return appendJSONL(OutcomePath(dir), rec)
}

func appendJSONL(path string, rec any) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(rec)
}

func inputExists(path, id string) bool {
	return scanExists(path, func(b []byte) bool {
		var r ProspectiveInput
		return json.Unmarshal(b, &r) == nil && r.SignalID == id
	})
}

func outcomeExists(path, id string) bool {
	return scanExists(path, func(b []byte) bool {
		var r ProspectiveOutcome
		return json.Unmarshal(b, &r) == nil && r.SignalID == id
	})
}

func scanExists(path string, match func([]byte) bool) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		if match(sc.Bytes()) {
			return true
		}
	}
	return false
}

type ProspectiveStatus struct {
	Start          string `json:"start"`
	Signals        int    `json:"signals_observed"`
	Exhaustion     int    `json:"exhaustion_signals"`
	Mature15m      int    `json:"mature_15m"`
	Mature1h       int    `json:"mature_1h"`
	Pending        int    `json:"pending"`
	Long           int    `json:"long"`
	Short          int    `json:"short"`
	Milestones     []string `json:"milestones"`
}

var ProspectiveHorizons = []time.Duration{
	time.Minute, 5 * time.Minute, 15 * time.Minute, 30 * time.Minute, time.Hour,
}

func ListInputs(dir string) ([]ProspectiveInput, error) {
	f, err := os.Open(InputPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []ProspectiveInput
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var r ProspectiveInput
		if json.Unmarshal(sc.Bytes(), &r) != nil {
			continue
		}
		out = append(out, r)
	}
	return out, sc.Err()
}

// LabelDue appends outcomes for signals whose horizons are complete. Never edits inputs.
func LabelDue(dir string, prices []PricePoint, now time.Time) (int, error) {
	ins, err := ListInputs(dir)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, in := range ins {
		if outcomeExists(OutcomePath(dir), in.SignalID) {
			continue
		}
		if now.Before(in.Timestamp.Add(time.Minute)) {
			continue
		}
		entry := priceAt(prices, in.Timestamp)
		if entry <= 0 {
			continue
		}
		labs := Labels(prices, in.Timestamp, entry, in.LegacyDirection, ProspectiveHorizons)
		hz := map[string]Label{}
		for _, lb := range labs {
			hz[lb.Horizon.String()] = lb
		}
		if err := AppendOutcome(dir, ProspectiveOutcome{SignalID: in.SignalID, ObservedAt: now.UTC(), Horizons: hz}); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func ReadProspectiveStatus(dir string) ProspectiveStatus {
	st := ProspectiveStatus{Start: "2026-09-13"}
	f, err := os.Open(InputPath(dir))
	if err != nil {
		return st
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	now := time.Now().UTC()
	for sc.Scan() {
		var r ProspectiveInput
		if json.Unmarshal(sc.Bytes(), &r) != nil {
			continue
		}
		st.Signals++
		if r.V1Classification == exhaustion.ClassExhaustion {
			st.Exhaustion++
		}
		if r.LegacyDirection > 0 {
			st.Long++
		} else if r.LegacyDirection < 0 {
			st.Short++
		}
		if now.Sub(r.Timestamp) >= 15*time.Minute {
			st.Mature15m++
		}
		if now.Sub(r.Timestamp) >= time.Hour {
			st.Mature1h++
		}
	}
	st.Pending = st.Signals - st.Mature15m
	for _, n := range []int{25, 50, 100, 200} {
		if st.Exhaustion >= n {
			st.Milestones = append(st.Milestones, fmt.Sprintf("EXHAUSTION_n=%d", n))
		}
	}
	return st
}
