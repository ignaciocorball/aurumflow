# 10 — Storage and Data

No server database is required to run the bot.

## Stores

### 1) Config JSON / YAML

```text
Qué almacena: API host/mode, risk, strategy, notifications, telemetry paths
Quién escribe: Operator (and backtest_runner.py when generating run configs)
Quién lee: config.Load, cmd/bot
Lifecycle: file on disk; config/config.json gitignored
Schema: config.Config structs
Estado: IMPLEMENTED_UNVERIFIED
```

Secrets often sit in the same file as strategy params. See `15_SECURITY_AND_SECRETS.md`.

### 2) File logs

```text
Qué almacena: INFO/WARN/ERROR lines
Quién escribe: logger.InitFile
Quién lee: humans
Lifecycle: logs/ (gitignored); ~661 files observed
Schema: `{ts} [LEVEL] msg`
Estado: IMPLEMENTED_UNVERIFIED
```

### 3) Journal JSONL

```text
Qué almacena: setup_evaluated, signal_*, order_*, position_closed
Quién escribe: journal.FileWriter from Loop or backtest.Run
Quién lee: analytics_extractor.py, humans
Lifecycle: journals/ or backtesting/**/jsonl; gitignored trees
Schema: internal/journal/events.go
Estado: PARTIAL (opt-in; live close incomplete)
```

### 4) Candle JSON / CSV

```text
Qué almacena: historical OHLC for replay
Quién escribe: runBacktestDownload / SaveCandlesToFile; Capital.com
Quién lee: LoadCandlesFromFile, backtest.Run
Lifecycle: backtesting/**, lab/; gitignored except lab/
Schema: Capital prices array or internal candle JSON
Estado: IMPLEMENTED_UNVERIFIED
```

`lab/candles.json` is **tracked** (sample). `lab/candles.toon` tracked.

### 5) Backtest CSV / HTML / PNG

```text
Qué almacena: closed trades, leaderboards, charts
Quién escribe: backtest.Run + scripts/backtest_runner.py
Quién lee: research / firebase_rtdb.py
Lifecycle: backtesting/ gitignored
Estado: IMPLEMENTED_UNVERIFIED (local only)
```

### 6) Redis (optional)

```text
Qué almacena: notification dedup keys (POSITION_OPENED/CLOSED by deal_ref)
Quién escribe: notifications.RedisGlobalStore
Quién lee: same
Lifecycle: TTL days (default 7)
Schema: opaque keys
Estado: IMPLEMENTED_UNVERIFIED
```

Not used for market data or orders.

### 7) Firebase Realtime Database (optional)

```text
Qué almacena: live instance meta/status/positions; backtest analytics (Python)
Quién escribe: internal/telemetry (uncommitted), scripts/firebase_rtdb.py
Quién lee: planned dashboard (not in repo)
Lifecycle: cloud; credentials via service-account.json (local, gitignored)
Schema: docs/RTDB_SCHEMA_V1.md (local)
Estado: IMPLEMENTED_UNVERIFIED
```

### 8) Not present

PostgreSQL, SQLite, MongoDB, ClickHouse, time-series DB, object store — **no client code**.

## Data quality controls

`config.DataQuality` (defaults in `config.go`):

- max age M15 20m, H1 75m, H4 300m
- reopen warmup 3 M15 candles
- `block_on_data_frozen` default true

These skip **signal evaluation**; they do not persist a quality ledger.

## Lifecycle risk

Most operational history lives **outside git**. A clone of `aurumflow` does **not** include bots, backtests, logs, or scripts. The tracked product is the engine only.
