# ATTENTION_SCORE_V1 audit (no tuning)

Spec: `internal/opportunity` `ATTENTION_SCORE_V1`.  
Weights frozen. This document does not change them.

## What V1 actually measures

**Evidence completeness + known-state bonuses**, not move size.

`knownBonus(s)` is 1 for any non-UNKNOWN string.  
`momBonus` is 1 for `UP|DOWN|STRONG_UP|STRONG_DOWN|FLAT` — **FLAT and STRONG_DOWN award the same 15 Momentum points**.

So Attention is primarily:

1. which institutional/context fields are **observed** (CFTC, live quote, gold/silver ratio)
2. whether price trend is **classified at all**
3. data-quality HEALTHY (10 pts)
4. microstructure available (10 pts, BTC only)

It is **not** market activity, not |return|, not institutional *direction*.

## Coupling

Coverage = fraction of those same components that are “known”.  
Attention = weighted sum of the same known flags.  
Flag: **ATTENTION_COVERAGE_COUPLING** (partial). They are not identical (weights differ) but both ignore magnitude. Ranks stayed static for 2h because **no component flipped known→unknown** and Momentum does not see −0.78% vs +0.12%.

## Forensic (open / +1h / +2h) — components that awarded points

Typical TRADEABLE metal/index at this soak (no CFTC for indices):

| Component | Weight | GOLD | SILVER | J225 | US100 |
|---|---|---|---|---|---|
| MacroAlignment | 15 | 0 (UNKNOWN) | 0 | 0 | 0 |
| CapitalFlowAlignment | 15 | 0 (UNKNOWN) | 0 | 0 | 0 |
| PositioningAsymmetry | 15 | 15 (CFTC) | 15 (CFTC) | 0 | 0 |
| CrossAssetConfirmation | 10 | 10 (RS observed) | 10 | 10 | 10 |
| Momentum | 15 | 15 (any classified trend) | 15 | 15 (incl. STRONG_DOWN) | 15 |
| VolatilitySuitability | 10 | 10 | 10 | 10 | 10 |
| MicrostructureReadiness | 10 | 0 | 0 | 0 | 0 |
| DataQuality | 10 | 10 | 10 | 10 | 10 |
| **Total** | | **60** | **60** | **45** | **45** |

What changed over 2h: PriceTrend label (FLAT/UP/DOWN/STRONG_DOWN) and mids.  
What did **not** change: known-flags → **score identical**.

CN50 CLOSED: DataQuality degraded, momentum not live → Attention 0 / IGNORE. Closed markets correctly stop contributing live momentum.
