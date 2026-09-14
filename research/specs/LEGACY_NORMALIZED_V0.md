# LEGACY_NORMALIZED_V0

Spec version: `LEGACY_NORMALIZED_V0`  
Purpose: research/shadow only. Preserve GOLD Legacy **hypothesis** while removing GOLD-specific dimensional assumptions.

GOLD execution Legacy is **not** replaced by this spec.

This file is frozen before performance results. Do not change formulas after seeing holdout numbers.

## Hypothesis (unchanged)

1. **Structure** — directional market structure (BULLISH/BEARISH) must agree with the setup.
2. **Liquidity** — a sweep / liquidity event against the prior pool precedes the trade.
3. **Intent** — composer score ≥ 5 from the existing Legacy feature set (sweep, structure alignment, equilibrium).
4. **Equilibrium** — equilibrium touch remains a score bonus, not a standalone entry.

## What is removed

- Absolute ATR point gates (`MinATR`/`MaxATR` in price units).
- Hard-coded GOLD ATR buckets `8–12` / `12–18`.
- A single London+NY session filter applied to every market.

## Normalization (dimensionless, own-history, past-only)

At signal time `t0`, using bars with `bar_end <= t0` only:

1. `atr_pct = ATR(14) / close`
2. `atr_pct_rank` = percentile of `atr_pct` among the prior 200 M15 `atr_pct` values (fewer if history is shorter; minimum 40).
3. Volatility suitability: `0.20 <= atr_pct_rank <= 0.80`.
4. Stop distance: `1.0 * ATR` (same R construction as Legacy ATR fallback).
5. Target distance: `1.5 * ATR` minimum RR (same sanity gate as Legacy).
6. Fib SL/TP remain ATR-relative (`0.2 ATR` / `0.5 ATR`) when Fib is valid.

No per-market ATR, RSI, or stop search.

## Sessions (objective hours, not fitted)

| Markets | Allowed sessions |
|---|---|
| GOLD, SILVER, OIL_CRUDE | ALL |
| US100, US500, US30 | NY |
| DE40, UK100 | LONDON |
| J225, CN50 | ASIA |
| BTC | not an execution market in this spec |

Session labels use the existing `sessions` package. Weekend/closed gaps are not synthesized.

## Cost methodology

Every simulated fill is `SHADOW_NET_RETURN`, not guaranteed executable PnL.

Spread is applied at entry and exit (half-spread each side unless a live Capital spread is known, in which case that value is used).

Slippage scenarios (price units of the instrument, not GOLD points):

| Name | Extra slippage |
|---|---|
| ZERO | 0 |
| NORMAL | 0.5 × current/typical spread |
| STRESS | 2.0 × current/typical spread |

Primary promotion metric uses **NORMAL**.

## Evaluation horizons

Signal `t0` is immutable. Forward path uses prices with `t > t0`.

Trade simulation: first of stop, target, or 96 M15 bars (~24h). Weekend gaps → `PARTIAL_SESSION` / skip fill if the next bar opens through the stop beyond STRESS slippage.

## Chronological splits (frozen)

On each dataset hash, split by **bar time** (not shuffled):

- DISCOVERY: first 60% of `[first_bar, last_bar]`
- VALIDATION: next 20%
- HOLDOUT: final 20%

Do not tune after holdout. Dataset checksum must be recorded on every result row.

## Metrics

Signals, trades, long, short, mean/median net return (R), hit rate, expectancy (R), profit factor, MFE, MAE, max DD (R), holding time, turnover, cost drag (ZERO vs NORMAL vs STRESS).

R = signed price change to exit / ATR at `t0`, after costs.

## Promotion (not a score)

`SHADOW_VALIDATED` requires **all**:

- holdout n ≥ 20 trades
- holdout expectancy > 0 after NORMAL costs
- profit factor > 1.1
- max DD < 20R
- at least 2 of 3 chronological thirds of holdout have expectancy ≥ 0
- largest winner < 40% of total holdout R
- broker mechanics feasible
- data quality not INVALID

Discovery and Attention/Salience must not promote.

## Execution influence

NONE. GOLD `:8765` Legacy is unchanged. DEMO_MIRROR may consume this spec only after a later explicit promotion gate.
