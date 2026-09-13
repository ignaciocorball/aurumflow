# 28 — Sunday Readiness

## Overall

The platform is ready to **begin** a one-week Capital.com DEMO evaluation when GOLD becomes TRADEABLE.

It is not ready to fire GOLD strategy orders until `--calibrate-instrument GOLD` produces `RUNTIME_VALIDATED` monetary semantics.

```text
P1  READY  (DEMO mutation proven on BTCUSD; final positions 0)
P2  READY  (dual plane, identity, bus, isolation)
P3  READY as SHADOW MVP  (public BTC feed + radar; soak practical duration)
GOLD strategy execution  BLOCKED until TRADEABLE + RUNTIME_VALIDATED
LIVE  FAIL CLOSED
Radar AUTO_EXECUTE  not implemented
```

## Sunday order of operations

1. Confirm DEMO env vars only
2. `go run ./cmd/bot --prepare-instrument GOLD`
3. If TRADEABLE: `go run ./cmd/bot --calibrate-instrument GOLD`
4. Confirm `journals/instruments/GOLD.json` validation=`RUNTIME_VALIDATED`
5. Confirm open positions = 0
6. `go run ./cmd/bot --demo-week --epic GOLD`
7. Keep Radar SHADOW (`--shadow-soak` optional, never mutates)
8. If uncertain: `--halt-new-orders` then inspect `--status`

## What this week is

A runtime/evidence test of the existing Composer + DEMO execution + SHADOW radar.

Not: strategy optimization, institutional certainty, CME identity, or LIVE quality.
