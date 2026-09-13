# 02 — Current Architecture

Reconstructed from code. Components listed only if they exist.

## Real topology

This is a **single-binary monolith**, not a layered platform with replaceable providers.

```text
                    ┌─────────────────────────┐
                    │  cmd/bot/main.go        │
                    │  flags: --backtest*     │
                    └───────────┬─────────────┘
                                │
              ┌─────────────────┴──────────────────┐
              │                                    │
     LIVE/DEMO PATH                         BACKTEST PATH
              │                                    │
              ▼                                    ▼
     config.Load()                          LoadCandlesFromFile
     market.NewClient                       (optional Capital
     CreateSession                           download first)
     ResolveEpic / AURUMFLOW_EPIC
              │                                    │
              ▼                                    ▼
        core.NewLoop                         backtest.Run
        core.Loop.Run  (60s)                 (in-process, no HTTP)
              │                                    │
              ▼                                    ▼
     ┌────────────────────────────────────────────┐
     │ Shared intelligence                        │
     │  structure.*  strategy.*  indicators.*     │
     │  marketstate.TrendFromCandles              │
     │  core.StateMachine                         │
     └────────────────────────────────────────────┘
              │
              ▼
     risk.Manager  →  execution.Executor  →  Capital.com
```

## Requested pipeline vs reality

| Requested layer | Exists? | Actual component | Status |
|-----------------|---------|------------------|--------|
| DATA SOURCE | Yes | Capital.com REST only | `IMPLEMENTED_UNVERIFIED` |
| INGESTION | Yes | `Client.GetPrices` / `GetMarketDetails` / `GetPositions` | `IMPLEMENTED_UNVERIFIED` |
| NORMALIZATION | Partial | Bid OHLC → `models.Candle`. Volume = `lastTradedVolume`. No event schema | `PARTIAL` |
| ANALYSIS | Yes | Swings, structure, liquidity zones, sweep, fib, RSI, ATR, sessions | `IMPLEMENTED_UNVERIFIED` |
| STRATEGY | Yes | Single `SignalComposerWithZones` | `IMPLEMENTED_UNVERIFIED` |
| RISK | Yes | `risk.Manager` | `PARTIAL` |
| EXECUTION | Yes | `Executor.SendOrder` | `IMPLEMENTED_UNVERIFIED` |
| BROKER | Yes | Capital.com REST | `IMPLEMENTED_UNVERIFIED` |
| Event bus | No | — | `NOT_IMPLEMENTED` |
| Feature store | No | — | `NOT_IMPLEMENTED` |
| UI / dashboard process | No | RTDB schema only (local docs) | `NOT_IMPLEMENTED` |

## Requested separations

| Concern | Separated? | Evidence |
|---------|------------|----------|
| Market data | No interface | `*market.Client` concrete in `Loop` |
| Strategy engine | Functions, not plugins | `internal/strategy` |
| Signal engine | Same package as strategy | `SignalComposerWithZones` |
| Execution engine | Separate package | `internal/execution` — still hard-wired to Capital client |
| Risk engine | Separate package | `internal/risk` |
| Portfolio | No | Single epic per process; other epics ignored when counting |
| Persistence | Files | logs + optional JSONL journal + optional RTDB |
| Reporting | Notifications + journal + local HTML reports | No in-repo UI |
| Backtesting | Separate package | Shares strategy functions, reimplements gating |

## Runtime objects (`core.Loop`)

Evidence: `internal/core/loop.go` `type Loop struct`.

Wired in `NewLoop`:

- `Config *config.Config`
- `Client *market.Client`
- `Epic string`
- `Risk *risk.Manager`
- `Exec *execution.Executor`
- `State *StateMachine`
- optional `Journal`, `NotifEmitter`, `Telemetry`, `ReloginFn`

There is **one epic per process**. Multi-instrument “platform” is achieved by running multiple OS processes (`bots/*`).

## State machine (execution gate)

Evidence: `internal/core/state.go`.

```text
IDLE → SEEK_LIQUIDITY → WAIT_SWEEP → WAIT_PULLBACK → READY → (SendOrder) IN_TRADE → COOLDOWN → IDLE
```

`CanSendOrder()` is true **only** in `READY`.

TTL (default 12 M15 candles) and opposite-structure invalidation reset to IDLE.

These states are **bot operational states**, not institutional regime states. Mapping to future ACCUMULATION/DISTRIBUTION/… is incomplete; see section 17 in `13_INSTITUTIONAL_RADAR_GAP_ANALYSIS.md`.

## Live tick (what actually happens every 60s)

Evidence: `Loop.tick` in `internal/core/loop.go`.

1. Optional heartbeat / status log.
2. Market status gate (`TRADEABLE` vs cache).
3. Session filter `strategy.CanTrade`.
4. Ping session every 5 minutes; relogin on 401.
5. Refresh accounts + positions (balance bug: first account).
6. Fetch M15; optionally H1, H4, M5.
7. Freshness / reopen warmup.
8. `BuildComposerInput` → state `Transition`.
9. `SignalComposerWithZones` + diagnostics.
10. London+NY overlap hard skip if configured.
11. H4/session **score modifiers** (not always hard reject).
12. Optional H1 hard filter.
13. State READY check.
14. Optional M5 timing recycle.
15. `LiveConfirmRequired` skip.
16. `Risk.ValidateSignal`.
17. Optional spread check (`max_spread`, default 0 = off).
18. `PositionSize` → `SendOrder` → `ConfirmDeal`.

## Backtest path

Evidence: `cmd/bot/main.go` `runBacktest` / `runBacktestDownload`, `internal/backtest/backtest.go`.

- Offline: JSON/CSV candles, no broker.
- Online download: same Capital session, chunked 7-day `GetPrices`.
- Gating is a **reimplementation**, not a call into `Loop.tick`.
- Spread/slippage fields exist; live backtest invocation sets them to **0**.

## Observability attached to the loop

| Sink | Package | When |
|------|---------|------|
| Stdout / file logs | `internal/logger` | Always (file if `log_dir`) |
| JSONL journal | `internal/journal` | If `logging.journal_enabled` |
| Pushover / Telegram | `internal/notifications` | If enabled + credentials |
| Redis dedup | `notifications/redisdedup.go` | If `global_dedup.type=redis` |
| Firebase RTDB | `internal/telemetry` (uncommitted) | If `telemetry.firebase.enabled` |

## What this architecture is *not*

- Not a market-data gateway with multiple adapters.
- Not a strategy framework (no `Strategy` interface).
- Not a portfolio engine.
- Not a microservice mesh.
- Not an Institutional Radar.

Those absences are documented, not proposed as a rewrite in P0.
