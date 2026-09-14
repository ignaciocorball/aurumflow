# MARKET_SALIENCE_V0

Spec version: `MARKET_SALIENCE_V0`  
Spec hash input: this file's frozen formula (not AttentionScore, not Legacy).  
Purpose: research-only diagnostic of **unusual current market behavior vs each market's own recent past**.

This is **not** Opportunity, Attention, Buy, or a trade score.  
Legacy / risk / session / AttentionScore V1 must not consume it.

## Complementary axes

- **Attention (ATTENTION_SCORE_V1):** where structural/institutional evidence suggests looking.
- **Salience (this spec):** where price/vol behavior is unusually active **now**.

Do not fuse them into a trade score in P8.8.

## Inputs (past-only, no lookahead)

All features use bars with `bar_end <= t0`.

1. `abs_ret_z` — |close(t0)-close(t0-lookback)| / own realized vol over the lookback (default 12 M15 bars ≈ 3h).
2. `rv_expand` — current RV / median RV of the prior 24 M15 bars (capped).
3. `range_expand` — current M15 range / median M15 range of prior 24 bars.
4. `rs_disp` — |relative strength vs cross-section median return| / cross-section RV.
5. `xs_pct` — percentile of |return| in the current cross-section of TRADEABLE markets (0–1).
6. `dq_pen` — 0 if DataQuality HEALTHY; 0.35 if STALE; 1.0 if DEGRADED/UNKNOWN/CLOSED.

No macro. No flow inference. No strategy features.

## Normalization

Cross-asset prices are never compared in points. All terms are dimensionless.

Each of (1)–(5) is clipped to [0, 3] then scaled to [0, 1] via `min(x, 3)/3`.

## Frozen weights (not optimized)

| Term | Weight |
|---|---|
| abs_ret_z | 0.30 |
| rv_expand | 0.20 |
| range_expand | 0.15 |
| rs_disp | 0.15 |
| xs_pct | 0.20 |

`raw = Σ w_i * scaled_i`  
`Salience = 100 * raw * (1 - dq_pen)`  
Clipped to [0, 100].

Missing bars → that term is 0 (not fabricated). Fewer than 8 lookback bars → Salience 0 and reason `INSUFFICIENT_HISTORY`.

## Output

```
score          0–100
label          MARKET SALIENCE
components     map of scaled terms
reason         short
```

## Research pairing

At the same Opportunity snapshot t0 store Attention and Salience.  
Later label absolute return, MFE, MAE, realized vol after t0. Do not rewrite t0.
