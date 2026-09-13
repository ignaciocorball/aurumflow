# 25 — P3 Microstructure Sandbox

This is a weekend BTC sandbox. It is not CME NQ/GC.

```text
Sensor: Binance USD-M BTCUSDT public market data
Execution: Capital BTCUSD CFD (separate instrument)
Relationship: proxy — not IDENTICAL
```

## Adapter

`internal/binanceusdm`

```text
REST snapshot: GET https://fapi.binance.com/fapi/v1/depth?symbol=BTCUSDT&limit=1000
WS: wss://fstream.binance.com/stream?streams=btcusdt@depth@100ms/btcusdt@aggTrade
Auth: none
Trading: none
Private API: none
```

Lifecycle: connect, subscribe, ping/pong, exponential reconnect (cap 60s), 23h rotation, context cancel, health events.

## Local book

```text
1. start depth stream and buffer
2. REST snapshot
3. discard obsolete u <= lastUpdateId
4. first event: U <= lastUpdateId+1 <= u
5. later events: pu continuity
6. gap → unsync + resync counter
```

Exposed: `book_synced`, `last_update_id`, `resync_count`, `book_age`.

No book-derived signal is valid while `book_synced=false`.

## Flow / CVD / book features

- Aggressor from aggTrade `m` (buyer is maker ⇒ aggressive sell)
- Rolling CVD
- Windows conceptually 1s/5s/30s/1m/5m (engine + soak snapshots)
- Imbalance top 5/10/20, microprice, spread, mid, depth
- Depletion / replenishment / persistence require time evidence
- `POTENTIAL_ABSORPTION` only — not institution identity
- Thresholds use relative/typical quantity, not hardcoded “100 BTC”
