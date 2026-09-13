# 54 — Sunday Operations

GOLD DEMO execution remains Legacy-only. BTC intelligence is SHADOW research and must not feed trade decisions.

## Startup order

```text
1. DEMO environment verification
   --ops-preflight
   Expect READY or READY_WITH_WARNINGS. BLOCKED stops the day.

2. Shadow runtime BTC
   go run ./cmd/bot --shadow-runtime --l2-provider auto --soak-minutes 1440 --status-addr 127.0.0.1:8766

3. Inspect console
   http://127.0.0.1:8766/
   Confirm LIVE IMPOSSIBLE, SHADOW badges, book synced, deltas increasing.

4. Prepare GOLD
   go run ./cmd/bot --prepare-instrument GOLD

5. Calibrate GOLD when TRADEABLE
   go run ./cmd/bot --calibrate-instrument GOLD

6. Confirm
   GOLD RUNTIME_VALIDATED
   open positions = 0

7. Start demo-week
   go run ./cmd/bot --demo-week --status-addr 127.0.0.1:8765

8. Keep BTC intelligence SHADOW running on 8766
```

Optional supervisor (does **not** trade, calibrate, or bypass gates):

```text
go run ./cmd/bot --sunday-runtime --l2-provider auto
```

## Preflight

`--ops-preflight` is read-only. Result: `READY` | `READY_WITH_WARNINGS` | `BLOCKED`.

## Collision

If 8765/8766 are taken, pass `--status-addr`. Do not bind the console on a public interface.
