# FLOW_EXHAUSTION_V1

Pre-registered before EXTERNAL_HOLDOUT_1 download and before any holdout outcome inspection.

- spec_id: `FLOW_EXHAUSTION_V1`
- created_at: `2026-09-13T04:30:00Z`
- git_commit: `e8183c55e32a11bfddb49e641c3af66968443a2d`
- feature_version: `radar_trade_flow_pressure_v1+composer_crypto_research_v1`

## Datasets

Discovery (already inspected; not confirmatory):

```text
DISCOVERY_DATASET
BINANCE_BTCUSDT_USDM
2026-06-15 → 2026-09-11
```

Any V1 numbers on that range are `POST_HOC_DISCOVERY`.

External holdout (confirmatory only):

```text
EXTERNAL_HOLDOUT_1
BINANCE_BTCUSDT_USDM
2026-03-17 → 2026-06-14
must end before 2026-06-15T00:00:00Z
```

## Frozen mechanics

PressureScore, CVD, trade velocity, flow windows, Legacy strategy/thresholds, Radar thresholds, and confidence are unchanged.

Threshold remains `|PressureScore| >= 15`.

```text
D = LegacyDirection  (+1 LONG, -1 SHORT)
P = existing PressureScore at t0
directional_pressure = D * P

D*P <= -15  → FLOW_EXHAUSTION_CONFIRM
D*P >= +15  → FLOW_CONTINUATION_CONFIRM
|P|  <  15  → FLOW_NEUTRAL
```

Original PressureScore is stored, not overwritten.

This is not absorption. Book data is unavailable.

## Hypothesis

H0: FLOW_EXHAUSTION_CONFIRM does not improve the conditional 15-minute directional-return distribution of Legacy signals.

H1: FLOW_EXHAUSTION_CONFIRM identifies Legacy setups with a better 15-minute directional-return distribution.

Primary endpoint: 15-minute mean directional return, exhaustion vs Legacy baseline.

Secondary endpoints may not override the primary conclusion.

## Decision (holdout only)

SUPPORTED only if all of:

1. exhaustion n >= 50
2. 15m mean directional return > Legacy baseline
3. sign of (exhaustion mean − legacy mean) agrees in at least 2 of 3 chronological thirds
4. no catastrophic MAE: when legacy MAE < 0, |exhaustion MAE| <= 2 × |legacy MAE|

INCONCLUSIVE if n < 50 or subperiod signs are unstable.

NOT_SUPPORTED if n >= 50 and primary effect <= 0, or the effect materially reverses.

## Other rules

- LONG and SHORT measured separately; symmetry is not assumed.
- Magnitude buckets 15–20 / 20–40 / 40–60 / 60+ are descriptive only.
- t0−1m / t0−3m / t0−5m pressure is `POST_HOC_DELAY_DIAGNOSTIC`.
- GOLD and US100 remain UNVALIDATED regardless of BTC holdout.
- RADAR_MODE stays SHADOW. No live execution change.
