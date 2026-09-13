# P7.1 + P8 freeze

## P7.1 READY

Production WorldState contains no fixtures. Official current sources were fetched. Freshness is auditable. Core Fed / TIC / ICI(blocked 403) / Cboe(historical only) / CFTC / JPX / ECB / BIS reported honestly. OpportunitySurface recomputed from official data — GOLD is not preserved as a high-confidence #1.

## P8 READY_SHADOW

Eligibility registry, portfolio risk, Capital discovery, DecisionProposal, and `--global-shadow` work. Multi-market proposals recorded. Broker mutation capability = NONE. `MULTI_MARKET_DEMO_EXECUTION=OFF`.

## Safety

LIVE fail-closed at every layer. No new LIVE config option. GOLD `sunday_go.ps1` remains the only enabled autonomous DEMO path when TRADEABLE.
