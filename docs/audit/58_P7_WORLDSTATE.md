# 58 — P7 WorldState

`WorldState.At(t)` assembles only `ContextObservation` rows with `AvailableAt <= t`.

Composite liquidity is labelled HEURISTIC_V1. Risk regime is descriptive (RISK_ON/OFF/MIXED/TRANSITION) from explicit ICI, Cboe, and Fed components. Weights are not fit to returns.

Historical as-of uses weekly points over 12 months. Revisions keep period + retrieved_at + available_at.
