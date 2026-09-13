# 29 — P4 Free Data Mesh

Official-first adapters. No scraping. No paid APIs. No LIVE broker mutations.

```text
Capital.com   historical OHLC (DEMO read-only) + streaming adapter
Binance Vision USD-M daily aggTrades + SHA256 checksum
SEC data.sec.gov submissions / companyfacts / tickers
CFTC public disaggregated COT (GOLD identity locked)
FINRA OTC weekly summary API
FRED interface (PENDING_FREE_KEY unless FRED_API_KEY)
CME public / options = NOT_CONNECTED
```

Commands:

```text
go run ./cmd/bot --download-research-data --preset free-core --research-days 90
go run ./cmd/bot --collect-market-data --collect-minutes 60
go run ./cmd/bot --collect-free-mesh --collect-minutes 60
go run ./cmd/bot --wss-diag
```

Local store: `data/` (gitignored). Catalog: `data/catalog.json`.

## Runtime this freeze

| Source | Status | Evidence |
| --- | --- | --- |
| Binance Vision aggTrades | CONNECTED | 89 complete days BTCUSDT 2026-06-15 → 2026-09-11; checksums ok; 2026-09-12 unpublished (HTTP 404) |
| Capital historical | CONNECTED_ADAPTER / RUNTIME_LIMITED | DEMO AUTH PASS; 10/10 universe epics found; `GetPrices` returned `error.prices.not-found` (weekend/closed window) |
| Capital streaming | RUNTIME_LIMITED | Adapter exists; no live CST/WSS session proven this run |
| Binance live collector | CONNECTED_PATH | `--collect-market-data` persists gzip JSONL chunks; WSS first-frame read timed out → REST fallback |
| SEC | CONNECTED | 5 companies, 205 filings, 152 facts |
| FINRA | CONNECTED | 52 weeks × AAPL/MSFT/NVDA/META/AMZN |
| CFTC | CONNECTED | GOLD `GOLD - COMMODITY EXCHANGE INC.`; current disagg file rows=2 |
| FRED | PENDING_FREE_KEY | no key in environment |
| CME / options | NOT_CONNECTED | interfaces only |

Do not treat Capital BTCUSD as Binance BTCUSDT. Do not treat GOLD CFD as CME GC.
