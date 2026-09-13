package exhaustion

// PassiveLiquidityEvidence is the future L2/book bridge.
// It is never populated from trade-only historical data.
type PassiveLiquidityEvidence struct {
	Available         bool    `json:"available"`
	BidReplenishment  float64 `json:"bid_replenishment,omitempty"`
	AskReplenishment  float64 `json:"ask_replenishment,omitempty"`
	BidDepletion      float64 `json:"bid_depletion,omitempty"`
	AskDepletion      float64 `json:"ask_depletion,omitempty"`
	QueuePersistence  float64 `json:"queue_persistence,omitempty"`
	BookImbalance     float64 `json:"book_imbalance,omitempty"`
	RepeatLevelHits   float64 `json:"repeat_execution_same_level,omitempty"`
}

// AbsorptionEvidence is FlowExhaustion + valid passive liquidity.
// Trade-only engines must leave Book.Available=false and must not emit ABSORPTION_CONFIRMED.
type AbsorptionEvidence struct {
	FlowExhaustion Snapshot
	Book           PassiveLiquidityEvidence
}

func TradeOnlyAbsorption(s Snapshot) AbsorptionEvidence {
	return AbsorptionEvidence{FlowExhaustion: s, Book: PassiveLiquidityEvidence{Available: false}}
}
