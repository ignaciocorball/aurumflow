# 42 — P5.2 Mechanism Study

Role: `EXPLORATORY_MECHANISM_RESEARCH`.

This is not a new performance validation. Discovery and `EXTERNAL_HOLDOUT_1` are reported separately. Their 15m returns are not pooled.

Command:

```text
go run ./cmd/bot --research exhaustion-mechanism
```

## Central question

> Do `FLOW_EXHAUSTION_CONFIRM` events show stronger aggressive flow but weaker price impact than comparable Legacy signals?

Answered independently on each dataset, then overall.

Comparable groups: Neutral Legacy signals, plus stratified nearest unused controls matched on Legacy direction, UTC hour, rolling-vol tertile, and Legacy score bucket. No future returns enter the match.

## Result

```text
DISCOVERY          YES
EXTERNAL_HOLDOUT   YES
OVERALL            YES
```

The YES is carried by normalized descriptors, not by a smaller raw return:

- FlowMagnitude (robust z vs past-only median/MAD) is ~3× Neutral on both datasets.
- ImpactFailure (FlowMagNorm − PriceDispNorm) is similarly elevated vs Neutral and vs matched controls.
- Raw 5m displacement-vs-flow is *not* smaller than Neutral. Exhaustion prints more flow *and* a somewhat larger raw tick. The inefficiency is that displacement does not scale with the extra aggression.

Continuation also sits in a high-flow region, especially on the holdout. The distinctive contrast is Exhaustion vs ordinary Neutral Legacy prints.

CVD *level* is INCONSISTENT across datasets and is not a candidate mechanism. Trade-velocity from 1-minute lumped tape is uninformative (two synthetic prints per minute); path-level radar velocity remains the usable intensity series.

PRE_2026_03_17 was not opened.
