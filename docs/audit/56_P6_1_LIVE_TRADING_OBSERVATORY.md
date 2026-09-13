# 56 — P6.1 Live Trading Observatory Freeze

```text
P6.1 FREEZE: READY
Tests: 244 PASS / 0 FAIL
Vet: PASS
Build: PASS
URL: http://127.0.0.1:8766/
```

UI/observability only. Legacy, FLOW_EXHAUSTION_V1, PressureScore, Absorption, risk, and Capital execution are unchanged.

## Fixes

- Unavailable values render — / N/A / WAITING, not numeric 0
- L2 Book spread binds `l2_spread`, not `gold_spread`
- GOLD CLOSED omits quotes and fake balances
- Console remains GET-only, 127.0.0.1, no mutation controls, no secrets

## Surface

Institutional terminal: safety bar, BTC/GOLD charts, pressure/CVD/efficiency, L2 depth, decision rail, timeline, prospective, frozen research card, GOLD closed workstation, health footer.

SSE `/api/stream` at 500 ms. In-memory `/api/timeseries` last 60 minutes, max 3600 points.

BTC remains SHADOW microstructure. GOLD execution is NOT started by this freeze.
