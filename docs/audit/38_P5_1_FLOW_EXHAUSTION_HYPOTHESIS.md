# 38 — P5.1 Flow Exhaustion Hypothesis

Pre-registered as `FLOW_EXHAUSTION_V1` before EXTERNAL_HOLDOUT_1 was downloaded.

- Spec: `research/specs/FLOW_EXHAUSTION_V1.json`
- Narrative: `research/specs/FLOW_EXHAUSTION_V1.md`
- Registry: `research/specs/REGISTRY.json`
- git_commit at freeze: `e8183c55e32a11bfddb49e641c3af66968443a2d`
- spec_hash: `f66e744ec7f8e5785bc6d43b2fc8211941ac725bd463baa33782d955f7399b5d`

Discovery `2026-06-15 → 2026-09-11` is already inspected. It is not confirmatory evidence.

```text
D * P <= -15  → FLOW_EXHAUSTION_CONFIRM
D * P >= +15  → FLOW_CONTINUATION_CONFIRM
|P| < 15      → FLOW_NEUTRAL
```

PressureScore formula, Legacy, and the |P|>=15 threshold were not changed.

Primary endpoint: 15-minute mean directional return, exhaustion vs Legacy baseline.

This is not absorption. Book data is unavailable.

RADAR_MODE remains SHADOW.
