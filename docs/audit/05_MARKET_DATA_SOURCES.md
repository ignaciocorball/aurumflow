# 05 — Market Data Sources

**Do not invent integrations.** Only sources with code or local artifacts are listed.

## Matrix

| Fuente | Tipo | Realtime | Histórico | Instrumentos | Implementada | Usada |
|--------|------|----------|-----------|--------------|--------------|-------|
| Capital.com REST | Broker CFD API | Poll 60s (not stream) | `GetPrices` + `--backtest-from/to` | Any epic the account can see | Yes | Yes (live path + download) |
| Local JSON candles | File | No | Yes | GOLD, US100, US500, US30, ETHUSD, MSFT, SILVER, NDAQ, MNQ1, lab | Yes | Backtest |
| Local CSV candles | File | No | Yes (`LoadCandlesFromFile`) | Any if formatted | Yes | Backtest if path is CSV |
| `lab/candles.json` | Capital dump | No | Sample | Looks like gold (~2034 bid) | Yes | Lab / sample |
| Firebase RTDB | Telemetry sink | Status snapshots | Backtest analytics (scripts) | n/a | Uncommitted Go + gitignored Python | Optional |
| Redis | Dedup KV | n/a | n/a | n/a | Yes | Notifications only |
| MetaTrader / MQL5 | — | — | — | — | **No** | No |
| Yahoo | — | — | — | — | **No** | No |
| Alpha Vantage | — | — | — | — | **No** | No |
| Polygon | — | — | — | — | **No** | No |
| Finnhub | — | — | — | — | **No** | No |
| TwelveData | — | — | — | — | **No** | No |
| Binance | — | — | — | — | **No** | No |
| CME | — | — | — | — | **No** | No |
| Nasdaq exchange | — | — | — | — | **No** | No |
| Cboe | — | — | — | — | **No** | No |
| TradingView | RSI Wilder “matches TradingView” comment only | — | — | — | **No feed** | No |
| PostgreSQL / SQLite / Mongo / ClickHouse | — | — | — | — | **No** | No |

## Capital.com data semantics

Evidence: `internal/market/prices.go`.

- OHLC uses **bid** only (`OpenPrice.Bid` …).
- Ask exists on the wire and is discarded for candles.
- `lab/candles.json` shows both bid and ask (spread ~1.75 on that gold sample).
- `lastTradedVolume` is stored on `Candle.Volume` and **not used** by the composer.
- Snapshot times: `snapshotTimeUTC` parsed as `2006-01-02T15:04:05`.

This is **CFD candle data**, not CME futures ticks, not Nasdaq ITCH, not OPRA.

## Instrument identity (CRITICAL)

The same economic idea is represented as **different products**. This repo uses Capital CFD epics.

| Wanted Radar symbol | In this repo? | Actual representation | Status |
|---------------------|---------------|----------------------|--------|
| XAUUSD | Name in README / bots/xauusd | Capital gold CFD; default search `"gold"` → COMMODITIES epic (often `GOLD`) | `PARTIAL` (name vs epic) |
| GOLD | Backtest files `GOLD_*`, bot `-Epic GOLD` | Capital gold CFD | `IMPLEMENTED_UNVERIFIED` |
| GC | No | — | `NOT_IMPLEMENTED` |
| NAS100 / Nasdaq | Bot US100 README: “Tech 100 / NASDAQ 100” | Capital **US100 CFD**, not NQ | `PARTIAL` |
| US100 | `bots/us100`, `backtesting/us100` | Capital index CFD | `IMPLEMENTED_UNVERIFIED` |
| USTECH | No string | — | `NOT_IMPLEMENTED` |
| NQ | No | — | `NOT_IMPLEMENTED` |
| MNQ1 | `backtesting/mnq1/` folder | Local research folder; **not** a CME adapter | `UNKNOWN` (likely another Capital epic or alias) |
| QQQ | No | — | `NOT_IMPLEMENTED` |
| SPX | No | — | `NOT_IMPLEMENTED` |
| US500 | `bots/us500` | Capital S&P 500 CFD, not ES | `PARTIAL` |
| ES | No | — | `NOT_IMPLEMENTED` |
| NDAQ | `backtesting/ndaq` | Capital **NDAQ stock CFD**, not the index | `PARTIAL` / easy to confuse |
| ETHUSD | `bots/eth*` | Crypto CFD | `IMPLEMENTED_UNVERIFIED` |
| MSFT | `bots/msft` | Stock CFD | `IMPLEMENTED_UNVERIFIED` |
| SILVER | backtest files | Commodity CFD | `IMPLEMENTED_UNVERIFIED` |
| US30 | `backtesting/us30` | Dow CFD | `IMPLEMENTED_UNVERIFIED` |
| DXY / yields / VIX | No | — | `NOT_IMPLEMENTED` |

**Conclusion:** the system trades **Capital.com CFDs**. It does not consume futures, does not know GC/NQ/ES, and can confuse Nasdaq-the-index (US100) with Nasdaq-the-stock (NDAQ).

## How an instrument is selected

1. `AURUMFLOW_EPIC` or `run.ps1 -Epic` if set.
2. Else `ResolveEpic("gold")` — first COMMODITIES match.
3. Backtest download uses `--backtest-epic` or same env, else gold.

There is **no instrument master**, tick size table, or contract spec. `value_per_point` defaults to 1.0 for every epic — incorrect for most index CFDs if copied from gold.

## Realtime claim

Polling every 60 seconds is **not** realtime microstructure. Market snapshot bid/offer is available via `GetMarketDetails` and used for optional spread check and TRADEABLE status only.
