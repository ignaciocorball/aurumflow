# 35 — P5 A/B/C Experiment

Sensor: Binance USD-M BTCUSDT (not Capital BTCUSD). Execution plane does not participate.

```text
A Legacy  — same Composer + state machine on synthesized 15m/H1/H4, Crypto=true
B Radar   — TRADE_FLOW_RADAR on aggTrades (book caps off)
C Fusion  — ALIGNED_* / CONTRADICTED_* / LEGACY_ONLY / RADAR_ONLY
            align threshold = |pressure| >= 15 (transparent, not mined)
```

Dataset: 2026-06-15T00:00:00Z → 2026-09-11T23:58:00Z (89 days).
Trades: 107,444,915. Radar snapshots: 128,071.

## Primary 15m event study

| Cohort | N | Hit | Mean dir | Median | MFE/MAE |
| --- | ---: | ---: | ---: | ---: | ---: |
| Legacy all | 928 | 0.541 | +0.00010 | +0.00008 | 1.00 |
| Legacy long | 491 | 0.525 | +0.00005 | +0.00005 | 0.98 |
| Legacy short | 437 | 0.558 | +0.00015 | +0.00013 | 1.03 |
| Legacy aligned with Radar | 33 | 0.242 | −0.00073 | −0.00064 | 0.37 |
| Legacy contradicted by Radar | 56 | 0.750 | +0.00106 | +0.00063 | 3.41 |
| Radar directional | 12410 | 0.471 | −0.00005 | −0.00008 | 0.99 |
| Fusion aligned | 33 | 0.242 | −0.00073 | −0.00064 | 0.37 |

Bootstrap 95% CI on aligned mean: [−0.00112, −0.00018].
Bootstrap 95% CI on contradicted mean: [+0.00054, +0.00161].
Those intervals do not overlap.

## Filter test

Excluding Radar contradictions: N 928→872, hit 0.541→0.528, mean +0.00010→+0.00004, MFE/MAE 1.00→0.91.

Does excluding contradictions improve Legacy? **NO.**

## Pressure / confidence / state

Pressure mass sits in 0–20 (n=873). 20–40 n=39, 40–60 n=16, 60–100 n=0. Hit is not monotonic. High absolute pressure is rare on trade-only snapshots.

Confidence is not calibrated: all 928 Legacy rows fall in 80–101.

ABSORPTION = 0 (unavailable without historical book). EXPANSION n=40, hit 0.550, mean +0.00091.

## OOS

| | N | Hit | Mean |
| --- | ---: | ---: | ---: |
| Legacy OOS | 176 | 0.489 | +0.00008 |
| Radar OOS | 2352 | 0.481 | −0.00007 |
| Fusion aligned OOS | 10 | 0.400 | −0.00027 |

Reports: `research/reports/FREE_RESEARCH_BASELINE.md` and `.json`.
