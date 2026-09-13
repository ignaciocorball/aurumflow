# 33 — P5 Event Replay

`internal/replay` consumes the same `md.Event` model as the live intelligence plane.

Modes: REALTIME, ACCELERATED, MAX_SPEED, STEP. Research uses MAX_SPEED; original event timestamps are preserved.

Ordering (deterministic, no goroutine sort):

```text
EventTime, then Seq, then Provider
```

Clocks: `LiveClock` / `ReplayClock` in `internal/clock`. Feature logic must not depend on `time.Now()` for researched decisions.

Candle synthesis: `internal/candles` builds 1m/5m/15m/1h/4h from trades so Legacy Composer runs on the same BTCUSDT tape as trade-flow Radar.

Trade-only Radar: `radar.NewTradeFlowEngine` — CVD / velocity / displacement only. No invented imbalance, depletion, replenishment, iceberg, or ABSORPTION from book.

89-day MAX_SPEED replay:

```text
events      107,444,915
runtime     98.5s
events/sec  1,090,928
alloc       34.1 MB
```
