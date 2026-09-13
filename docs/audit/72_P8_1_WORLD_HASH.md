# P8.1 WorldState hash

`WORLD_HASH_V1` is SHA-256 of a canonical JSON document:

- `as_of` (RFC3339Nano UTC)
- source observations used: source, metric, origin, available_at, retrieved_at, value
- RegionState inputs
- AssetClassState inputs
- MarketState inputs (trend, RS, vol, flow, positioning, macro, cross, micro, quality, eligibility)

Omitted: clock-now fields, UI proposal text, `Confidence` (official-source completeness, not a decision probability).

Same WorldState → same hash. One evidence value change → different hash.

Every `OpportunitySnapshot` / `DecisionProposal` / `journals/opportunity-research.jsonl` record stores a non-empty `world_hash`.
