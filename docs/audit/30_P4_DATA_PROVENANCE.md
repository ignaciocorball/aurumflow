# 30 — P4 Data Provenance

Every `md.Event` may carry `Provenance`:

```text
Provider Venue Instrument
Quality: DIRECT | PROXY | DELAYED | EXPERIMENTAL
FreshnessClass: DIRECT | PROXY | DELAYED | SLOW_CONTEXT | EXPERIMENTAL
EventTime ReceivedAt PublishedAt AvailableAt
IsProxy IsDelayed
```

Rule: a context observation is usable at research time `t` only if `AvailableAt <= t`.

| Dataset | EventTime | AvailableAt policy |
| --- | --- | --- |
| Binance aggTrades | trade timestamp | same as event (DIRECT) |
| Capital OHLC | candle open/close | broker published time (when download succeeds) |
| SEC filings / facts | report period | `filed_at + 24h` conservative |
| CFTC COT | Tuesday as-of | `asOf + 4d` (Friday publish, Saturday UTC conservative) |
| FINRA weekly OTC | week ending | week itself; missing ≠ 0 |
| FRED | observation period | series vintage / available_at when key present |

COT Tuesday as-of is not available on Tuesday.

Tests: `internal/md/provenance_test.go`, `internal/secctx`, `internal/cftc`, `internal/macro`, `internal/ctxsnap`.
