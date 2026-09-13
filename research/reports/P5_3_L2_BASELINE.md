# P5.3 L2 baseline

No performance claim. Live public L2 from this run.

```text
PRIMARY_L2_SENSOR = binance_usdm_public
Instrument        = BTCUSDT
Relation          = CORRELATED_PROXY to Capital BTCUSD CFD (not IDENTICAL)
Authentication    = none
Trading capability= none
```

Binance WSS (official 2026 split): OPERATIONAL after replacing retired `/ws`.
OKX public SWAP: also OPERATIONAL (alternate, unused as primary this soak).

60-minute SHADOW soak completed (no panic):

```text
duration        60m
trades          8563
book snapshots  3598   (REST + applySnapshot; ~1 Hz book)
wss depth deltas 0     (public/stream depth not counted this soak)
gaps            0
drops           0
p50 latency     145.6 ms
p95 latency     399.7 ms
peak alloc      12.5 MB
panics          0
prospective     0 Legacy / 0 Exhaustion / 0 Continuation
broker mutations 0
```

Book features (spread, microprice, imbalance, depletion/replenishment/persistence) run on the snapshot book. WSS aggTrades are live. WSS depth deltas were not observed in this soak — treat L2 as snapshot-MBP, not 100ms incremental, until that stream is confirmed.

Prospective Exhaustion vs Continuation L2 comparison is not yet populated. Do not claim Absorption improves Exhaustion.
