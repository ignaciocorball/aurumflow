# P8.2 freeze report

## Freeze

`READY_PREOPEN` if traditional markets remain CLOSED during the engineering run.

`ACTIVE_SESSION_VALIDATED` requires ≥2h unified runtime with multiple traditional markets TRADEABLE, mature live features, natural Opportunity changes, no stale-data contamination, no process failure, valid prospective records.

## Invariants held

- P8.1 committed independently (`feat: freeze live cross-asset surface and multi-market readiness`)
- Unified `--intelligence-runtime` exists; no ExecutionProvider
- BTC micro and Global Surface coexist in one process
- History warmup + FeatureReadiness + INSUFFICIENT_DATA vs NO_SETUP
- Legacy compatibility audited; COMPATIBLE ≠ VALIDATED
- WORLD_HASH_V1 semantics unchanged
- ATTENTION_SCORE_V1 / FLOW_EXHAUSTION_V1 / Absorption / Legacy untuned
- No LIVE, no mass calibration, no multi-market DEMO
- GOLD `sunday_go.ps1` remains the only autonomous DEMO path

## Active soak (when markets open)

```text
go run ./cmd/bot --intelligence-runtime --status-addr 127.0.0.1:8766 --soak-minutes 120
```
