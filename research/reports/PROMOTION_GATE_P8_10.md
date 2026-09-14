# PROMOTION GATE P8.10

Frozen strategy: `LEGACY_NORMALIZED_V0` (`LEGACY_NORMALIZED_V0` / `aurumflow/internal/legacynorm`).

Promotion uses holdout after NORMAL costs plus the STRESS catastrophic-reversal gate. Attention and Salience cannot promote or create orders.

| Market | Research status | Holdout n | Net expectancy | PF | MaxDD | Cost stress | Subperiod robustness | Broker spec | Monetary | Live shadow | DEMO status |
|---|---|---:|---:|---:|---:|---|---|---|---|---|---|
| GOLD | SHADOW_VALIDATED | 20 | 0.568 | 2.53 | 4.23 | 0.512 (OK) | 3/3 | SPEC_READY | RUNTIME_VALIDATED | RUNNING_8765 | GOLD_STRATEGY |
| SILVER | RESEARCH_READY | 6 | -1.269 | 0.00 | 7.61 | -1.537 (FAIL) | 0/3 | SPEC_READY | — | SHADOW | NO |
| OIL_CRUDE | RESEARCH_READY | 5 | -1.137 | 0.00 | 5.68 | -1.273 (FAIL) | 0/3 | SPEC_READY | — | SHADOW | NO |
| US100 | DEMO_ELIGIBLE | 86 | 0.152 | 1.28 | 11.72 | 0.084 (OK) | 3/3 | SPEC_READY | RUNTIME_VALIDATED | ARMED | DEMO_ELIGIBLE |
| US500 | RESEARCH_REJECTED | 40 | -0.089 | 0.86 | 8.65 | -0.198 (OK) | 1/3 | SPEC_READY | — | SHADOW | NO |
| US30 | RESEARCH_REJECTED | 58 | -0.062 | 0.90 | 13.66 | -0.116 (OK) | 1/3 | SPEC_READY | — | SHADOW | NO |
| DE40 | RESEARCH_REJECTED | 33 | -0.084 | 0.87 | 7.91 | -0.229 (OK) | 1/3 | SPEC_READY | — | SHADOW | NO |
| UK100 | RESEARCH_REJECTED | 54 | -0.387 | 0.53 | 25.78 | -0.746 (FAIL) | 0/3 | SPEC_READY | — | SHADOW | NO |
| J225 | RESEARCH_REJECTED | 54 | -0.346 | 0.55 | 20.51 | -0.432 (OK) | 0/3 | SPEC_READY | — | SHADOW | NO |
| CN50 | RESEARCH_REJECTED | 34 | -0.348 | 0.55 | 15.54 | -1.019 (FAIL) | 0/3 | SPEC_READY | — | SHADOW | NO |

## Hard gates

- holdout n ≥ 20
- net holdout expectancy > 0 after NORMAL modeled costs
- no catastrophic reversal under STRESS costs
- profit factor > 1.1 (frozen research contract; also > 1)
- max DD < 20R
- ≥ 2 of 3 chronological holdout thirds non-negative
- largest winner < 40% of total holdout R
- data quality VALID
- exact broker identity confirmed
- broker spec usable
- market tradeable as a CFD on the explicit DEMO account
- monetary RUNTIME_VALIDATED with complete evidence
- explicit DEMO account VERIFIED; LIVE rejected

## GOLD exact evidence

- Market: GOLD
- Dataset hash: `0aaa92986ddc4911f3ebd56615194757e48d0dbb40ff68c80d2199f9379d3676`
- Holdout dates: 2026-08-12 → 2026-09-10
- Holdout n: 20
- Net expectancy after NORMAL costs: 0.568
- Net expectancy under STRESS costs: 0.512
- Hit rate: 0.650
- Profit factor: 2.53
- Max drawdown (R): 4.23
- MFE: 1.026
- MAE: -0.575
- Long expectancy: 0.611
- Short expectancy: 0.550
- Positive chronological subperiods: 3 / 3
- Largest-trade contribution: 0.128
- Data-quality status: VALID
- Broker identity: capital.com
- Broker spec status: SPEC_READY
- Monetary status: RUNTIME_VALIDATED

## US100 exact evidence

- Market: US100
- Dataset hash: `28484e65ee51d242df7615aa9624d21cf7c8834ec9b9d68333ba622407582e1d`
- Holdout dates: 2026-08-09 → 2026-09-10
- Holdout n: 86
- Net expectancy after NORMAL costs: 0.152
- Net expectancy under STRESS costs: 0.084
- Hit rate: 0.488
- Profit factor: 1.28
- Max drawdown (R): 11.72
- MFE: 0.858
- MAE: -0.657
- Long expectancy: 0.057
- Short expectancy: 0.291
- Positive chronological subperiods: 3 / 3
- Largest-trade contribution: 0.111
- Data-quality status: VALID
- Broker identity: capital.com
- Broker spec status: SPEC_READY
- Monetary status: RUNTIME_VALIDATED

## US100 monetary

- Previous: RUNTIME_VALIDATED
- Current: RUNTIME_VALIDATED
- MPU: 1.000000
- Currency: USD
- Samples: 80
- Price range: 5.6000
- UPL range: 0.0058
- Estimator: THEIL_SEN
- Dispersion (slope MAD): 0.160000
- Metadata comparison: METADATA_RUNTIME_AGREE
- Position closed: YES
- Broker positions after: 0

## DEMO account

- Environment: DEMO
- Account: VERIFIED
- Masked: `30**************30`
- LIVE usable: NO

## US100 promotion decision

DEMO_ELIGIBLE — DEMO_MIRROR_US100=ON
