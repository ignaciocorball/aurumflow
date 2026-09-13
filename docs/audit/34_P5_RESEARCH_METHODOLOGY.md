# 34 — P5 Research Methodology

Event study ≠ executable PnL.

Forward labels: +1m, +5m, +15m, +30m, +1h of directional return, MFE, MAE, time-to-extrema.
+10s / +30s are RESOLUTION_LIMITED on 1-minute synthesized candles and are not reported as independent horizons.

Walk-forward on this sample (proportional 60 / 15 / 15 of the 89-day span):

```text
TRAIN       2026-06-15 → 2026-08-13
VALIDATION  2026-08-13 → 2026-08-28
OUT_OF_SAMPLE 2026-08-28 → 2026-09-11
```

Fixed Composer / Radar parameters. No grid search. No parameter mining.

Costs: this report is an event-study of forward returns. It is not net strategy PnL. Mark `ESTIMATED_COST` if a fill simulator is added later. `costs_default_zero` remains true for the old candle backtester.

Command:

```text
go run ./cmd/bot --research btc-radar --from 2026-06-15 --to 2026-09-11
```

Reproducibility fields in `research/reports/FREE_RESEARCH_BASELINE.json`: git commit, date range, counts, bootstrap seed 42.
