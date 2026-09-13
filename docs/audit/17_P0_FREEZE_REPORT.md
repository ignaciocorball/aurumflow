# 17 — P0 Freeze Report

Equivalent to a Patagonia Shield-style freeze: **what is true, what is not, what we will not pretend.**

---

## P0 STATUS

```text
P0 FREEZE: READY WITH WARNINGS
```

Discovery is complete enough to decide P1. The freeze is **not** blocked by missing source. It **is** warned because:

- Capital.com DEMO was **not** runtime-verified (all local configs LIVE; API not called).
- Live exposure on this workstation is **HIGH**.
- Operational research corpus (`bots/`, `backtesting/`, `scripts/`) is **outside git**.
- Execution cannot close positions.

---

## Confirmed working

| Item | Evidence |
|------|----------|
| Go module builds | `go build ./cmd/bot` exit 0 |
| Unit tests | `go test ./...` 66 pass, 0 fail |
| `go vet` | clean |
| Strategy units | structure, swings, sessions, RSI, ATR, composer SL/TP sanity, state TTL, H1 filter, notifications |
| Config load/validate (code) | `config.Load` + `Validate` exist |
| Session/ping/accounts/prices/positions **code** | `internal/market` |
| Order send **code** | `Executor.SendOrder` |
| State machine READY gate | `state.go` + tests |
| Session hours | `sessions.go` + tests |
| Backtest function runs | `backtest.Run` + smoke test |

---

## Confirmed broken or unsafe

| Item | Evidence |
|------|----------|
| LIVE_CONFIRM bypass if `mode` empty | `main.go` `liveConfirmRequired := cfg.API.Mode == ModeLive && ...` |
| Example + local configs default LIVE | `example_config.json`; all inspected JSON `mode=live` |
| Loop risk uses first account balance | `loop.go` GetAccounts `break` on first |
| Cannot close/modify positions | no DELETE/PUT |
| Live close forensics | `exit_reason=UNKNOWN`, `exit_price=0` |
| `BlockH4Range` / `UsePartial` advertised but unused | config vs loop/backtest |
| Bot READMEs vs files | READMEs say no secrets; files have keys |
| Min size can exceed intended risk | `PositionSize` floors up to min |

---

## Unverified

| Item | Why |
|------|-----|
| Capital.com DEMO login | No demo config; not called |
| Capital.com LIVE login | Refused |
| Market data freshness in production | Not run |
| Order fill quality / slippage live | Not run |
| Multi-bot LIVE operation | Local folders exist; not started |
| Firebase publish | Code uncommitted; not run |
| Python runner | `python` not on PATH |
| MNQ1 folder meaning | Name only |
| Statistical edge of any profile | Backtests untrustworthy by construction |

---

## Security blockers

1. Populated LIVE credentials in root and all seven `bots/*` configs.
2. `service-account.json` on disk.
3. Committed template teaches `mode: live`.
4. Confirm flag does not protect empty-mode live hosts.
5. Logs print account ids and balances.

**No secrets were committed at HEAD.** Local disk is the exposure.

---

## Demo trading readiness

```text
NOT READY
```

Code path for demo host exists. This machine has **zero** demo configs. There is no integration test. Close API missing. Journal off in current configs.

---

## Live trading exposure

```text
HIGH
```

A default `go run ./cmd/bot` with `config/config.json` would authenticate to **live** Capital.com. Orders would be skipped only while `mode` remains `live` and `AURUMFLOW_LIVE_CONFIRM` is unset. Relogin, price fetch, and account listing would still hit LIVE.

This audit performed **no** live calls.

---

## Tests

```text
PASS: 66
FAIL: 0
BLOCKED: Capital.com integration, Python research scripts
PACKAGES WITHOUT TESTS: market, execution, risk, config, journal, cmd/bot
```

---

## Architecture

Frozen as: **single-binary Capital.com CFD bot** with a candle SMC composer, 60s poll loop, optional notifications/journal, optional (dirty-tree) Firebase telemetry.

Not frozen as: Institutional Trading Radar, futures platform, or multi-broker gateway.

`main` and `feature/claude-code` share commit `4250f7b`. Dirty tree adds telemetry.

---

## Technical debt

Top freeze-relevant: D-001…D-004, D-010…D-018. Full list: `14_TECHNICAL_DEBT.md`.

---

## Reusable assets

- Capital REST client (WRAP)
- Composer + structure + sessions + indicators (KEEP)
- Risk numeric gates (KEEP_AND_REFACTOR)
- State machine (KEEP_AND_REFACTOR)
- Journal schema (KEEP)
- Notifications (KEEP_AND_REFACTOR)
- OHLC backtest harness (KEEP_AND_REFACTOR)
- Local research corpus (KEEP, do not delete)
- RTDB schema draft (KEEP for P6)

---

## Critical gaps vs Institutional Radar

No order flow, book, options, CME/Nasdaq, event bus, feature store, pressure score, or dashboard app. CFD ≠ NQ/GC. Score inputs ~18%. Platform ~14%.

---

## Next phase recommendation

**P1 — Execution Foundation (DEMO-only).**

Concrete first checkpoint after freeze:

1. Commit/keep this `docs/audit/**` set as the freeze record.
2. Do **not** implement Radar.
3. When approved: demo config, host hard-gate, httptest Capital client, close position API, `EXECUTION_MODE`, journal on, fix account balance, kill switch.
4. Prove DEMO read-only session before any order.

---

## Freeze declaration

The git-tracked AurumFlow engine and the local research overlay have been photographed.

No strategies, datasets, or experiments were deleted.

P1 has not been started.
