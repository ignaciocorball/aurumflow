# P8.2 — Market history warmup

CLOSED does not block warmup. Broker `GetPrices` is read-only. If the broker returns no bars while closed:

```text
HISTORY_UNAVAILABLE_WHILE_CLOSED
```

That is not an error. Retry after `TRADEABLE`.

## Status

| Status | Meaning |
|---|---|
| READY | M5≥50, H1≥20, H4≥10 and latest bar younger than 36h |
| WARMING | some bars, not yet READY |
| STALE | latest bar older than 36h |
| HISTORY_UNAVAILABLE_WHILE_CLOSED | broker returned empty/unavailable |
| UNKNOWN | no attempt yet |

Per market: bar counts (M5 / M15 / H1 / H4), oldest, latest, gaps, stale.

## FeatureReadiness

No feature runs on insufficient history.

| Flag | Requires |
|---|---|
| PRICE_LIVE | live TRADEABLE quote, not stale |
| MOMENTUM_READY | PRICE_LIVE + 1h return + non-closed momentum |
| VOLATILITY_READY | PRICE_LIVE + 1h vol state |
| CROSS_ASSET_READY | PRICE_LIVE + relative strength vs live peers |
| LEGACY_READY | history Status=READY |

Closed markets may still be WARMING/READY on history. Momentum is not ranked live from a CLOSED quote.

## WorldState layers

`SlowContextSnapshot` (`SlowAt`) + `LiveMarketFrame` (`LiveAt`) + `MicrostructureFrame` (`MicroAt`). Separately timestamped. **WORLD_HASH_V1 is unchanged** — layer timestamps are not hashed. No V2.
