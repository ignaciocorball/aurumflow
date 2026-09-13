# 14 — Technical Debt

Each item: severity, impact, location, recommendation. No rewrite in P0.

## P0 CRITICAL

| ID | Issue | Impact | Location | Recommendation |
|----|-------|--------|----------|----------------|
| D-001 | Committed example defaults to LIVE | Operators copy-paste into real money | `config/example_config.json` | Change default to `demo` in P1 |
| D-002 | LIVE_CONFIRM only if `mode==live` | Empty mode + live URL sends orders | `cmd/bot/main.go`, `config.Validate` | Require confirm whenever host is live |
| D-003 | All local configs LIVE + secrets | Accidental live auth | `config/config.json`, `bots/*/config.json` | Demo-only config; env-only secrets |
| D-004 | Loop balance = first account | Wrong risk size / DD | `internal/core/loop.go` GetAccounts loop | Select `CurrentAccountID` |

## P1 HIGH

| ID | Issue | Impact | Location | Recommendation |
|----|-------|--------|----------|----------------|
| D-010 | No close/update position | Cannot flatten or BE live | `internal/market/positions.go`, `execution` | Implement DELETE/PUT wrappers |
| D-011 | No EXECUTION_MODE | No dry-run / disable | config + loop | Add enum; default DISABLED or DEMO |
| D-012 | No kill switch | Cannot halt+flatten | — | File/env/Firebase command |
| D-013 | `value_per_point` default 1.0 | Mis-sized US100/ETH/MSFT | `config.Validate`, `risk.PositionSize` | Per-epic contract table |
| D-014 | Min size raises risk | Small accounts over-risk | `risk.PositionSize` | Reject if size==min and risk exceeded |
| D-015 | No market/execution/risk tests | Regressions | those packages | P1 test pack with httptest mocks |
| D-016 | Journal off by default | Cannot explain trades | example + local config | Default on for DEMO |
| D-017 | Live close journal UNKNOWN | No PnL forensics | `loop.go` PositionClosed | Poll deal history / confirm |
| D-018 | Scripts/bots/backtesting gitignored | Clone ≠ working system | `.gitignore` | Decide what is product vs lab |
| D-019 | `Loop` god-object | Untestable tick | `loop.go` ~1300 lines | Extract gates in P1/P2 |

## P2 MEDIUM

| ID | Issue | Impact | Location | Recommendation |
|----|-------|--------|----------|----------------|
| D-020 | Bid-only candles | Optimistic/pessimistic bias | `prices.go` | Store bid/ask/mid |
| D-021 | Live vs backtest drift | False confidence | loop vs `backtest.Run` | Share one evaluator |
| D-022 | H4 comment ≠ code | Operators think hard filter | config comments, loop | Align or rename to score |
| D-023 | `BlockH4Range` unused | Dead config | config + backtest struct | Use or remove in a later cleanup (do not delete now) |
| D-024 | ATR buckets gold-specific | Bad scores on indices | `composer.go` | Per-instrument buckets |
| D-025 | Hardcoded executor 0.1 lot | Ignores dealing rules | `NewLoop` | Use GetMarketDetails |
| D-026 | StreamingHost unused | False “realtime” expectation | `session.go` | Document; implement later or ignore |
| D-027 | Confirm fail still IN_TRADE | State lie | `loop.go` | Reconcile with GetPositions |
| D-028 | No Docker/CI | Baseline not portable | — | Optional later; not P0 |
| D-029 | Firebase deps on dirty tree | Heavy module graph | `go.mod` uncommitted | Isolate telemetry module |

## P3 LOW

| ID | Issue | Impact | Location | Recommendation |
|----|-------|--------|----------|----------------|
| D-040 | Dead `InFibZone*` | Confusion | `fibonacci.go` | Leave until strategy pass |
| D-041 | Unused notif types | Taxonomy lie | `events.go` | Emit or mark unused |
| D-042 | `UsePartial` stub | Operators enable with no effect | `config.go` | Document stub |
| D-043 | `PhaseExpansion` unused | False regime claim | `structure.go` | Leave named, unused |
| D-044 | Chile local time logs | Cosmetic | `main.go`, `loop.go` | Keep; not debt |
| D-045 | README workingorders | Docs lie | `README.md` | Fix when touching docs |
| D-046 | Bot READMEs say “no credentials in file” | Files **do** contain them | `bots/*/README.md` vs JSON | Rotate + env-only |

## Architectural debt (not a rewrite ticket)

The product grew from a gold bot into a **multi-CFD research lab** without an instrument contract, without a data/execution split, and without demo-first defaults. That is the core debt. Radar work before P1 would multiply it.
