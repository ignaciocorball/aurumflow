# 59 — P7 Global Data Sources

| Source | Status | Notes |
|---|---|---|
| BIS | OPERATIONAL (parser + fixtures) | global credit, CB assets, policy rates |
| FED_H41 | OPERATIONAL | WALCL, WRESBAL, WTREGEN, RRPONTSYD isolated behind parser |
| ECB | OPERATIONAL | policy rate, balance sheet, credit key |
| TIC | OPERATIONAL | monthly net LT / purchases / holdings |
| ICI | OPERATIONAL | fund/ETF capital-flow proxy |
| CFTC | OPERATIONAL (reuse + exact names) | GOLD, SILVER, CRUDE, ES, NQ |
| JPX | OPERATIONAL | foreign / individuals / trusts |
| CBOE | OPERATIONAL | OPTIONS_REGIME_PROXY |
| SEC/FINRA | OPERATIONAL reuse | 13F DELAYED; CIK-gated manager watchlist |
| HKEX | PENDING_PUBLIC_STRUCTURED_SOURCE | |
| EIA | PENDING_FREE_KEY + parser ready | |
| WGC / iShares | OPERATIONAL parsers | no intent claims |
| EVENT_CONTEXT | INTERFACE_ONLY | |
