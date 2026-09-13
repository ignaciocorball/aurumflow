# 55 — P5.4 + P6-lite Freeze Report

```text
P5.4 FREEZE: READY
P6-LITE FREEZE: READY
Tests before: 225 PASS
Tests after:  238 PASS / 0 FAIL
Vet: PASS
Build: PASS
```

V1 hash, threshold 15, PressureScore, Legacy unchanged. Radar / Exhaustion / Absorption remain SHADOW. PRE_2026_03_17 untouched. No AUTO_EXECUTE. No LIVE.

## True L2

```text
PRIMARY: binance_usdm_public BTCUSDT
ENDPOINT: wss://fstream.binance.com/public/ws/btcusdt@depth@100ms
Auth: none
Relationship to Capital BTCUSD: CORRELATED_PROXY
Relationship to V1 trades: SAME_VENUE
```

P5.3 failure mode: REST 1 Hz snapshots treated as the live book. Fixed: REST is resync-only; incremental diffs publish as `KindBookDelta`.

70-minute soak (exit 0):

```text
trades=18277
deltas=41094
snapshots=4
gaps=0
drops=0
panics=0
p50=115.5 ms
p95=289.7 ms
peak_mem=24.4 MB
disk≈10 MB including journals
```

OKX `BTC-USDT-SWAP` public instruments = live. Kept as alternate / cross-venue sensor.

## Console

`http://127.0.0.1:8766/` read-only. Safety bar present. No mutation endpoints.

## Central result

We can now capture incremental passive-liquidity response around future V1 signals, with provenance, without waiting for n=25 or GOLD TRADEABLE.
