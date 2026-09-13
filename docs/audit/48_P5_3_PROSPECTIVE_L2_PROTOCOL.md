# 48 — P5.3 Prospective L2 protocol + concurrent runbook

## Recorders

```text
research/prospective/FLOW_EXHAUSTION_V1/   V1 + microflow + book/absorption fields
research/prospective/L2_MECHANISM/         same inputs, Exhaustion vs Continuation L2 study
```

Inputs immutable. Outcomes append-only. Duplicate `signal_id` refused.

Deterministic id: `instrument|RFC3339Nano|dir|legacy`

Event window t−60s…t+60s is referenced by raw segment paths, not duplicated.

## Raw capture

```text
data/live/<provider>/<instrument>/YYYY/MM/DD/
  HH-trades.jsonl.gz + .meta.json
  HH-depth.jsonl.gz
  HH-snapshots.jsonl.gz
  HH-health.jsonl.gz
```

Gitignored. Each segment records provider, times, sequences, events, gaps, resyncs, checksum.

Checkpoint: `data/live/shadow-checkpoint.json` (IDs only; book is not treated as live after restart).

## Concurrent processes

```text
Process A  GOLD Capital DEMO Legacy
           --demo-week --epic GOLD --status-addr 127.0.0.1:8765
           journals/ from the execution loop

Process B  BTC intelligence SHADOW
           --shadow-runtime --l2-provider binance --soak-minutes 60 --status-addr 127.0.0.1:8766
           data/live/ + journals/shadow-runtime-metrics.json
           no ExecutionProvider

Process C  status / reporting
           --status --status-addr 127.0.0.1:8765   (GOLD)
           --status --status-addr 127.0.0.1:8766   (SHADOW)
```

Distinct status ports avoid bind conflicts. Distinct journal/data roots avoid file locks.

GOLD week still: Legacy execution only. Radar / Exhaustion / Absorption remain SHADOW.
