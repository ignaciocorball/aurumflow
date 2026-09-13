# P8.2 — Unified intelligence runtime

P8.1 docs remain `72`–`76`. This file starts the P8.2 series.

## Command

```text
go run ./cmd/bot --intelligence-runtime --status-addr 127.0.0.1:8766
```

Optional bounded soak: `--soak-minutes 120`.

`--soak-minutes 1` (flag default) is treated as unbounded until SIGINT (24h cap). GOLD DEMO remains `:8765` via `.\scripts\sunday_go.ps1`.

## Composition

One process. No `ExecutionProvider`. No broker mutation.

```text
Official WorldState → Capital surface → OpportunitySurface → MarketScanner
        → Legacy (GOLD only when history READY)
        → Cross-asset live frame
        → BTC micro (aggTrades + incremental L2 + book + Radar + Exhaustion + Absorption)
        → DecisionOrchestrator → SHADOW proposals
```

`--intelligence-runtime` replaces the need to run `--shadow-runtime` and `--global-shadow` separately for research/observability.

Canonical instances (not duplicated): WorldState rematerialization, Binance/L2 collector and book engine, opportunity prospective recorder.

## BTC MicroAvailable

`true` only when all six are healthy: aggTrades, incremental L2, book synced, Radar, Exhaustion, Absorption.

Binance down → BTC micro unavailable. Official / Capital analysis continues.

Capital feed down → live Opportunity components unavailable. Slow WorldState remains valid. Stale frames are not presented as live (`lastFrameOK` cleared after 3 minutes without quotes).

## Ports

| Process | Port |
|---|---|
| GOLD Legacy DEMO | 127.0.0.1:8765 |
| Unified intelligence | 127.0.0.1:8766 |

Intelligence may **read** GOLD DEMO `/api/gold`. It cannot mutate process A.

## Safety

LIVE fail-closed. `MULTI_MARKET_DEMO_EXECUTION = OFF`. `CanMutateBroker() = false`. No mass calibration.
