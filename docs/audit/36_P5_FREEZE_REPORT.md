# 36 — P5 Freeze Report

```text
P5 FREEZE: READY_RESEARCH_BASELINE
```

Operational:

- EventReplay MAX_SPEED deterministic
- No-lookahead tests pass
- Historical BTC run complete (89 days, 107,444,915 aggTrades)
- Legacy / Radar / Fusion benchmarks complete
- Forward labels + walk-forward / OOS complete
- Quant report generated with real N

Hypotheses (15m, this sample):

```text
H1 Radar alignment improves Legacy conditional outcomes : NOT SUPPORTED
H2 |pressure| orders forward outcomes                   : NOT SUPPORTED
H3 Fusion aligned beats Legacy OOS                      : NOT SUPPORTED
```

Negative results are the finding. Radar-as-filter did not help. Trade-only PressureScore did not produce a usable high-pressure tail. Fusion aligned OOS n=10 is worse than Legacy OOS.

See `research/reports/FREE_RESEARCH_BASELINE.md`.
