# 23 — P2 Dual Plane Architecture

```text
              INTELLIGENCE PLANE
                     │
                MarketEvent
                     │
                FeatureEngine
                     │
                  Radar
                     │
                DecisionContext
                     │
                     ▼
             EXISTING STRATEGY
                     │
                     ▼
                Risk Engine
                     │
                     ▼
               EXECUTION PLANE
                     │
               Capital DEMO
```

The Intelligence Plane has no `ExecutionProvider` and cannot mutate the broker.

## Providers

```text
ExecutionProvider   KEEP (P1)
MarketDataProvider  NEW  (internal/md)
  Name / Capabilities / Run(ctx, *Bus)
  Caps: Quotes, Trades, Candles, BookSnapshot, BookDelta, Health
```

Missing capabilities are not stubbed as empty streams.

## Canonical identity

```text
Asset GOLD
  Sensor  CME GC          relationship = correlated_underlying_proxy  NOT_CONNECTED
  Execution Capital GOLD CFD

Asset BITCOIN
  Sensor  Binance BTCUSDT futures   relationship = proxy
  Execution Capital BTCUSD CFD      relationship = proxy

Asset NASDAQ100
  Sensor  CME NQ          relationship = correlated_underlying_proxy  NOT_CONNECTED
  Execution Capital US100 CFD
```

Never IDENTICAL across venues.

## MarketEvent

`TradeEvent`, `QuoteEvent`, `BookSnapshotEvent`, `BookDeltaEvent`, `CandleEvent`, `ProviderHealthEvent`

Each event carries event timestamp, receive timestamp, provider, venue, instrument, sequence when available.

## Bus

In-process Go channel, bounded buffer, context cancellation, drop counter. No Kafka. No Redis requirement.

## Persistence

Weekend storage is JSONL:

```text
journals/demo-week.jsonl
journals/radar.jsonl
journals/instruments/{EPIC}.json
journals/soak-metrics.json
```

Raw book capture is not retained indefinitely.

## Legacy Composer

`ComposerInput.Context *DecisionContext` is optional and ignored by scoring. Existing signals remain reproducible without Radar.
