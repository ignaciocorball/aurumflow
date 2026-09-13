# 41 — P5.2 Exhaustion Engine

Domain package: `internal/exhaustion`.

V1 classification is frozen and bitwise-identical to `internal/research.ClassifyFlow`:

```text
DP = D * PressureScore
DP <= -15 → FLOW_EXHAUSTION_CONFIRM
DP >= +15 → FLOW_CONTINUATION_CONFIRM
abs(DP) < 15 → FLOW_NEUTRAL
```

Spec hash unchanged:

```text
f66e744ec7f8e5785bc6d43b2fc8211941ac725bd463baa33782d955f7399b5d
```

`PressureScore` is not inverted. Aggressive selling remains bearish directional pressure. A Legacy LONG plus bearish pressure is sell-side exhaustion of the aggressor, not a short signal from the engine.

## Domain split

```text
DirectionalPressureEngine  where is aggressive flow pushing?
FlowEfficiencyEngine       how much price displacement is that flow producing?
ExhaustionEngine           is strong aggression failing against a valid Legacy setup?
AbsorptionEngine           requires valid L2; not implemented from trade-only history
```

Trade-only snapshots never emit `ABSORPTION_CONFIRMED`. Book capability is `BOOK_CAPABILITY_LIMITED` unless a live book is actually synced. Future L2 fields live on `PassiveLiquidityEvidence` and do not rewrite Exhaustion.

## Live mode

```text
Mode = SHADOW
MayMutateBroker() = false
```

The engine may observe, classify V1, journal `EXHAUSTION_SNAPSHOT` / `FLOW_EXHAUSTION_SIGNAL`, and append prospective inputs. It cannot open, close, filter, or veto capital trades. GOLD DEMO week still executes Legacy only.

`ExhaustionEvidenceScore` is `EXPERIMENTAL_RESEARCH_ONLY`. It is a transparent clamp of ImpactFailure and is not a trading threshold.
