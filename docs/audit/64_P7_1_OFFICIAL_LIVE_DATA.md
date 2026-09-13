# P7.1 Official live data

Parser working is not the same as a live source working. Production WorldState now loads only `LIVE_OFFICIAL` or `CACHE_OFFICIAL` observations.

## Runtime (2026-09-13)

| Source | Fetch | Latest observation | Notes |
| --- | --- | --- | --- |
| FED H.4.1 | LIVE_OFFICIAL | 2026-09-10 | HTML current release. Assets 6740.6bn, reserves 2991.3, TGA 883.3, ON RRP 351.7. FRED CSV timed out; FedProvider still abstract. 1/52 weeks (current print only). |
| TIC | LIVE_OFFICIAL | 2026-06 | Official SLT table1+table5. July 2026 is **not published**. Next release 2026-09-16. 9/12 months. |
| ICI | FAIL 403 | — | Public weekly page forbidden. No fixture substitution. 0/52. |
| Cboe | historical file only | 2019-10-04 | Official `totalpc.csv` ends 2019. Not used as current WorldState. 0/252. |
| JPX | LIVE_OFFICIAL | 2026-09-01 | Current index lists post-April-2026 weekly XLS (`stock_val_1_260901.xls`). Values not parsed from XLS this run. 5/52 published weeks. |
| BIS | LIVE_OFFICIAL | 2026-Q1 | `WS_GLI` `.USD` CSV. Slow context. 2/4 quarters. |
| ECB | LIVE_OFFICIAL | 2026-06-17 usable | MRR 2.40. Official series also lists 2026-09-16; rejected as future. 2/52 change-dates. |
| CFTC | LIVE_OFFICIAL | 2026-09-08 | GOLD and SILVER exact names resolved. Crude / ES / NQ absent → UNKNOWN. 1/52. |
| EIA | PENDING_FREE_KEY | — | No `EIA_API_KEY`. |
| HKEX | PENDING_PUBLIC_STRUCTURED_SOURCE | — | No stable structured Stock Connect file. |
| WGC | CACHE_MANUAL_OR_PUBLIC | — | No stable machine-readable public file. |
| iShares | PENDING_VERIFIED_FUND_IDENTITY | — | No fund added without official identity. |

Raw archive: `data/world/raw/` (gitignored). Normalized cache: `data/world/normalized/`.
