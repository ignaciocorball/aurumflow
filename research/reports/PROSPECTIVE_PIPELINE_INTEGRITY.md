# Prospective pipeline integrity

## Recorders (independent)

| Recorder | Path | Gate |
|---|---|---|
| **OpportunityResearchRecorder** | `journals/opportunity-research.jsonl` | 5m cadence or Tier/Eligibility/Setup/DataQuality change. Requires `world_hash`. **No V1 signal required.** |
| **FlowExhaustionProspectiveRecorder** | `research/prospective/FLOW_EXHAUSTION_V1/` | V1 / research ScanLegacy on BTC micro only |
| Outcomes | `journals/opportunity-outcomes.jsonl` | Append-only; **does not rewrite t0** |

P8.7 `Inputs=0` counted **V1** (`FLOW_EXHAUSTION_V1`). Opportunity snapshots **did** exist in `opportunity-research.jsonl` from `worldsnap`. That was a measurement error, not a missing recorder.

## Outcome labeler

`LabelMature` runs on the intelligence live frame (~30s). It calls `BuildOutcome` / `AppendOutcomeOnly` and writes **only** `journals/opportunity-outcomes.jsonl`.  
Horizons: 15m / 1h / 4h / 1d using prices with `t > t0` and `t <= t0+horizon`.  
Weekend/empty path → `UNAVAILABLE`. Sparse path → `PARTIAL_SESSION`.  
t0 JSONL is immutable. V1 FLOW_EXHAUSTION recorder is a different directory and does not gate Opportunity snapshots.

## Data-quality intervals

`internal/dataint.Interval`: VALID / DEGRADED / INVALID from drops, gaps, resyncs, book sync. Prospective samples store `integrity_status`. Do not discard degraded rows.

## Retroactive 2h

Missing V1 inputs are **not fabricated**. Historical opportunity lines already on disk stay as-is (some early rows have empty `world_hash` and are not valid prospective). New samples after P8.8 deploy include Salience + Attention components.

`LabelMature` only labels rows with `integrity_status` (P8.8+ shape) and a real post-t0 price path. A one-time pass that attempted to label the entire historical jsonl was quarantined as `journals/opportunity-outcomes-RECONSTRUCTED_DIAGNOSTIC.jsonl` and must never be mixed with true prospective samples.
