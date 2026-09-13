# 52 — P5.4 Absorption Observation

Absorption is SHADOW observation. It does not change `FLOW_EXHAUSTION_V1`, threshold 15, PressureScore, or Legacy.

Statuses (unchanged names):

```text
UNAVAILABLE
INSUFFICIENT_DATA
NOT_SUPPORTIVE
MIXED
SUPPORTIVE
STRONGLY_SUPPORTIVE
```

Rules (t0, past-only):

```text
book not synced            → UNAVAILABLE
book age > 5s              → INSUFFICIENT_DATA
else count present components:
  0      NOT_SUPPORTIVE
  1–2    MIXED
  3–4    SUPPORTIVE
  ≥5     STRONGLY_SUPPORTIVE
```

Components (transparent):

```text
aggression_against_legacy
high_impact_failure
supporting_replenishment
supporting_persistence
opposing_depletion
book_imbalance_response
microprice_refuses_aggression
```

Evidence labels are HIGH/MEDIUM/LOW/YES/NO. No opaque trading score.

Continuation control stores the same book-response features so the scientific question remains:

> what distinguishes aggressive flow that continues from aggressive flow that exhausts?

Post-t0 book data may be retained for research windows (`t-60s`…`t+60s`) but must not enter `AbsorptionEvidenceStatus` at t0.

`MayMutateBroker() = false`.
