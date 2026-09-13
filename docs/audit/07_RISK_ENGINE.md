# 07 — Risk Engine

## Implementation

```text
Package: internal/risk
Type:    Manager
Status:  PARTIAL
Tests:   none
Evidence: internal/risk/manager.go
```

Constructed in `core.NewLoop` from `cfg.Risk.{RiskPerTrade,MaxTrades,DailyDrawdownLimit}` and startup balance.

## Control matrix

| Risk control | Existe | Estado | Evidencia |
|--------------|--------|--------|-----------|
| Risk per trade | Yes | `IMPLEMENTED_UNVERIFIED` | `PositionSize`: `riskAmount = balance * (riskPercent/100)` then `/ (stopDist * valuePerPoint)` |
| Max open positions | Yes | `PARTIAL` | `CanOpenTrade`: `OpenCount >= MaxTrades`. Count is **per epic**, other instruments ignored |
| Daily drawdown limit | Yes | `PARTIAL` | `(DailyStartBalance - Balance) / DailyStartBalance * 100 < limit`. Reset UTC midnight. **Balance may be wrong account** (loop takes first `GetAccounts` item) |
| Max drawdown (equity, all-time) | No | `NOT_IMPLEMENTED` | Only daily |
| Leverage cap | No | `NOT_IMPLEMENTED` | No leverage read/write |
| Correlated positions | No | `NOT_IMPLEMENTED` | Multi-bot local setup can open GOLD+US100+ETH independently |
| Stop loss | Yes on entry | `PARTIAL` | Signal SL sent; not managed after |
| Take profit | Yes on entry | `PARTIAL` | Same |
| Trailing | No | `NOT_IMPLEMENTED` | |
| Break-even | Backtest only | `PARTIAL` | `UseBreakEven` in `backtest.Run`; live loop never updates SL |
| Partials / runner | Config only | `STUB` | `UsePartial`, `PartialR`, `RunnerR` unused |
| Circuit breaker | Notify only | `PARTIAL` | Heartbeat can emit `TRADING_HALTED` when DD ≥ limit; **loop still runs**; next trade blocked only via `CanOpenTrade` |
| Kill switch | No | `NOT_IMPLEMENTED` | No env/file/API to flatten and halt |
| Session limits | Session allow-list + overlap block | `IMPLEMENTED_UNVERIFIED` | Strategy layer, not risk package |
| Spread cap | Optional | `PARTIAL` | `api.max_spread`; example and local root config = 0 (disabled) |
| Live confirm | Yes | `PARTIAL` | Env flag; bypass if mode empty |
| Min size floor | Yes | `PARTIAL` | If computed size < min, **size is raised to min** — can exceed intended risk |
| Data quality halt | Yes | `IMPLEMENTED_UNVERIFIED` | Stale candles / non-TRADEABLE / warmup skip evaluation |
| Account isolation | Weak | `BROKEN` / `PARTIAL` | No check that switched account is DEMO; live default |

## Position sizing formula

Evidence: `risk.PositionSize`.

```text
size = floor( (balance * risk%/100) / (|entry-sl| * valuePerPoint) / step ) * step
if size < minSize → size = minSize
```

`valuePerPoint` default **1.0**. For US100/US500 this is unlikely to be correct unless the operator set it per bot. Bot READMEs do not document a calibrated `value_per_point`.

## What “trading halted” actually means

`logStatusSummary` emits `TypeTradingHalted` when `ddPct >= DailyDrawdownLimit`. That is a **notification**, not a process halt. `ValidateSignal` will reject new entries while DD remains ≥ limit (strictly `dd < limit` to open). Existing positions are not closed.

## Reuse verdict

`KEEP_AND_REFACTOR` — the three numeric gates are real and simple. They are not an institutional risk engine. P1 should add: correct account selection, kill switch, flatten, per-instrument contract specs, and a hard execution disable.
