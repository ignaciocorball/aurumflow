# 13 — Institutional Radar Gap Analysis

Compare **desired modules** to **this repo**. No design of Kafka/K8s/ML.

## Module gap

| Module | Classification | Evidence / note |
|--------|----------------|-----------------|
| Market Data Gateway | **inexistente** | Concrete `market.Client` only |
| Event Normalizer | **inexistente** | `models.Candle` is the only normalized event |
| Event Bus | **inexistente** | In-process function calls |
| CME Adapter | **inexistente** | |
| Nasdaq Adapter | **inexistente** | US100 is a CFD poll |
| Options Adapter | **inexistente** | |
| Capital.com Adapter | **parcialmente existente** / **reutilizable** | REST session, prices, positions, orders |
| Order Flow Engine | **inexistente** | |
| Book Engine | **inexistente** | |
| Absorption Engine | **inexistente** | |
| Liquidity Engine | **parcialmente existente** | Candle wick zones, not book depletion |
| Options Engine | **inexistente** | |
| Cross Asset Engine | **inexistente** | Multi-process bots are isolated |
| Session Engine | **existente** / **reutilizable** | `strategy/sessions.go` |
| Regime Engine | **parcialmente existente** | H4/H1 swing trend + unused `PhaseExpansion` |
| Feature Store | **inexistente** | |
| Institutional Pressure Engine | **inexistente** | Composer score is **not** this |
| Alert Engine | **parcialmente existente** | Notifications + journal |
| Explanation Engine | **parcialmente existente** | Reject reasons + setup journal; no narrative model |
| Historical Replay | **parcialmente existente** | OHLC bars only |
| Research / Backtesting | **parcialmente existente** | Go + local Python |
| Dashboard | **inexistente** (schema only) | `docs/RTDB_SCHEMA_V1.md` local |
| Execution Adapter | **parcialmente existente** | Capital POST only |
| Risk Engine | **parcialmente existente** | Size / max / daily DD |

## Future operational states vs code

| Future state | Equivalent today? | Map? |
|--------------|-------------------|------|
| ACCUMULATION | No | Do not invent |
| DISTRIBUTION | No | Do not invent |
| ABSORPTION | No | Do not invent |
| EXPANSION | Name-only constant | Do not equate to impulse |
| NO_TRADE | Yes — IDLE, OFF_HOURS, MARKET_CLOSED, DATA_FROZEN, rejects, COOLDOWN | Operational only |

Sweep + pullback-to-equilibrium is a **setup narrative**, not a verified accumulation/distribution detector.

## Institutional Pressure Score — data available TODAY

Target card vs current fields:

| Card field | Available now? | Source | % contribution (equal weight 11 fields) |
|------------|----------------|--------|----------------------------------------|
| Instrument | Yes (CFD epic, not NQ) | `Loop.Epic` | 0.5 / 1 (wrong identity for NQ) |
| Direction | Yes | `TradeSignal.Direction` | 1 |
| Pressure 0–100 | No (score ~0–10 + context) | Composer | 0.3 |
| Aggressive Flow | No | — | 0 |
| Absorption | No | — | 0 |
| Book Imbalance | No | — | 0 |
| Liquidity Depletion | Candle sweep only | IntentDetector | 0.2 |
| Options Pressure | No | — | 0 |
| Cross Asset Confirmation | No | — | 0 |
| Regime | Weak H4/H1 trend | marketstate | 0.4 |
| Persistence | State TTL / candles in setup | StateMachine | 0.3 |
| Key Levels | Swings, fib, zones | structure/strategy | 0.6 |
| Invalidation | Opposite structure | StateMachine | 0.4 |
| Confidence | `score/10` | TradeSignal | 0.4 |

**Approx. 18% of score inputs exist, and several are the wrong asset class.**

Radar platform readiness (modules + data): **~14%**.

## Identity warning

A future “Instrument: NQ” card **cannot** be filled from `US100` CFD candles without an explicit mapping layer and a disclaimer. That layer does not exist.
