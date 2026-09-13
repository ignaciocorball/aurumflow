# P8.1 Eligibility and portfolio risk

## Single authority

`ExecutionEligibilityRegistry` is the only writer of execution eligibility.

Canonical states (no jumps):

```
ANALYSIS_ONLY
  → discover unique Capital identity
DEMO_DISCOVERED
  → prepare (read-only monetary spec)
DEMO_SPEC_VALID
  → explicit calibrate command
DEMO_CALIBRATED
  → tradeable + demo host + known monetary risk
DEMO_ELIGIBLE

any → BLOCKED
```

`DEMO_NOT_CALIBRATED` is not emitted. LIVE maps to `BLOCKED` / `LIVE_PROHIBITED` reason.

MarketState, DecisionProposal, and UI all read the registry after overlay.

## Portfolio risk units

Caps are **account_currency_risk**: expected stop loss in the DEMO account currency.

- max 3 open positions
- max 1 correlated group
- 300 aggregate account-currency risk
- 150 per region
- 150 per asset class

Unknown monetary risk still blocks. Blockers are a sorted semantic set (no duplicates).
