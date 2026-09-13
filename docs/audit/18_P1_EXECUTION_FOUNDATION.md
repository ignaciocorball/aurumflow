# 18 — P1 Execution Foundation

P0 remains the historical freeze (`17_P0_FREEZE_REPORT.md`). This document records what P1 changed.

## Objective

Make AurumFlow a **DEMO-only**, fail-closed execution platform that can observe, size, open, confirm, update, and close a Capital.com position without any accidental LIVE path.

## What shipped

| Capability | Status | Evidence |
|------------|--------|----------|
| `APIEnvironment` vs `ExecutionMode` | Done | `config/safety.go` |
| LIVE fail-closed (config + HTTP) | Done | `ApplyP1BrokerLock`, `Client.assertNotLive` |
| Operational host pinned to DEMO | Done | `DemoAPIHost`; custom/live URL refused |
| `DISABLED` / `DRY_RUN` / `DEMO` | Done | `ParseExecutionMode`; loop + `DryRunProvider` |
| LIVE as parseable mode | Done | Always `ErrLiveDisabled` |
| httptest Capital client suite | Done | `internal/market/client_test.go` |
| `ExecutionProvider` | Done | `internal/execution/provider.go` |
| Close / Update / GetPosition | Done | `market.ClosePosition`, `UpdatePosition`; `Executor.*` |
| Deal lifecycle journal | Done | `journal.Lifecycle` + loop writes |
| Account selection | Done | `risk.SelectAccount` — no `accounts[0]` |
| Min-size reject | Done | `risk.ComputeSize` → `MIN_SIZE_EXCEEDS_RISK_BUDGET` |
| InstrumentSpec | Done | `market.SpecFromDetails`; incomplete → no trade |
| Kill switch | Done | `internal/killswitch` (config/env/file) |
| Flatten DEMO CLI | Done | `--flatten` + `--flatten-confirm` |
| Journal default on | Done | `Validate` unless `AURUMFLOW_JOURNAL=0` |
| Startup banner | Done | `cmd/bot/banner.go` |
| Read-only DEMO canary CLI | Done | `--canary-readonly` + `AURUMFLOW_DEMO_*` |
| example_config | Done | demo + disabled + journal on |

## What P1 did not do

- No strategy changes (composer/H1/H4/sessions).
- No backtest research rewrite.
- No LIVE trading, CME, options, Radar, UI.
- No DEMO runtime canary (credentials missing — see `19_P1_DEMO_VALIDATION.md`).

## Operator model

```text
API ENVIRONMENT = demo          (only accepted value)
EXECUTION MODE  = DISABLED | DRY_RUN | DEMO

DISABLED : market + signals + journal; no order intent mutation
DRY_RUN  : full sizing + would_have_sent journal; no POST/PUT/DELETE
DEMO     : broker mutations against DEMO host only
```

Historical LIVE configs on disk are **not migrated**. Loading them fails even if `AURUMFLOW_API_ENVIRONMENT=demo` is set:

```text
STARTUP REFUSED: LIVE broker environment is disabled in P1.
```

## CLI

```text
go run ./cmd/bot
go run ./cmd/bot --canary-readonly
go run ./cmd/bot --flatten --flatten-confirm
go run ./cmd/bot --backtest path.json   # LoadOffline, no broker
```
