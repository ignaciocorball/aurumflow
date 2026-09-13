package research

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

const (
	FlowExhaustion    = "FLOW_EXHAUSTION_CONFIRM"
	FlowContinuation  = "FLOW_CONTINUATION_CONFIRM"
	FlowNeutral       = "FLOW_NEUTRAL"
	SpecIDV1          = "FLOW_EXHAUSTION_V1"
	DefaultAlignAbs   = 15.0
	MinExhaustionN    = 50
	RoleHoldout       = "EXTERNAL_HOLDOUT_1"
	RoleDiscovery     = "POST_HOC_DISCOVERY"
	LabelPostHoc      = "POST_HOC_DIAGNOSTIC"
	LabelPostHocDelay = "POST_HOC_DELAY_DIAGNOSTIC"
)

func ClassifyFlow(legacyDir int, pressure, thresh float64) string {
	if thresh <= 0 {
		thresh = DefaultAlignAbs
	}
	if legacyDir == 0 || math.Abs(pressure) < thresh {
		return FlowNeutral
	}
	dp := float64(legacyDir) * pressure
	if dp <= -thresh {
		return FlowExhaustion
	}
	if dp >= thresh {
		return FlowContinuation
	}
	return FlowNeutral
}

func DirectionalPressure(legacyDir int, pressure float64) float64 {
	return float64(legacyDir) * pressure
}

func AnnotateFlow(rows []SignalRow, thresh float64) []SignalRow {
	out := make([]SignalRow, len(rows))
	for i, s := range rows {
		s.OriginalPressure = s.Pressure
		s.DirPressure = DirectionalPressure(s.Direction, s.Pressure)
		s.FlowInterp = ClassifyFlow(s.Direction, s.Pressure, thresh)
		out[i] = s
	}
	return out
}

func OverlapsDiscovery(from, to, discFrom, discTo time.Time) bool {
	return from.Before(discTo) && to.After(discFrom)
}

func HoldoutValid(from, to, discStart time.Time) bool {
	return !to.After(discStart) && from.Before(to)
}

func DecideExhaustion(exh, base BucketStats, subSigns []int) string {
	if exh.N < MinExhaustionN {
		return "INCONCLUSIVE"
	}
	effect := exh.Mean - base.Mean
	if effect <= 0 || catastrophicMAE(exh.MAE, base.MAE) {
		return "NOT_SUPPORTED"
	}
	pos := 0
	for _, s := range subSigns {
		if s > 0 {
			pos++
		}
	}
	if len(subSigns) >= 3 && pos < 2 {
		return "INCONCLUSIVE"
	}
	return "SUPPORTED"
}

func catastrophicMAE(exhMAE, baseMAE float64) bool {
	if baseMAE >= 0 {
		return exhMAE < -2*math.Abs(baseMAE+1e-12)
	}
	return math.Abs(exhMAE) > 2*math.Abs(baseMAE)
}

func SymmetryLabel(longN, shortN int, longEffect, shortEffect float64) string {
	lOK, sOK := longN >= 20, shortN >= 20
	if !lOK && !sOK {
		return "INSUFFICIENT_SAMPLE"
	}
	lPos, sPos := longEffect > 0, shortEffect > 0
	if lOK && sOK && lPos && sPos {
		return "SYMMETRIC"
	}
	if lOK && lPos && (!sOK || !sPos) {
		return "LONG_ONLY"
	}
	if sOK && sPos && (!lOK || !lPos) {
		return "SHORT_ONLY"
	}
	return "NEITHER"
}

type SpecFile struct {
	SpecID        string          `json:"spec_id"`
	CreatedAt     string          `json:"created_at"`
	GitCommit     string          `json:"git_commit"`
	FeatureVer    string          `json:"feature_version"`
	ThresholdAbs  float64         `json:"-"`
	Raw           json.RawMessage `json:"-"`
	SpecHash      string          `json:"spec_hash"`
	PressureBlock json.RawMessage `json:"pressure"`
}

func CanonicalSpecHash(raw []byte) (string, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return "", err
	}
	delete(m, "spec_hash")
	can, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(can)
	return hex.EncodeToString(sum[:]), nil
}

func LoadSpec(path string) (hash string, raw []byte, err error) {
	raw, err = os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	hash, err = CanonicalSpecHash(raw)
	return hash, raw, err
}

func DatasetHash(checksums []string) string {
	h := sha256.New()
	for _, c := range checksums {
		_, _ = fmt.Fprintln(h, c)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func SubperiodBounds(from, to time.Time) [3][2]time.Time {
	from, to = from.UTC(), to.UTC()
	span := to.Sub(from)
	a := from.Add(span / 3)
	b := from.Add(2 * span / 3)
	return [3][2]time.Time{{from, a}, {a, b}, {b, to}}
}

func SubperiodSign(exhMean, baseMean float64) int {
	d := exhMean - baseMean
	if d > 0 {
		return 1
	}
	if d < 0 {
		return -1
	}
	return 0
}

func latestRadarAt(radar []RadarPoint, at time.Time) RadarPoint {
	i := sort.Search(len(radar), func(i int) bool { return radar[i].Time.After(at) })
	if i == 0 {
		return RadarPoint{}
	}
	return radar[i-1]
}
