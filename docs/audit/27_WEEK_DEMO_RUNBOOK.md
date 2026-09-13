# 27 — One-Week DEMO Runbook

Capital.com remains DEMO-only. LIVE is impossible.

## Commands

Start Sunday (after GOLD is TRADEABLE and calibrated):

```text
go run ./cmd/bot --prepare-instrument GOLD
go run ./cmd/bot --calibrate-instrument GOLD
go run ./cmd/bot --demo-week --epic GOLD --status-addr 127.0.0.1:8765
```

Inspect status:

```text
go run ./cmd/bot --status --status-addr 127.0.0.1:8765
```

Also: `GET http://127.0.0.1:8765/healthz` and `GET http://127.0.0.1:8765/status`

Halt new orders:

```text
go run ./cmd/bot --halt-new-orders
```

or create `.aurumflow.kill` or set `AURUMFLOW_KILL_SWITCH=1`

Flatten DEMO if necessary (requires `AURUMFLOW_DEMO_*`, never LIVE config):

```text
go run ./cmd/bot --flatten --flatten-confirm --epic GOLD
```

Infrastructure canary (TRADEABLE instrument):

```text
go run ./cmd/bot --canary-lifecycle
go run ./cmd/bot --canary-lifecycle --epic BTCUSD
```

SHADOW soak:

```text
go run ./cmd/bot --shadow-soak --soak-minutes 60
```

## Environment

```text
AURUMFLOW_DEMO_API_KEY
AURUMFLOW_DEMO_IDENTIFIER
AURUMFLOW_DEMO_PASSWORD
AURUMFLOW_CONFIG=config/demo_config.json
```

Do not set LIVE `account_id`. Do not load `config/config.json` for operational DEMO commands.

## `--demo-week` sequence

1. Safety boot — DEMO credentials only
2. DEMO host proof
3. DEMO account selection
4. Kill switch
5. GOLD market state
6. GOLD InstrumentSpec + monetary validation
7. Strategy prerequisites (existing Composer unchanged)
8. Journal health
9. Restart recovery: `GetPositions` classified MANAGED / RECOVERED / UNKNOWN
10. Enable DEMO execution only if all gates pass; otherwise analysis continues and new orders stay blocked

Refuse trades when:

```text
LIVE detected
wrong host
wrong account
GOLD CLOSED
InstrumentSpec invalid
monetary spec not RUNTIME_VALIDATED
journal failure
daily DD blocked
kill switch active
stale market data
UNKNOWN positions
```

## Recovery

Never assume local state after restart. Unknown positions block new orders until resolved. Existing positions remain observable; flatten is DEMO-only.

## Daily report

`internal/report` reads JSONL journals for signals, trades, pnl, rejections, radar states, sessions, kill-switch and recovery events. One week is not statistical edge.
