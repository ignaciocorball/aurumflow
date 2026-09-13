# 31 — P4 Institutional Context

Phase-1 graph (not Bloomberg):

```text
Company → Filing
Company → FinancialFact
Insider Form 3/4/5 → Company
13F → manager_holding_placeholder (information table not bulk-parsed)
```

Output: `research/context/institutional_graph.json`

Measured this freeze:

```text
companies = 5   (AAPL MSFT NVDA META AMZN)
filings   = 205
facts     = 152   (Assets, Revenues, NetIncomeLoss, StockholdersEquity tails)
edges     = 486
```

Slow context lookup at research time `t` (`internal/ctxsnap`):

```text
latest SEC filing with AvailableAt <= t
latest CFTC GOLD row with AvailableAt <= t
FINRA week missing if symbol/week absent
macro series Latest(id, t)
```

FINRA: missing is `missing=true`, not zero.

CFTC GOLD identity is exactly `GOLD - COMMODITY EXCHANGE INC.`.
Features: MM long/short/net, SD/PM/OR net, weekly delta, 26w/52w percentile, z-score, crowding.
Current public `f_disagg.txt` supplied 2 GOLD rows this run — historical depth is thin until a yearly zip is attached.

Cross-asset math (`internal/xasset`) is PROXY_CROSS_ASSET. Capital H1 prices were not downloadable this weekend, so GOLD/US100/megacap frames were not joined to the BTC experiment.
