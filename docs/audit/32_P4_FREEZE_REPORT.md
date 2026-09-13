# 32 — P4 Freeze Report

```text
P4 FREEZE: READY WITH RUNTIME LIMITS
```

Connected:

- Capital historical pagination/resume/dedup/gaps (DEMO). Prices 404 this weekend — adapter proven, rows=0.
- Capital streaming adapter exists; live WSS not fully proven (`RUNTIME_LIMITED`).
- Binance Vision aggTrades + SHA256 checksum: 89 days BTCUSDT, 2026-06-15 → 2026-09-11.
- Live collector (`--collect-market-data`) REST+WSS best-effort; hourly gzip JSONL rotation.
- SEC tickers/submissions/companyfacts + graph (5 firms, 205 filings, 152 facts, 486 edges).
- CFTC GOLD identity + no-lookahead features.
- FINRA weekly OTC API: 52 weeks × 5 megacaps; missing ≠ 0.
- Macro interface; FRED = PENDING_FREE_KEY.
- Provenance + catalog operational.

Limits:

- Binance WSS: DNS/TLS/connect OK; first frame read timed out → `ENVIRONMENT_BLOCKED_OR_UNVERIFIED`. REST fallback remains the live path.
- No historical L2. No fabricated book. Radar book capabilities stay off on archives.
- Capital OHLC cache empty this run (`error.prices.not-found`).
- CME MBO / options = NOT_CONNECTED.
- CFTC historical depth = current file only (2 GOLD rows).

Safety: research download authenticated DEMO only. Balance 999.98. No orders. LIVE remains fail-closed.
