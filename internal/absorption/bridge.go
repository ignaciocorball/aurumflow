package absorption

import (
	"aurumflow/internal/bookfeatures"
	"aurumflow/internal/exhaustion"
)

const (
	Unavailable        = "UNAVAILABLE"
	InsufficientData   = "INSUFFICIENT_DATA"
	NotSupportive      = "NOT_SUPPORTIVE"
	Mixed              = "MIXED"
	Supportive         = "SUPPORTIVE"
	StronglySupportive = "STRONGLY_SUPPORTIVE"
	ModeShadow         = "SHADOW"
)

type Component struct {
	Name    string  `json:"name"`
	Value   float64 `json:"value"`
	Present bool    `json:"present"`
}

type EvidenceRow struct {
	Name  string `json:"name"`
	Level string `json:"level"`
	Value float64 `json:"value"`
}

type Snapshot struct {
	Status              string                             `json:"status"`
	Mode                string                             `json:"mode"`
	V1Class             string                             `json:"v1_classification"`
	PressureScore       float64                            `json:"pressure_score"`
	DirectionalPressure float64                            `json:"directional_pressure"`
	FlowEfficiency      float64                            `json:"flow_efficiency"`
	ImpactFailure       float64                            `json:"impact_failure"`
	BookSynced          bool                               `json:"book_synced"`
	Passive             bookfeatures.PassiveLiquidityResponse `json:"passive_liquidity_response"`
	Components          []Component                        `json:"components"`
	Evidence            []EvidenceRow                      `json:"evidence,omitempty"`
	Why                 string                             `json:"why,omitempty"`
	EvidenceStatus      string                             `json:"evidence_status"`
}

func MayMutateBroker() bool { return false }

// Observe is SHADOW-only and uses t0 evidence already computed (no future book).
func Observe(v1 exhaustion.Snapshot, bookOK bool, bookAgeSec float64, plr bookfeatures.PassiveLiquidityResponse) Snapshot {
	s := Snapshot{
		Mode: ModeShadow, V1Class: v1.Classification,
		PressureScore: v1.PressureScore, DirectionalPressure: v1.DirectionalPressure,
		FlowEfficiency: v1.Features.FlowEffNorm, ImpactFailure: v1.Features.ImpactFailure,
		BookSynced: bookOK, Passive: plr, EvidenceStatus: "DESCRIPTIVE_RESEARCH_ONLY",
	}
	if !bookOK {
		s.Status = Unavailable
		return s
	}
	if bookAgeSec > 5 {
		s.Status = InsufficientData
		return s
	}
	comps := []Component{
		{Name: "aggression_against_legacy", Value: v1.DirectionalPressure, Present: v1.Classification == exhaustion.ClassExhaustion},
		{Name: "high_impact_failure", Value: v1.Features.ImpactFailure, Present: v1.Features.ImpactFailure > 0},
		{Name: "supporting_replenishment", Value: plr.SupportingReplenishment, Present: plr.SupportingReplenishment > 0},
		{Name: "supporting_persistence", Value: plr.SupportingPersistence, Present: plr.SupportingPersistence > 0},
		{Name: "opposing_depletion", Value: plr.OpposingDepletion, Present: plr.OpposingDepletion > 0},
		{Name: "book_imbalance_response", Value: plr.BookImbalanceResponse, Present: plr.BookImbalanceResponse > 0},
		{Name: "microprice_refuses_aggression", Value: plr.MicropriceResponse, Present: plr.MicropriceResponse > 0},
	}
	s.Components = comps
	n := 0
	for _, c := range comps {
		if c.Present {
			n++
		}
	}
	switch {
	case n == 0:
		s.Status = NotSupportive
	case n <= 2:
		s.Status = Mixed
	case n <= 4:
		s.Status = Supportive
	default:
		s.Status = StronglySupportive
	}
	s.Evidence = []EvidenceRow{
		{Name: "AggressionAgainstLegacy", Level: yn(v1.Classification == exhaustion.ClassExhaustion), Value: v1.DirectionalPressure},
		{Name: "ImpactFailure", Level: band(v1.Features.ImpactFailure, 0.5, 1.5), Value: v1.Features.ImpactFailure},
		{Name: "SupportingRefill", Level: band(plr.SupportingReplenishment, 0.1, 1), Value: plr.SupportingReplenishment},
		{Name: "SupportingPersistence", Level: band(plr.SupportingPersistence, 0.5, 2), Value: plr.SupportingPersistence},
		{Name: "OpposingDepletion", Level: band(plr.OpposingDepletion, 0.1, 1), Value: plr.OpposingDepletion},
		{Name: "MicropriceResistance", Level: band(plr.MicropriceResistance, 0.00001, 0.0001), Value: plr.MicropriceResistance},
	}
	s.Why = Why(s)
	return s
}

func yn(v bool) string {
	if v {
		return "YES"
	}
	return "NO"
}

func band(v, lo, hi float64) string {
	if v <= 0 {
		return "NONE"
	}
	if v < lo {
		return "LOW"
	}
	if v < hi {
		return "MEDIUM"
	}
	return "HIGH"
}

// Why is deterministic prose from feature values. No LLM.
func Why(s Snapshot) string {
	switch {
	case s.Status == Unavailable:
		return "Absorption evidence is unavailable because the book is not synced."
	case s.Status == InsufficientData:
		return "Book data is too stale to classify passive-liquidity response."
	case s.V1Class == exhaustion.ClassExhaustion:
		return "POTENTIAL SELL OR BUY EXHAUSTION. Legacy direction is opposed by elevated aggression. Price-impact efficiency is reduced. Passive-liquidity response is observed separately and does not change FLOW_EXHAUSTION_V1."
	case s.V1Class == exhaustion.ClassContinuation:
		return "FLOW CONTINUATION CONTROL. Aggressive flow is aligned with Legacy. The same book-response features are stored so continuation can be compared with exhaustion."
	default:
		return "FLOW NEUTRAL. Directional pressure is below the frozen V1 threshold of 15. No execution implication."
	}
}
