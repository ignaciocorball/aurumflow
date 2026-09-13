# V2 candidate findings (exploratory only)

Status: `EXPLORATORY_MECHANISM_RESEARCH`.  
No `FLOW_EXHAUSTION_V2` spec is created. A future V2 would need a new frozen spec and a still-untouched validation dataset (`PRE_2026_03_17` remains reserved).

## What looked real on both examined BTCUSDT windows

1. **FlowMagnitudeNormalized** — Exhaustion ≫ Neutral (discovery 4.55 vs 1.54; holdout 4.44 vs 1.41). Matched-control mean differences +3.29 and +2.52. CONSISTENT.
2. **ImpactFailure** — Exhaustion ≫ Neutral (3.08 vs 0.87; 2.86 vs 0.80). Matched-control differences +2.83 and +1.62. CONSISTENT.
3. **FlowEfficiencyNormalized** — Exhaustion more negative than Neutral. CONSISTENT.

These are pre-signal, past-only descriptors. They were not optimized against forward returns in this run.

## What did not survive as a mechanism claim

- **Raw price displacement vs flow** is *larger* on Exhaustion than Neutral. The inefficiency is relative (displacement does not keep up with extra aggression), not an absolute quiet tape.
- **CVD level** sign flips between discovery (+) and holdout (−). INCONSISTENT. Do not promote.
- **1-minute lumped trade velocity** is an artifact of the research tape (two synthetic prints/minute). Use path-level radar velocity if a V2 ever needs intensity.
- **Continuation** also lives in a high-flow / high-ImpactFailure region on the holdout. A V2 that merely thresholds ImpactFailure would not isolate Exhaustion from Continuation.

## P5.3 event-resolution (still not a V2)

True timestamp TradeVelocity is now informative. Discovery Exhaustion 1s mean 26.7 vs Neutral 9.2 vs Continuation 7.7. Holdout Exhaustion 1s 17.5 vs Neutral 13.0, but Continuation 1s 32.5 is hotter. Short-window velocity does not isolate Exhaustion from Continuation on the confirmatory window.

Do not promote a velocity threshold. Same rule as ImpactFailure.

## If a V2 is ever specified later

A candidate shape, not a spec:

```text
Legacy valid
AND DP <= -15                 (keep frozen V1 membership)
AND FlowMagNorm high
AND ImpactFailure high
```

That still requires a new hash, a pre-registered threshold chosen without the next holdout, and `PRE_2026_03_17` left closed until then. Do not use the prospective stream to retune V1.
