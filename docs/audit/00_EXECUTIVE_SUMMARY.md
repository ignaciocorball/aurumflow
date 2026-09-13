# 00 — Executive Summary

**Project (git module):** `aurumflow`  
**Workspace folder:** `institutional-trading-bot`  
**Remote:** `https://github.com/ignaciocorball/aurumflow.git`  
**Audit version:** 1.0  
**Audit date:** 2026-09-12  
**Branch audited:** `feature/claude-code` (HEAD `4250f7b`) plus uncommitted working tree  
**Sister branch:** `main` — **identical commit** to `feature/claude-code`  
**P0 freeze:** `P0 FREEZE: READY WITH WARNINGS`

---

## What this repository is today

This is **AurumFlow**: a single-process **Go 1.24** trading bot that:

1. authenticates to the **Capital.com REST API**;
2. pulls **OHLC candles** (M5 / M15 / H1 / H4);
3. runs a **Liquidity → Structure → Intent → Equilibrium** scoring pipeline;
4. gates orders with a **state machine + risk manager**;
5. **opens CFD positions** via `POST /api/v1/positions` with SL/TP levels;
6. optionally journals events, notifies via Pushover/Telegram, and (uncommitted) publishes live status to Firebase RTDB.

It is **not** an institutional microstructure radar. There is no order book, no Market-by-Order, no options flow, no CME/Nasdaq adapter, no dashboard binary in git, and no futures feed.

The product vision in README.md is explicit: automated **XAUUSD (gold) CFD** trading with institutional-*style candle logic*. The local working tree later generalized that same engine to **US100, US500, ETHUSD, MSFT, SILVER, NDAQ, US30, MNQ1** via epic override and gitignored `bots/` + `backtesting/` corpora.

---

## Direct answers (28 questions)

| # | Question | Answer | Status |
|---|----------|--------|--------|
| 1 | What is this project today? | AurumFlow: Capital.com CFD bot + candle SMC-style strategy + file backtester | `VERIFIED_WORKING` (as a codebase) |
| 2 | Real architecture? | Monolith: `cmd/bot` → `core.Loop` → market / strategy / risk / execution | `VERIFIED_WORKING` |
| 3 | Languages / frameworks? | Go 1.24.3 (`aurumflow`). Local-only Python research scripts. No Docker/K8s/UI in git | `VERIFIED_WORKING` |
| 4 | Strategies? | One composer pipeline (not a strategy plugin framework). See `04_STRATEGY_INVENTORY.md` | `IMPLEMENTED_UNVERIFIED` live |
| 5 | Instruments? | Any Capital.com epic via `AURUMFLOW_EPIC`. Default resolve = gold COMMODITIES. Local bots: GOLD/XAUUSD, US100, US500, ETHUSD, MSFT | `PARTIAL` |
| 6 | Data sources? | Capital.com REST OHLC only. Local JSON/CSV candle caches. No other vendors in code | `VERIFIED_WORKING` (code) |
| 7 | Capital.com integrated? | Yes — REST client with CST / X-SECURITY-TOKEN | `IMPLEMENTED_UNVERIFIED` runtime |
| 8 | Capital.com DEMO works? | Code supports `api.mode=demo`. **All local configs found are LIVE.** Runtime DEMO not exercised this audit | `UNKNOWN` |
| 9 | Can we query market data? | `GET /api/v1/prices/{epic}` implemented and used by loop + backtest download | `IMPLEMENTED_UNVERIFIED` |
| 10 | Can we send demo orders? | `Executor.SendOrder` posts positions. DEMO untested. LIVE would send if `AURUMFLOW_LIVE_CONFIRM=1` | `IMPLEMENTED_UNVERIFIED` |
| 11 | Position management? | Read open positions. Confirm deal. **No close / update / trailing API** | `PARTIAL` |
| 12 | Safe demo/live isolation? | Host remap + `AURUMFLOW_LIVE_CONFIRM` exist. **Example and all local configs default LIVE.** Empty `mode` + live URL bypasses confirm | `PARTIAL` / `BROKEN` (bypass) |
| 13 | Real risk management? | Risk % size, max trades, daily DD. No kill switch, no correlation, no circuit breaker that halts the process | `PARTIAL` |
| 14 | Backtesting? | Candle replay on M15 (+ optional H1/H4). No tick/MBO replay, no walk-forward, no Monte Carlo | `IMPLEMENTED_UNVERIFIED` (stats not trusted) |
| 15 | Are backtests reliable? | No. Bid-only OHLC, spread/slippage default 0, same-bar SL-first heuristic, no commissions, short windows, CFDs ≠ futures | `PARTIAL` |
| 16 | Tests that pass? | 66 unit tests, `go test ./...` PASS, `go vet` PASS, `go build ./cmd/bot` PASS | `VERIFIED_WORKING` |
| 17 | Tests that fail? | None in this baseline | `VERIFIED_WORKING` |
| 18 | Dead modules? | `InFibZone*`, `PhaseExpansion`, several notification types never emitted; `UsePartial` / `BlockH4Range` unused | `DEAD_CODE` |
| 19 | Technical debt? | See `14_TECHNICAL_DEBT.md`. Largest: live-default, no execution modes, gitignored operational corpus | `VERIFIED_WORKING` (as findings) |
| 20 | Exposed secrets? | Not in git index. **Local `config.json`, all `bots/*/config.json`, `service-account.json` contain credentials.** `/docs` was being ignored | `PARTIAL` |
| 21 | Reusable for Radar? | Capital adapter, risk, sessions, journal, state machine, backtest harness, notifications | See `12_REUSE_MATRIX.md` |
| 22 | What to refactor? | Broker interface, demo/live hard-gate, instrument contract, value-per-point, account selection | P1 |
| 23 | What to build from scratch? | Order flow, book, options, cross-asset, CME/Nasdaq, feature store, radar score, dashboard, replay | P2–P6 |
| 24 | How much Radar exists? | **~14%** — candle SMC + CFD execution only | See `13_INSTITUTIONAL_RADAR_GAP_ANALYSIS.md` |
| 25 | Minimum path to DEMO bot? | Dedicated demo config, read-only session proof, close/update, kill switch, journal on, never default live | P1 |
| 26 | Path from there to Radar? | Split market-data vs execution, add institutional feeds, score, human console, automation last | P2–P7 |
| 27 | Next checkpoint after P0? | **P1 — Execution Foundation (DEMO-only)** | See `16_ROADMAP.md` |

---

## Architecture in one paragraph

```text
Capital.com REST
    → market.Client (session, prices, accounts, positions)
    → core.Loop (60s poll)
        → strategy.BuildComposerInput + SignalComposer
        → core.StateMachine (READY gate)
        → risk.Manager
        → execution.Executor.SendOrder
    → Capital.com POST /api/v1/positions
```

No event bus. No separate market-data service. No UI process. Backtest is an offline fork of the same scoring functions.

---

## Capital.com (code vs runtime)

| Capability | Code | Runtime this audit |
|------------|------|--------------------|
| Authentication (CST + X-SECURITY-TOKEN) | Implemented | **Not executed** (all local configs are LIVE) |
| Market data OHLC | Implemented | Not executed |
| Streaming / WebSocket / Lightstreamer | `streamingHost` parsed, never used | N/A |
| List / switch accounts | Implemented | Not executed |
| Balance | Implemented (bug: loop uses first account) | Not executed |
| Create order + SL/TP | Implemented | Not executed |
| Read positions | Implemented | Not executed |
| Close / update position | **Not implemented** | N/A |
| Demo/live isolation | Partial, bypassable | Local state is LIVE |

---

## Critical blockers for a trustworthy next phase

1. **P0 CRITICAL — Live default + populated LIVE credentials locally.** `config/example_config.json` (committed) and every inspected local config use `api.mode=live`. This audit refused to call the API.
2. **P0 CRITICAL — LIVE_CONFIRM bypass.** Confirm is required only when `mode == "live"`. Empty mode + live `api_base_url` can send orders without the env flag (`cmd/bot/main.go`, `config.Validate`).
3. **P1 HIGH — No position close/update.** The bot can open; it cannot flatten or modify. Close is inferred when `openCount == 0`.
4. **P1 HIGH — No `EXECUTION_MODE`.** There is no DISABLED / DRY_RUN / DEMO / LIVE enum. Demo is only a URL switch.
5. **P1 HIGH — Operational corpus is gitignored.** `bots/`, `scripts/`, `backtesting/`, most `docs/` are local-only. Git does not contain the research platform that actually exists on disk.

---

## Institutional Radar readiness

**14%.**

Justification: we can already produce candle-derived structure, liquidity *zones* (wicks/equals), session labels, a discrete bot state, and a numeric strategy score. We cannot observe aggressive flow, book imbalance, absorption, depletion/replenishment, icebergs, CVD, options/GEX, or true futures identity (NQ/GC vs US100/GOLD CFDs). Those layers are `NOT_IMPLEMENTED`.

Institutional Pressure Score inputs available today: **~18%** (levels + session + H4/H1 regime + confidence score). See `13_INSTITUTIONAL_RADAR_GAP_ANALYSIS.md`.

---

## P0 freeze recommendation

Treat the **git-tracked Go bot** as frozen and documented.  
Treat the **local gitignored corpus** as research evidence, not as a second undocumented product.

**Do not start P1 implementation in this iteration.**  
Next human decision: approve P1 Execution Foundation with a **DEMO-only** config that this machine does not currently have.

---

## Documentation set

All files live under `docs/audit/`. Manifest: `docs/audit/MANIFEST.json`.
