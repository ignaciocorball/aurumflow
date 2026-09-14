# BTC drop incident 2026-09-14

## Counter

`dropped_events` on `/status` is **`md.Bus.dropped`**, not WebSocket transport loss, not UI SSE, not book sequence gaps.

| Field | Value |
|---|---|
| Counter | `md.Bus.dropped` → `ops.Status.Drops` |
| Package | `internal/md` |
| Producer | Binance USDM (and optional second L2) `Provider.Run` → `bus.Publish` |
| Queue | `md.NewBus(8192)` single channel |
| Consumer | one `shadowrt` `select` loop (book + microflow + radar + exhaustion + segmented recorder + rawbuf) |
| Event types | whatever was in the channel at overflow (`trade` / `book_delta` / …) |
| Increment | `Publish` `default` branch: **backpressure** (also `ctx_cancel`) |

## Timeline

- First observed 0: ~2026-09-13T23:27Z (`trades≈75k`, `deltas≈48k`, `gaps=0`, `resyncs=1`)
- Last observed 966: ~2026-09-13T23:52Z through 00:07Z (counter frozen after burst)
- Total drops: **966**
- Book: **synced**, **gaps=0**, resyncs **2**

## Why drops>0 and gaps=0 and book synced can coexist

**Proven, not inferred:** `internal/binanceusdm` applies `Book.ApplyFuturesDelta` / `ApplySnapshot` in the provider goroutine **before** `bus.Publish`. The 8192 (now 32768) bus is a downstream fan-out to microflow, radar, exhaustion, rawbuf, and the segmented recorder. Overflow cannot unsync the book. `Gaps` increment only on depth sequence mismatch. `gaps=0` plus `book_synced=true` is therefore expected when the dropped events never reached those downstream consumers.

The soak also showed a sticky `book_unsynced` alert while `book_synced=true` — that alert was never cleared after recovery (status-plane bug, fixed in P8.8).

Startup handshake typically records two `IncResyncs` with **zero gaps**: `buffered_not_applicable` then `rest_unsynced`. That is not a runtime sequence incident.

## Raw capture

`rawbuf.Add` and `collector.WriteEvent` run **after** dequeue. The 966 never reached disk. Interval is **DEGRADED** (`BUS_BACKPRESSURE`), not reconstructable from raw.

## Impact

| Surface | Affected |
|---|---|
| Core book sequence | No (gaps=0, synced) |
| Raw capture / research tape | Yes — 966 events missing |
| Microflow / V1 trade features | Yes if dropped kinds include `trade` |
| UI SSE | No (not this counter) |

## Root cause

Single slow consumer on an 8192 bus during a trade burst. Disk writes in the same loop increase dequeue latency.

## Fix in P8.8

- Attributed drops (`drops_by_event_type`, `drops_by_provider`, `drops_by_reason`)
- Queue capacity / depth / high-water / enqueue and dequeue rates
- Async persist queue (65536) so disk IO no longer blocks the live consume loop
- Bus capacity 32768
- `DataIntegrityInterval` DEGRADED on bus or persist loss
- Loss class: bus/recorder/rawbuf/book-depth = `LOSSLESS_REQUIRED`; book-side trades = `LOSS_TOLERANT`; UI = `DISPLAY_ONLY`

## Residual risk

If persist high-water hits 65536, research tape degrades while the book stays live. After restart, watch `drops_by_kind`, `bus_high_water`, `persist_drops`, and `integrity_status`.

## P8.8 post-restart (2026-09-14T00:28Z soak)

`dropped_events=0` `persist_drops=0` `bus_high_water≈200` `capacity=32768` `gaps=0` `integrity=VALID`.  
Consumer p95 ~6ms. Enqueue rate equals dequeue rate. Startup resync reason: `rest_unsynced` only.

## P8.8 post-restart (2026-09-14T00:28Z soak)

`dropped_events=0` `persist_drops=0` `bus_high_water≈200` `capacity=32768` `gaps=0` `integrity=VALID`.  
Consumer p95 ~6ms. Enqueue rate equals dequeue rate. Startup resync reason: `rest_unsynced` only.
