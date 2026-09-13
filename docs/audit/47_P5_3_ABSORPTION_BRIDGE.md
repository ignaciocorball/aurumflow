# 47 — P5.3 Absorption bridge

Package: `internal/absorption`

Input at t0 only: Legacy/V1 snapshot, DirectionalPressure, FlowEfficiency, ImpactFailure, `PassiveLiquidityResponse`, book sync/age.

Statuses (descriptive, no weights, no broker effect):

```text
UNAVAILABLE
INSUFFICIENT_DATA
NOT_SUPPORTIVE
MIXED
SUPPORTIVE
STRONGLY_SUPPORTIVE
```

Components exposed independently:

- aggression against Legacy
- high ImpactFailure
- supporting replenishment
- supporting persistence
- opposing depletion
- book imbalance response
- microprice refuses aggression

`MayMutateBroker() = false`. Mode = SHADOW.

ImpactFailure alone does not identify Exhaustion (P5.2 continuation also high). The bridge exists so live L2 can later discriminate Exhaustion vs Continuation. No performance claim until prospective sample matures.
