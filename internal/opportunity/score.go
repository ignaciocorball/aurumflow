package opportunity

import (
	"sort"

	"aurumflow/internal/worlddomain"
	"aurumflow/internal/worldstate"
)

const Spec = "ATTENTION_SCORE_V1"

var Weights = map[string]float64{
	"MacroAlignment":         15,
	"CapitalFlowAlignment":   15,
	"PositioningAsymmetry":   15,
	"CrossAssetConfirmation": 10,
	"Momentum":               15,
	"VolatilitySuitability":  10,
	"MicrostructureReadiness": 10,
	"DataQuality":            10,
}

type Ranked struct {
	Rank                 int
	Market               string
	Score                float64
	Coverage             float64
	CoverageConfidence   float64 // deprecated alias of EvidenceConfidence; not a probability
	EvidenceConfidence   float64 // Coverage/100: fraction of observed components
	Tier                 worlddomain.AttentionTier
	State                worldstate.MarketState
	Why                  []string
	Risks                []string
	Freshness            worlddomain.Frequency
	Components           map[string]float64
}

func Score(st worldstate.MarketState, ws worldstate.WorldState) Ranked {
	c := map[string]float64{}
	why := append([]string{}, st.Evidence...)
	c["MacroAlignment"] = knownBonus(st.MacroAlignment) * Weights["MacroAlignment"]
	c["CapitalFlowAlignment"] = flowBonus(st.CapitalFlowContext) * Weights["CapitalFlowAlignment"]
	c["PositioningAsymmetry"] = knownBonus(st.Positioning) * Weights["PositioningAsymmetry"]
	c["CrossAssetConfirmation"] = knownBonus(st.RelativeStrength) * Weights["CrossAssetConfirmation"]
	c["Momentum"] = momBonus(st.PriceTrend) * Weights["Momentum"]
	c["VolatilitySuitability"] = knownBonus(st.Volatility) * Weights["VolatilitySuitability"]
	micro := 0.0
	if st.MicroAvailable {
		micro = 1
		why = append(why, "microstructure available")
	}
	c["MicrostructureReadiness"] = micro * Weights["MicrostructureReadiness"]
	dq := 0.0
	if st.DataQuality == worlddomain.HealthHealthy {
		dq = 1
	} else if st.DataQuality == worlddomain.HealthStale {
		dq = 0.4
	}
	c["DataQuality"] = dq * Weights["DataQuality"]
	sum := 0.0
	for _, v := range c {
		sum += v
	}
	risks := append([]string{}, st.Risks...)
	if st.DataQuality == worlddomain.HealthUnknown {
		risks = append(risks, "incomplete evidence")
	}
	if st.Eligibility != worlddomain.EligDemo {
		risks = append(risks, "not DEMO_ELIGIBLE")
	}
	cov := CoverageScore(st)
	ev := cov / 100
	return Ranked{
		Market: st.Market, Score: sum, Coverage: cov,
		CoverageConfidence: ev, EvidenceConfidence: ev,
		Tier: tier(sum), State: st,
		Why: why, Risks: risks, Freshness: freshness(st), Components: c,
	}
}

func CoverageScore(st worldstate.MarketState) float64 {
	n, known := 0, 0
	check := func(s string) {
		n++
		if knownBonus(s) > 0 || flowBonus(s) > 0 || momBonus(s) > 0 {
			known++
		}
	}
	check(st.MacroAlignment)
	if st.CapitalFlowContext == "INFLOW" || st.CapitalFlowContext == "OUTFLOW" || knownBonus(st.CapitalFlowContext) > 0 {
		n++
		known++
	} else {
		n++
	}
	check(st.Positioning)
	check(st.RelativeStrength)
	if momBonus(st.PriceTrend) > 0 {
		n++
		known++
	} else {
		n++
	}
	check(st.Volatility)
	n++
	if st.MicroAvailable {
		known++
	}
	n++
	if st.DataQuality == worlddomain.HealthHealthy || st.DataQuality == worlddomain.HealthStale {
		known++
	}
	if n == 0 {
		return 0
	}
	return 100 * float64(known) / float64(n)
}

func Rank(ws worldstate.WorldState) []Ranked {
	var out []Ranked
	for _, st := range ws.Markets {
		empty := st.DataQuality == worlddomain.HealthUnknown && len(st.Evidence) == 0 && st.CapitalFlowContext == "UNKNOWN" && st.Positioning == "UNKNOWN" && st.MacroAlignment == "UNKNOWN" && st.PriceTrend == "UNKNOWN"
		if empty && !st.Resolved {
			continue
		}
		out = append(out, Score(st, ws))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Market < out[j].Market
		}
		return out[i].Score > out[j].Score
	})
	for i := range out {
		out[i].Rank = i + 1
	}
	return out
}

func tier(score float64) worlddomain.AttentionTier {
	switch {
	case score >= 70:
		return worlddomain.TierA
	case score >= 50:
		return worlddomain.TierB
	case score >= 30:
		return worlddomain.TierC
	default:
		return worlddomain.TierIgnore
	}
}

func knownBonus(s string) float64 {
	switch s {
	case "", "UNKNOWN", "unknown":
		return 0
	case "PENDING_PUBLIC_STRUCTURED_SOURCE", "PENDING_FREE_KEY":
		return 0
	default:
		return 1
	}
}

func flowBonus(s string) float64 {
	if s == "INFLOW" || s == "OUTFLOW" {
		return 1
	}
	return knownBonus(s)
}

func momBonus(s string) float64 {
	switch s {
	case "UP", "DOWN", "STRONG_UP", "STRONG_DOWN", "FLAT":
		return 1
	default:
		return 0
	}
}

func freshness(st worldstate.MarketState) worlddomain.Frequency {
	if st.MicroAvailable {
		return worlddomain.FreqLive
	}
	return worlddomain.FreqWeekly
}

func ConceptsSeparated(attention float64, setup worlddomain.SetupState, elig worlddomain.ExecEligibility) bool {
	if setup == worlddomain.SetupPotential && elig == worlddomain.EligLiveProhibited {
		return true
	}
	if attention >= 70 && elig != worlddomain.EligDemo {
		return true
	}
	return attention >= 0 && setup != "" && elig != ""
}
