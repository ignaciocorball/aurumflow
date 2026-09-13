# P8 Execution eligibility

`ExecutionEligibilityRegistry` is per Capital instrument. LIVE is always `LIVE_PROHIBITED`.

```
ANALYSIS_ONLY → DEMO_DISCOVERED → DEMO_SPEC_VALID → DEMO_CALIBRATED → DEMO_ELIGIBLE
                                                                 → BLOCKED
```

Each market has its own `MonetaryInstrumentSpec`. GOLD MPU is never reused for US100.

## Commands

- `--prepare-market <canonical>` — READ-ONLY discovery / spec / status
- `--calibrate-market <canonical>` — DEMO-only, explicit, always closes
- No bulk auto-calibration

## Capital catalog (2026-09-13 DEMO, weekend)

| Market | Epic | Status | Eligibility |
| --- | --- | --- | --- |
| GOLD | GOLD | CLOSED | DEMO_NOT_CALIBRATED (existing Sunday path) |
| SILVER | SILVER | CLOSED | ANALYSIS_ONLY / discovered |
| OIL | OIL_CRUDE | CLOSED | ANALYSIS_ONLY / discovered |
| US100 | US100 | CLOSED | ANALYSIS_ONLY / discovered |
| US500 | (not uniquely resolved) | UNKNOWN | ANALYSIS_ONLY |
| US30 | US30 | CLOSED | discovered |
| Europe | DE40 | CLOSED | discovered |
| UK | UK100 | CLOSED | discovered |
| Japan | J225 | CLOSED | discovered |
| China | CN50 | CLOSED | discovered |
