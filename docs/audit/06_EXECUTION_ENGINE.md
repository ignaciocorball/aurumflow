# 06 — Execution Engine

## What exists

```text
Package: internal/execution
Type:    Executor
Status:  IMPLEMENTED_UNVERIFIED
Tests:   none
```

`Executor` is a thin Capital.com order sender, not a generic execution adapter.

```text
SendOrder(ctx, signal, size)
  → POST /api/v1/positions
  → { epic, direction, size, stopLevel, profitLevel, guaranteedStop:false }

ConfirmDeal(ctx, dealReference)
  → GET /api/v1/confirms/{dealReference}
```

Wired in `core.NewLoop` with hardcoded `minSize=0.1`, `sizeStep=0.1` regardless of `GetMarketDetails`.

## Order path (live process)

Evidence: `internal/core/loop.go` after signal generation.

```text
signal
 → state == READY
 → M5 timing (optional recycle)
 → LiveConfirmRequired? skip
 → Risk.ValidateSignal
 → spread vs max_spread (if > 0)
 → risk.PositionSize(balance, risk%, stopDist, min, step, valuePerPoint)
 → journal OrderAttempt + notif ORDER_SENT
 → Exec.SendOrder
 → on error: COOLDOWN + ORDER_REJECTED
 → on success: IN_TRADE + POSITION_OPENED + ConfirmDeal (best-effort)
```

## Execution modes (desired vs actual)

| Desired | Exists? | Actual |
|---------|---------|--------|
| `EXECUTION_MODE=DISABLED` | No | Process always intends to trade when READY |
| `DRY_RUN` | No | — |
| `DEMO` | Partial | `api.mode=demo` changes host only |
| `LIVE` | Partial | `api.mode=live` + `AURUMFLOW_LIVE_CONFIRM` |

Status: `NOT_IMPLEMENTED` as a mode enum. `PARTIAL` as host + confirm flag.

## Position lifecycle

| Action | Implemented | Status |
|--------|-------------|--------|
| Open | POST positions | `IMPLEMENTED_UNVERIFIED` |
| Confirm fill | GET confirms | `IMPLEMENTED_UNVERIFIED` |
| Read open | GET positions | `IMPLEMENTED_UNVERIFIED` |
| Filter by epic | Loop | `IMPLEMENTED_UNVERIFIED` |
| Close | — | `NOT_IMPLEMENTED` |
| Partial close | config `use_partial` | `STUB` |
| Modify SL/TP | — | `NOT_IMPLEMENTED` |
| Break-even | backtest only | `PARTIAL` |
| Working orders | README only | `NOT_IMPLEMENTED` |
| Guaranteed stop | always false | `STUB` |

When the last position disappears, journal writes `exit_reason=UNKNOWN`, `exit_price=0`. Decision reconstruction of *why it closed* is **not possible** from live journal.

## Safety behavior this audit

- Did **not** run the bot.
- Did **not** call Capital.com.
- Reason: every local `config.json` is LIVE with credentials present.
- `AURUMFLOW_LIVE_CONFIRM` is unset, so a careless `go run` would still **authenticate LIVE** and would skip orders only while `mode=="live"`.

## Rate limits (documented vs code)

| Rule | README | Code |
|------|--------|------|
| Session 10 min / ping 5 min | Yes | Ping every 5 min |
| Max 10 req/s | Yes | Client limiter 5/s |
| 1 POST / 100ms | Yes | Executor mutex + 100ms |
| Demo 1000 orders/hour | Yes | Not enforced locally |

## Bugs / debt touching execution

1. Confirm failure does not roll back `IN_TRADE`.
2. `ORDER_FILLED` notification type never emitted (confirm success is journal-only).
3. Size not clamped to `maxDealSize` from market details.
4. `value_per_point` default 1.0 — wrong size on index CFDs.
5. No idempotency key beyond broker `dealReference`.
6. No disable switch short of killing the process or failing READY/risk.

## Reuse verdict

`WRAP` — keep Capital POST/confirm; hide behind `ExecutionProvider` in P1/P2. Do not generalize until DEMO close/update exist.
