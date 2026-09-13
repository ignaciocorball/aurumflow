# 08 — Backtesting and Research

## Git-tracked engine

```text
Package: internal/backtest
Entry:   backtest.Run(candlesM15, candlesH1, candlesH4, cfg, initialBalance, journal, csvPath)
CLI:     --backtest <file>
         --backtest-from / --backtest-to / --backtest-epic / --backtest-dir / --backtest-force
Status:  IMPLEMENTED_UNVERIFIED
Tests:   backtest_test.go — only asserts TotalTrades>=0 and MaxDrawdownPct>=0 on synthetic data
```

## What it can do

| Capability | Exists | Status | Evidence |
|------------|--------|--------|----------|
| Historical candle replay (M15) | Yes | `IMPLEMENTED_UNVERIFIED` | `Run` loop from bar 50 |
| H1 / H4 aligned by timestamp | Yes | `PARTIAL` | index walk; H4 used as score |
| Tick replay | No | `NOT_IMPLEMENTED` | |
| Event / MBO replay | No | `NOT_IMPLEMENTED` | |
| Strategy backtest | Yes | `PARTIAL` | Same composer, **reimplemented** gates vs live |
| Optimization / grid | Local Python runner only | `IMPLEMENTED_UNVERIFIED` | `scripts/backtest_runner.py` profiles A–Z (gitignored) |
| Walk-forward | No | `NOT_IMPLEMENTED` | ripgrep no matches |
| Out-of-sample protocol | No formal | `NOT_IMPLEMENTED` | date windows exist as files, no locked OOS procedure in Go |
| Monte Carlo | No | `NOT_IMPLEMENTED` | |
| Slippage | Field | `PARTIAL` | `SlippagePoints`; live CLI sets **0** |
| Commissions | No | `NOT_IMPLEMENTED` | |
| Spreads | Field | `PARTIAL` | `SpreadPoints`; live CLI sets **0** |
| Latency | No | `NOT_IMPLEMENTED` | |
| Break-even sim | Yes | `PARTIAL` | `UseBreakEven` inside `Run` |
| Same-bar SL/TP | Conservative SL-first | `IMPLEMENTED_UNVERIFIED` | `resolveOutcome` |
| Metrics | Yes | `IMPLEMENTED_UNVERIFIED` | trades, winrate, PF (capped 99), max DD %, expectancy |
| Journal / CSV | Yes | `IMPLEMENTED_UNVERIFIED` | jsonl + trades CSV |

## Download path

`runBacktestDownload` logs into Capital.com, pulls 7-day chunks (max 1000) for M15/H1/H4, writes:

```text
backtesting/{EPIC}_M15_{from}_{to}.json
backtesting/{EPIC}_H1_...
backtesting/{EPIC}_H4_...
```

Source tag in saver: `"capital.com"`.

**This path authenticates.** Do not point it at LIVE configs for research if a demo host is available.

## Local research corpus (gitignored)

On disk under `backtesting/`:

| Folder | Meaning (from names / bot READMEs) |
|--------|--------------------------------------|
| gold | GOLD CFD, many dated profile runs |
| us100 | US100 CFD; bot README cites PF 2.92 / 6 trades (tiny N) |
| us500 | US500 CFD; README cites 68 trades PF 1.71 |
| us30 | Dow CFD |
| ethusd | ETH CFD; defensive PF 1.97 / 69; offensive PF 4.06 / 35 |
| msft | MSFT stock CFD |
| silver | SILVER CFD |
| ndaq | NDAQ stock CFD |
| mnq1 | Unknown epic naming — **not a CME feed** |

Artifacts per run (typical): `config.json`, `report.html`, `leaderboard.csv`, `runs.csv`, equity/drawdown PNGs, jsonl, csv.

`scripts/backtest_runner.py` orchestrates profile grids (`A_baseline`, `B_selective`, … `Z_current_institutional`), ranking, Highcharts HTML. Default date constants in that file include `2025-10-01` → `2026-02-05`.

`scripts/analytics_extractor.py` + `firebase_rtdb.py` publish aggregates to RTDB `/v1/backtests`.

## Are results statistically reliable?

**No. Do not treat any local PF/winrate as edge.**

Reasons with evidence:

1. **Bid-only candles** — ask/spread not in OHLC (`prices.go`).
2. **Costs zeroed** in `main.go` `runBacktest` (`SpreadPoints: 0`, `SlippagePoints: 0`).
3. **No commissions.**
4. **OHLC path dependency** — same-bar SL-first is a heuristic, not ticks.
5. **Gold ATR buckets hardcoded** in composer (8–18) applied to indices.
6. **Short samples** — US100 selective README: **6 trades**.
7. **No walk-forward / OOS lock / multiple-testing correction.**
8. **CFD ≠ futures** — cannot claim NQ/GC performance.
9. **Live/backtest drift** — H4 readiness (60 bars) in live; backtest uses ≥10 H4 bars for trend. M5 refiner is live-only.
10. Unit test does **not** validate profitability or determinism of a golden journal.

Status of “strategy is profitable”: `UNKNOWN` (and must stay unknown).

## Reuse verdict

`KEEP_AND_REFACTOR` the Go replay harness.  
`WRAP` the Python runner as research tooling (it is not in git today).  
`REWRITE` any claim layer before using numbers for automation decisions.
