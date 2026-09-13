package bookfeatures

// PassiveLiquidityResponse is MBP evidence only. No participant identity.
type PassiveLiquidityResponse struct {
	SupportingReplenishment     float64 `json:"supporting_replenishment"`
	OpposingDepletion           float64 `json:"opposing_depletion"`
	SupportingPersistence       float64 `json:"supporting_persistence"`
	BookImbalanceResponse       float64 `json:"book_imbalance_response"`
	MicropriceResponse          float64 `json:"microprice_response"`
	SupportingRefillRatio       float64 `json:"supporting_refill_ratio"`
	SupportingRefillLatency     float64 `json:"supporting_refill_latency"`
	OpposingPersistentDepletion float64 `json:"opposing_persistent_depletion"`
	MicropriceResistance        float64 `json:"microprice_resistance"`
	MidpriceResistance          float64 `json:"midprice_resistance"`
	L2ProxyQuality              string  `json:"l2_proxy_quality,omitempty"`
}

// Normalize: LONG supporting=bid opposing=ask; SHORT supporting=ask opposing=bid.
func Response(legacyDir int, d Depth, liq Liquidity) PassiveLiquidityResponse {
	var r PassiveLiquidityResponse
	if !d.Available || legacyDir == 0 {
		return r
	}
	if legacyDir > 0 {
		r.SupportingReplenishment = liq.BidReplenishment
		r.OpposingDepletion = liq.AskDepletion
		r.SupportingPersistence = liq.BidPersistence
		r.BookImbalanceResponse = d.Imb5
	} else {
		r.SupportingReplenishment = liq.AskReplenishment
		r.OpposingDepletion = liq.BidDepletion
		r.SupportingPersistence = liq.AskPersistence
		r.BookImbalanceResponse = -d.Imb5
	}
	// Positive microprice response: microprice refuses the aggressive (anti-legacy) direction.
	// For LONG, aggression is typically selling; microprice holding up vs mid is supportive.
	if d.Mid > 0 {
		r.MicropriceResponse = float64(legacyDir) * (d.Microprice - d.Mid) / d.Mid
		r.MicropriceResistance = r.MicropriceResponse
	}
	if legacyDir > 0 {
		if liq.BidDepletion > 0 {
			r.SupportingRefillRatio = liq.BidReplenishment / liq.BidDepletion
		}
		r.SupportingRefillLatency = liq.BidRefillLatency
		r.OpposingPersistentDepletion = liq.AskPersistDepl
	} else {
		if liq.AskDepletion > 0 {
			r.SupportingRefillRatio = liq.AskReplenishment / liq.AskDepletion
		}
		r.SupportingRefillLatency = liq.AskRefillLatency
		r.OpposingPersistentDepletion = liq.BidPersistDepl
	}
	return r
}
