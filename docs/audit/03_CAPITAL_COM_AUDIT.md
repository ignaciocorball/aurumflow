# 03 — Capital.com Audit

**Rule:** no LIVE API calls were made. All local configs inspected are `mode=live`. Runtime cells below are therefore `UNKNOWN` unless a unit test exists (none do for `internal/market`).

---

## Client surface

| Item | Evidence |
|------|----------|
| Package | `internal/market` |
| Type | `market.Client` |
| Auth headers | `X-CAP-API-KEY`, `CST`, `X-SECURITY-TOKEN` (`client.go` `Do`) |
| Hosts | Demo `https://demo-api-capital.backend-capital.com` / Live `https://api-capital.backend-capital.com` (`config.Validate`) |
| Rate limit | 5 req/s, burst 2; 429 retry with backoff (`client.go`) |
| Timeout | 30s HTTP client |
| Tests | **None** (`go test` reports `? aurumflow/internal/market [no test files]`) |

README cites [Capital.com Public API](https://open-api.capital.com/) including WebSocket. **WebSocket is not implemented.**

---

## Capability matrix

### Authentication

```text
Capital.com authentication
Status: IMPLEMENTED_UNVERIFIED
Evidence:
  internal/market/session.go
  Client.CreateSession()  POST /api/v1/session
  captures CST + X-SECURITY-TOKEN response headers
  cmd/bot/main.go calls CreateSession at startup and on 401
Tests: none
Runtime this audit: NOT EXECUTED (LIVE configs present; refused)
```

### Market data

```text
Status: IMPLEMENTED_UNVERIFIED
Evidence:
  internal/market/prices.go
  Client.GetPrices()  GET /api/v1/prices/{epic}?resolution=&max=&from=&to=
  Maps bid OHLC → models.Candle; Volume = lastTradedVolume
  Resolutions: MINUTE, MINUTE_5, MINUTE_15, MINUTE_30, HOUR, HOUR_4, DAY, WEEK
Used by: Loop.tick, main startup fetch, runBacktestDownload
Runtime: NOT EXECUTED
```

### Streaming

```text
Status: NOT_IMPLEMENTED (field PARTIAL)
Evidence:
  session.go SessionResponse.StreamingHost json:"streamingHost,omitempty"
  No WebSocket client, no Lightstreamer, no subscribe loop
  StreamingHost is never read after unmarshal
```

### Accounts / demo account

```text
Status: IMPLEMENTED_UNVERIFIED
Evidence:
  Client.GetAccounts()  GET /api/v1/accounts
  Client.GetSession()   GET /api/v1/session
  Client.SwitchAccount() PUT /api/v1/session {accountId}
  main.go lists accounts and optionally switches cfg.API.AccountID
Demo vs live accounts: isolation is by API host, not by account-type check in code
Runtime: NOT EXECUTED
```

### Balance / margin

```text
Status: PARTIAL
Evidence:
  AccountInfo.Balance.{Balance,Available,Deposit,ProfitLoss}
  Logged at startup from session.Accounts matching CurrentAccountID
BUG: Loop.tick GetAccounts assigns the FIRST account in the slice
     (internal/core/loop.go refresh-balance loop with break on first item)
No dedicated margin/leverage endpoint wrapper
Runtime: NOT EXECUTED
```

### Orders

```text
Status: IMPLEMENTED_UNVERIFIED
Evidence:
  internal/execution/executor.go
  Executor.SendOrder() POST /api/v1/positions
  Body: epic, direction, size, stopLevel, profitLevel, guaranteedStop=false
  100ms spacing between POSTs
  README mentions POST /workingorders — NOT implemented
Tests: none
Runtime: NOT EXECUTED (would be LIVE if invoked with current configs)
```

### Positions (read)

```text
Status: IMPLEMENTED_UNVERIFIED
Evidence:
  internal/market/positions.go
  Client.GetPositions() GET /api/v1/positions
  PositionItem.GetEpic() from market.epic or position.epic
  Loop filters by Loop.Epic
Runtime: NOT EXECUTED
```

### Close

```text
Status: NOT_IMPLEMENTED
Evidence:
  No DELETE /api/v1/positions/{dealId}
  No close helper in execution or market
  Loop treats close as: StateInTrade && openCount==0 → COOLDOWN
  PositionClosed journal exit_reason="UNKNOWN", exit_price=0
```

### SL / TP

```text
Status: PARTIAL
Evidence:
  SendOrder sets stopLevel + profitLevel from signal
  No PUT to modify SL/TP after fill
  UseBreakEven implemented in backtest only
  UsePartial config never wired
Cannot assign/update SL/TP on an existing position from code
```

### Confirm

```text
Status: IMPLEMENTED_UNVERIFIED
Evidence:
  Executor.ConfirmDeal() GET /api/v1/confirms/{dealReference}
  Called after send; fill/slippage written to journal if present
  Confirm failure is WARN only — state already IN_TRADE
```

### Risk gate before execute

```text
Status: PARTIAL
Evidence:
  Loop: LiveConfirmRequired → Risk.ValidateSignal → optional spread → PositionSize → SendOrder
  ValidateSignal: max open trades + daily DD vs DailyStartBalance
  No kill-switch file/env that stops the process
  No dry-run path
```

### Demo / live isolation (CRITICAL)

```text
Status: PARTIAL with BROKEN bypass
Evidence:
  config.Validate:
    mode=demo → force demo host if URL looks like capital.com without "demo"
    mode=live + URL contains "demo-api" → error
    mode empty → any api_base_url accepted
  cmd/bot/main.go:
    liveConfirmRequired := cfg.API.Mode == config.ModeLive && env != 1/true
    If mode is empty and URL is live, orders are allowed without AURUMFLOW_LIVE_CONFIRM
  example_config.json (COMMITTED): "mode": "live"
  Local configs inspected (NOT printed): ALL mode=live, live host, credentials present
    config/config.json
    config/bk_config.json
    bots/{xauusd,us100,us500,ethusd,eth-defensive-core,eth-offensive-alpha,msft}/config.json
```

**This audit did not authenticate, did not fetch prices, and did not send orders.**

---

## Endpoints implemented

| Method | Path | Function | Status |
|--------|------|----------|--------|
| POST | `/api/v1/session` | `CreateSession` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/session` | `GetSession` | `IMPLEMENTED_UNVERIFIED` |
| PUT | `/api/v1/session` | `SwitchAccount` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/ping` | `Ping` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/accounts` | `GetAccounts` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/markets?searchTerm=` | `ResolveEpic` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/markets/{epic}` | `GetMarketDetails` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/prices/{epic}` | `GetPrices` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/positions` | `GetPositions` | `IMPLEMENTED_UNVERIFIED` |
| POST | `/api/v1/positions` | `Executor.SendOrder` | `IMPLEMENTED_UNVERIFIED` |
| GET | `/api/v1/confirms/{dealRef}` | `ConfirmDeal` | `IMPLEMENTED_UNVERIFIED` |
| DELETE | `/api/v1/positions/{id}` | — | `NOT_IMPLEMENTED` |
| PUT | `/api/v1/positions/{id}` | — | `NOT_IMPLEMENTED` |
| POST | `/api/v1/workingorders` | — | `NOT_IMPLEMENTED` (mentioned in README) |
| WS / Lightstreamer | — | — | `NOT_IMPLEMENTED` |

---

## Instrument resolution

`ResolveEpic(ctx, "gold")` searches markets and **prefers `InstrumentType == "COMMODITIES"`**, else first hit.

`AURUMFLOW_EPIC` / `--backtest-epic` / `scripts/run.ps1 -Epic` bypass search and use the raw epic string (e.g. `GOLD`, `US100`, `ETHUSD`).

These are **Capital.com CFD epics**, not CME symbols. See `05_MARKET_DATA_SOURCES.md`.

---

## Demo trading readiness (this machine)

| Check | Result |
|-------|--------|
| Demo host in code | Yes |
| Local demo config | **None found** |
| Env `AURUMFLOW_LIVE_CONFIRM` | Unset |
| Safe to run `go run ./cmd/bot` with default config | **No — would login LIVE** (orders blocked only if mode stays `live` and confirm unset) |
| Integration test | Missing |

---

## Recommended proof sequence (P1, not executed)

1. Create a **new** demo-only config (do not copy live).
2. Read-only: `CreateSession` + `GetAccounts` + `GetPrices` + `GetPositions`.
3. Assert host contains `demo-api`.
4. Only then consider a single minimum-size DEMO order in a later checkpoint.
