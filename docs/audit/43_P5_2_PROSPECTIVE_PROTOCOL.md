# 43 — P5.2 Prospective Protocol

Directory:

```text
research/prospective/FLOW_EXHAUSTION_V1/
  signal_inputs.jsonl      immutable
  signal_outcomes.jsonl    append-only, keyed by signal_id
```

An input row is written at signal time with `outcome_known=false`, V1 class, PressureScore, DirectionalPressure, pre-signal feature blob, git commit, `feature_version=P5.2-MECH-1`, and spec hash. Duplicate `signal_id` is refused. Inputs are never edited.

Outcomes are appended later:

```text
go run ./cmd/bot --label-prospective
```

Horizons: +1m, +5m, +15m, +30m, +1h plus MFE/MAE. Labeling cannot rewrite the input file.

Status:

```text
go run ./cmd/bot --prospective-status
```

Checkpoints `n=25/50/100/200` are observation milestones only. They do not retune V1 and do not promote execution.

BTCUSDT only. GOLD/US100 signals must not be written into this stream.

Start date: 2026-09-13. Stream is OPEN. Counts at freeze: 0 observed.
