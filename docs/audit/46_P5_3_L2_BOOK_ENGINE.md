# 46 — P5.3 L2 book engine

Local book (`internal/book`):

- snapshot accepted
- Binance futures delta + generic seq delta
- gap → UNSYNCED (`Gaps++`, `Resyncs++`)
- UNSYNCED → `FeaturesAccepted() == false`
- resync from REST snapshot

Exposed: synced, age, sequence, gaps, resyncs, best bid/ask, spread, mid, microprice, depth 1/5/10/20, imbalance 1/5/10/20.

`internal/bookfeatures` adds depletion, replenishment (same-level refill), persistence, `POTENTIAL_SWEEP` (never `INSTITUTIONAL_SWEEP`), and `PassiveLiquidityResponse` normalized to Legacy direction (LONG support=bid, SHORT support=ask).

## Providers

Binance USD-M public (no credentials, no trading):

```text
wss://fstream.binance.com/market/stream?streams=btcusdt@aggTrade
wss://fstream.binance.com/public/stream?streams=btcusdt@depth@100ms
```

Legacy `wss://fstream.binance.com/ws` was retired 2026-04-23 and caused FIRST FRAME TIMEOUT. Official 2026 split is OPERATIONAL in this environment.

OKX SWAP public (`BTC-USDT-SWAP`) is implemented as `CORRELATED_PROXY`, not IDENTICAL to Binance BTCUSDT.

Normalized events remain `Trade`, `BookSnapshot`, `BookDelta`, `ProviderHealth`.
