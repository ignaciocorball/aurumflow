# 44 — P5.2 Freeze Report

```text
P5.2 FREEZE: READY
HEAD at start: e8183c5
Tests before: 188 PASS
Tests after:  210 PASS / 0 FAIL
Vet: PASS
Build: PASS
```

## Invariants held

- V1 spec hash unchanged
- Threshold 15 unchanged
- PressureScore formula unchanged
- AurumFlow Legacy unchanged
- Radar / Exhaustion remain SHADOW
- GOLD DEMO week strategy path unchanged
- BTC_FLOW_EXHAUSTION_V1 = VALIDATED_EXTERNAL_HOLDOUT
- GOLD_FLOW_EXHAUSTION = UNVALIDATED
- US100_FLOW_EXHAUSTION = UNVALIDATED
- PRE_2026_03_17 remains UNTOUCHED
- No FLOW_EXHAUSTION_V2 spec

## Engine

`internal/exhaustion` classifies V1, computes past-only FlowEfficiency and ImpactFailure, and cannot mutate a broker.

## Mechanism

243,352,328 historical trades streamed in 4m18s (~944k events/s, peak alloc 108 MB). Discovery n=56 and holdout n=50 both answer YES to stronger aggression + weaker *normalized* impact versus Neutral/matched Legacy prints.

## Invalid rows

Previous holdout `invalid_rows=90` was one Binance CSV header per daily file (`agg_trade_id`). Headers are now counted as `Headers`, not malformed trades. Rescan: headers=90, invalid=0, rows=135,907,413.

## Prospective

Recorder, immutable inputs, append-only outcomes, and status CLI exist. Live counts start at zero on 2026-09-13.
