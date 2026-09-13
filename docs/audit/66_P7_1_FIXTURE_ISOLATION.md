# P7.1 Fixture isolation

## Modes

`FIXTURE` · `LIVE_OFFICIAL` · `CACHE_OFFICIAL`

Fixtures may be used only in tests, offline parser validation, or explicit `--fixture-world`.

They must not enter `--shadow-runtime`, `--sunday-runtime`, `--demo-week`, `/api/world`, OpportunitySurface, or DecisionOrchestrator.

## Gates

- Every `ContextObservation` carries `origin`, official URL, dataset/series IDs, period, `published_at`, `available_at`, `retrieved_at`, parser version, raw hash.
- Production `WorldState.At` drops any observation that is not `LIVE_OFFICIAL` or `CACHE_OFFICIAL`.
- `UsableAt` rejects `published_at > now` even when the observation period is earlier.
- `CURRENT_WORLD_STATE_VALID` means the **resulting** state has no fixtures. Unknown sources are allowed.

## Tests

- production + fixture observation → rejected
- observation published tomorrow → unavailable today
- live/cache origins accepted
- empty origin fail-closed in production
