# 39 — P5.1 External Holdout

`EXTERNAL_HOLDOUT_1` was downloaded after `FLOW_EXHAUSTION_V1` was written and hashed.

```text
start          2026-03-17
end            2026-06-14
days           90 complete archives
instrument     Binance USD-M BTCUSDT aggTrades
event_count    135,907,413
duplicates     0
out_of_order   0
invalid_rows   90  (one parse reject per file typical; not used as outcomes)
dataset_hash   47cddb104c0cd014440f91d95a943e97dce13ea57363a6dc7834eb16ef632205
```

Checksums recorded in `research/reports/EXTERNAL_HOLDOUT_1_INTEGRITY.json`.

Range ends before `2026-06-15T00:00:00Z`. No overlap with DISCOVERY_DATASET.

This dataset must not be reused to retune V1.
