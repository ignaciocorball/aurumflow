# P8.2 — Active session validation

## Soak command

```text
go run ./cmd/bot --intelligence-runtime --status-addr 127.0.0.1:8766 --soak-minutes 120
```

If traditional markets are still CLOSED at engineering freeze: **PREOPEN_READY**. Do not fake active-market results.

## Must observe (when TRADEABLE)

Market status transitions (CLOSED↔TRADEABLE, PREOPEN/OPEN/TRANSITION), live quotes, history warmup, feature maturity, WorldState rematerialization, Opportunity ranking, DecisionProposal changes, prospective writes.

## Metrics

duration, markets tradeable, quote updates, stale frames, history fetches, WorldState updates, Opportunity changes, Decision proposals, Legacy evaluations, BTC L2 deltas, drops, errors, memory.

## Live frames (past-only, when ≥2 markets live)

US: US100 vs US500 leadership, US30 divergence, mega-cap breadth/median/concentration.

Precious: Gold/Silver ratio, ratio change, RS, vol difference. CFTC remains weekly.

Europe: DE40 vs UK100 RS, breadth proxy, session return, vol. ECB slow.

Asia: J225 vs CN50 RS, session leadership, vol, handoff to Europe. JPX remains capital-flow evidence.

Energy: OIL price/momentum/vol/relative only. Physical oil UNKNOWN until EIA.

`SessionHandoffObservation` is descriptive. No strategy inference.

## Opportunity

ATTENTION_SCORE_V1 frozen. Do not weaken tiers to force TIER_A.

Journal only: attention move ≥5, tier/setup/eligibility/data-quality/proposal change.

Outcomes: 15m / 1h / 4h / 1d labeled VALID / PARTIAL_SESSION / UNAVAILABLE. Do not label closed/gap paths.

## Isolation

Quote staleness / provider failure / history failure / stale WorldState / BTC book unsync → degrade, block, recover. Never invent.
