# 12 — Reuse Matrix (Institutional Radar)

Classification of **existing** components only.

| Component | Verdict | Why | Radar fit |
|-----------|---------|-----|-----------|
| `internal/market.Client` (REST + session) | **WRAP** | Works as Capital CFD adapter; not a generic MD gateway | Capital.com Adapter |
| `GetPrices` / candle mapping | **KEEP_AND_REFACTOR** | Bid-only; needs mid/ask/contract metadata | Historical + poll feed |
| `ResolveEpic` | **REWRITE** | Gold/COMMODITIES heuristic; unsafe for multi-asset | Instrument registry |
| `internal/execution.Executor` | **WRAP** | Keep POST+confirm; add close/update; interface | Execution Adapter |
| `internal/risk.Manager` | **KEEP_AND_REFACTOR** | Real but thin; fix account + min-size-exceeds-risk | Risk Engine seed |
| `internal/core.StateMachine` | **KEEP_AND_REFACTOR** | Good execution gate; do not overload as regime model | Execution / setup FSM |
| `internal/core.Loop` | **REWRITE** | God-object tick; mix MD, strategy, risk, IO | Split later; freeze now |
| `internal/strategy` composer | **KEEP** | Valid candle SMC research engine | Session/regime *inputs*, not order-flow |
| Liquidity / sweep / eq / fib | **KEEP** | Research value; not book liquidity | Feature candidates (candle) |
| `internal/structure` | **KEEP** | Simple, tested | Structure features |
| `internal/indicators` | **KEEP** | Tested RSI/ATR | Filters |
| `internal/strategy/sessions.go` | **KEEP** | Clear UTC sessions | Session Engine seed |
| `internal/marketstate` | **KEEP** | H1/H4/M5 trend helper | Weak Regime Engine seed |
| `internal/backtest` | **KEEP_AND_REFACTOR** | Replay harness worth keeping | Research / later event replay is new |
| `internal/journal` | **KEEP** | Right idea for decision audit | Alert/explain evidence |
| `internal/logger` | **KEEP** | Fine for a bot | |
| `internal/notifications` | **KEEP_AND_REFACTOR** | Solid ops; unused event types | Alert Engine seed (human) |
| `internal/telemetry` (uncommitted) | **WRAP** | Live status bus to RTDB | Temporary dashboard feed |
| `config` | **KEEP_AND_REFACTOR** | Rich; live-default and secret-in-file must change | |
| `pkg/models` | **KEEP_AND_REFACTOR** | Candle/signal types; not institutional events | Event Normalizer is new |
| `cmd/bot/main.go` | **KEEP_AND_REFACTOR** | Useful flags; too much orchestration | |
| `bots/*` (local) | **KEEP** | Ops pattern: one process per epic | Instance model |
| `scripts/backtest_runner.py` (local) | **KEEP** | Research grid | P5 research |
| `scripts/analytics_extractor.py` | **KEEP** | Segment analytics | Research |
| Firebase backtest schema (local docs) | **KEEP** | Contract for later UI | P6 input |
| React dashboard | **UNKNOWN** / **NOT_IMPLEMENTED** | Schema only | Build in P6 |
| README live example | **DEPRECATE** default live | Educational but dangerous | |
| `InFibZone*`, unused notif types, `UsePartial` | **DEPRECATE** or finish | Dead / stub | |
| MQL5 / MT5 ideas | N/A | Not in repo | |

## Preserve especially

Requested preserve list → verdict:

| Theme | Verdict |
|-------|---------|
| Broker execution | WRAP Capital; do not delete |
| Strategy framework | KEEP functions; there is no plugin framework |
| Risk | KEEP_AND_REFACTOR |
| Sessions | KEEP |
| Instrument abstraction | **Does not exist** — must be built (P2) |
| Historical data | KEEP file candles + download |
| Backtesting | KEEP_AND_REFACTOR |
| Visualization | Local HTML/PNG + RTDB schema; no app |
| Logging / journal | KEEP |

## Do not discard

Per P0 rule: nothing was deleted. Local backtests, bots, and scripts stay. Classify, don’t erase.
