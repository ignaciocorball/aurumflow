# 45 — P5.3 Event-resolution flow

Package: `internal/microflow`

Feature family: `MECHANISM_EVIDENCE_V1` (not `FLOW_EXHAUSTION_V2`).

V1 membership is unchanged. These windows never enter `ClassifyV1`.

## Windows

Exact rolling windows on `aggTrade` EventTime:

```text
1s  5s  15s  30s  1m  3m  5m
```

Each window reports trade count, trades/sec, aggressive buy/sell qty and notional, signed qty/notional, imbalance, CVD delta, notional velocity, quantity velocity.

`Snapshot(t0)` uses only prints with `T < t0`.

Replay and live share the same engine.

## Artifact fixed

P5.2 TradeVelocity ≈ 0.033 was two synthetic prints per 1-minute candle. Historical re-run streams real timestamps via `binancehist.IterTrades`.

Label: `EXPLORATORY_MECHANISM_ONLY`.
