# P8.1 Live market surface

Slow official WorldState (TIC, CFTC, Fed, ECB, BIS, JPX, …) is cached.

Live Capital frame (5s materialize) overlays:

- bid/ask/mid/spread, marketStatus, event_time, received_at, age
- past-only returns 5m/15m/1h/4h/1d and realized vol
- momentum STRONG_UP/UP/FLAT/DOWN/STRONG_DOWN (fixed rules)
- volatility LOW/NORMAL/HIGH/EXTREME vs own history
- relative strength among live peers only
- sessions ASIA/EUROPE/US/OVERLAP/TRANSITION; per-market PREOPEN/OPEN/CLOSED
- US mega-cap breadth, gold/silver ratio, DE40 vs UK100, J225 vs CN50
- correlation-context graph (not causality)

CLOSED or stale quotes never rank as live momentum. EIA physical energy stays UNKNOWN.

BTC `MicroAvailable` is sensing capability from `MicrostructureCapabilityRegistry`. It is not a trade signal.

`--prepare-demo-universe` is READ-ONLY. Calibration remains an explicit command. `MULTI_MARKET_DEMO_EXECUTION=OFF`.

Prospective records: every 5m or material tier/eligibility change; versions `ATTENTION_SCORE_V1`, `WORLDSTATE_V1`, `MARKET_FEATURES_V1`, git commit.
