# P8.2 — DEMO operational trust

Operational trust ≠ profitability validation.

## Gates (all eight required)

1. Broker identity confirmed
2. Stable market data
3. History warmup
4. Strategy compatibility
5. Monetary calibration
6. Risk computation
7. Position lifecycle tested
8. Reconciliation tested

Calibration alone does **not** grant `DEMO_OPERATIONALLY_TRUSTED`.

## Current labels

| Market | Trust |
|---|---|
| GOLD | NOT_TRUSTED until Sunday calibrate **and** demo-week lifecycle evidence |
| others | DEMO_DISCOVERED or DEMO_SPEC_VALID |

## Calibration queue

States: `READY_TO_CALIBRATE` / `NOT_TRADEABLE` / `SPEC_INVALID` / `IDENTITY_UNRESOLVED` / `RISK_UNKNOWN`.

Even if 8 markets are READY: **do not calibrate**. Queue only.

Recommended next (explainable, not profit): evaluate SILVER / US100 / OIL / US500 after GOLD. Order is computed from open + identity + spec + Legacy compat + coverage + diversification.

## Portfolio

Unknown monetary risk → BLOCK. Caps are `account_currency_risk`: Aggregate $300 DEMO, region/asset $150 DEMO (USD account). Groups stay static; 2h correlation is diagnostic only.

`DEMO_CANDIDATE` still does not execute: `MULTI_MARKET_DEMO_EXECUTION = OFF`.
