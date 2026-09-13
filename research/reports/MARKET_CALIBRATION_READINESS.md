# MARKET CALIBRATION READINESS

MULTI_MARKET_DEMO_EXECUTION = OFF. Queue only. No automatic calibration.

## States

READY_TO_CALIBRATE · NOT_TRADEABLE · SPEC_INVALID · IDENTITY_UNRESOLVED · RISK_UNKNOWN

## Queue at P8.2 engineering (weekend / preopen)

| Market | State | Why |
|---|---|---|
| GOLD | NOT_TRADEABLE or RISK_UNKNOWN | Sunday `sunday_go.ps1` remains the only calibrate+demo path |
| SILVER | NOT_TRADEABLE | identity may be discovered; monetary UNKNOWN |
| OIL_CRUDE | NOT_TRADEABLE | monetary UNKNOWN |
| US100 | NOT_TRADEABLE | |
| US500 | IDENTITY_UNRESOLVED or NOT_TRADEABLE | exact epic required |
| US30 | NOT_TRADEABLE | |
| DE40 | NOT_TRADEABLE | |
| UK100 | NOT_TRADEABLE | |
| J225 | NOT_TRADEABLE | |
| CN50 | NOT_TRADEABLE | |
| BTC | not a Legacy calibration target | microstructure lab |

Ready: none while CLOSED.

Recommended next after GOLD (suitability, not forecasted profit): evaluate **SILVER / US100 / OIL_CRUDE / US500** when TRADEABLE, identity valid, spec valid. One at a time.

Why not mass enable: each market needs broker monetary semantics, Legacy technical compatibility, risk calculation, and position lifecycle proof.
