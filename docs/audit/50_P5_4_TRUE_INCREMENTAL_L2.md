# 50 — P5.4 True Incremental L2

P5.3 had real Binance trades plus ~1 Hz REST snapshots and **zero** incremental depth events. REST snapshot refresh was overwriting the live book, so official depth diffs never published.

## Official Binance USD-M depth (2026)

Do not use the retired `/ws` path. Current public split:

```text
trades:  wss://fstream.binance.com/market/stream?streams=btcusdt@aggTrade
depth:   wss://fstream.binance.com/public/ws/btcusdt@depth@100ms
```

Diff Book Depth update speed: 100ms. REST `/fapi/v1/depth` is snapshot/resync only.

Success requires all of:

```text
snapshot obtained
incremental events > 0
first-event bridge (U <= last+1 <= u)
continuity (pu == last)
book remains synced
gaps detectable
forced gap → resync (fresh snapshot; stale book discarded)
```

A WebSocket connect alone is not success.

## Book engine

`Snapshot` seeds levels. `Delta` applies `qty=0 → remove`. Replenishment and depletion are computed from deltas, not from treating every REST refresh as flow.

Language is MBP-honest:

```text
VISIBLE_LIQUIDITY_REMOVAL
VISIBLE_LIQUIDITY_REPLENISHMENT
```

MBP cannot separate cancel vs execution.

## OKX alternate

Public `wss://ws.okx.com:8443/ws/v5/public` channel `books` + `trades` for verified `BTC-USDT-SWAP`. Integrity is `seqId`/`prevSeqId`. Checksum is not used.

Selected primary sensor for this freeze: **Binance USD-M BTCUSDT** (same venue as V1 trade flow).
