# OPENING SESSION 2026-09-13

Status: **VALIDATING** — opening snapshot, not a completed 2h checkpoint.
Soak started: `2026-09-13T22:06:20Z` (`--intelligence-runtime` on `127.0.0.1:8766`).
GOLD demo-week: **NOT STARTED**. Monetary gate blocked (`BROKER_METADATA_ONLY`).

Broker is authoritative. Clock is not.

## Code

- HEAD: `c8e37df` (P9)
- P8.2 freeze: `53e3098`
- Tests: 341 PASS
- Vet: PASS
- Build: PASS

## P0 WorldState

Old `:8766` process (stopped 22:06:00Z, exit 0):

```
PID 51716
bot.exe --shadow-runtime --l2-provider auto --soak-minutes 1440 --status-addr 127.0.0.1:8766
started 2026-09-13T19:53:26Z
```

That binary served fixture-era `/api/world`: Liquidity CONTRACTING, Risk RISK_OFF, Opportunity GOLD 90 / US100 60 / US500 60 / BTC 55. No `Valid`, no `Hash`, no `Origins`.

Root cause: stale P7 `--shadow-runtime` `bot.exe`, not the official cache and not a UI seed. Disk `CURRENT_WORLD_STATE.json` was already production (UNKNOWN / MIXED / GOLD 15).

Graceful stop log: `trades=35529 deltas=77916 snaps=2 gaps=0 resyncs=0 p50=194.2 p95=471.3`. `:8765` was never up.

New runtime: `go run ./cmd/bot --intelligence-runtime --status-addr 127.0.0.1:8766 --l2-provider auto --soak-minutes 1`

## Production WorldState (live `/api/world`)

```
Valid=CURRENT_WORLD_STATE_VALID
RejectedFix=0
Origins=LIVE_OFFICIAL
truth=LIVE_OFFICIAL
Liquidity=UNKNOWN
Risk=MIXED
Hash (22:06:26Z)=5049c843f7e4ca296bfee442c41aeba49e3d32425cb38cbfb5e2b59f89194079
Hash (later live frame)=ac097c2c78d04988ab98ec7945fd91a86102c1ade0145aa7af812a0dcef2ea96
```

Official cache (retrieved 22:06:26Z):

| Source | Latest | Origin | Note |
|---|---|---|---|
| FED_H41 | 2026-09-10 | LIVE_OFFICIAL | H.4.1 current HTML |
| TIC | 2026-06 | LIVE_OFFICIAL | July not published; next expected 2026-09-16 |
| ECB | meta lists 2026-09-16 | LIVE_OFFICIAL | future publication rejected by UsableAt |
| BIS | 2026-01-01 | LIVE_OFFICIAL | WS_GLI |
| CFTC | 2026-09-08 | LIVE_OFFICIAL | GOLD/SILVER positioning |
| JPX | 2026-09-01 | LIVE_OFFICIAL | investor-type |
| ICI | absent | — | no production cache |
| CBOE | official file ends 2019-10-04 | LIVE_OFFICIAL | HISTORICAL_FILE_NOT_CURRENT; not testdata fixture |

## First observed TRADEABLE (Capital)

`journals/gold-observatory.json` was CLOSED at `2026-09-13T22:01:08Z`.
First Capital TRADEABLE this run: `2026-09-13T22:06:53Z` GOLD bid=4337.12 offer=4337.62.
World live frame already had TRADEABLE tapes at `2026-09-13T22:06:26Z`.

| Market | First observed | Bid | Ask | Mid | Spread | Tape / hist |
|---|---|---|---|---|---|---|
| GOLD | TRADEABLE 22:06:26Z | 4338.22 | 4338.72 | 4338.47 | 0.50 | LIVE / READY |
| SILVER | TRADEABLE 22:06:26Z | 63.953 | 64.033 | 63.993 | 0.08 | LIVE / READY |
| OIL_CRUDE | TRADEABLE 22:06:26Z | 99.196 | 99.241 | 99.2185 | 0.045 | LIVE / READY |
| US100 | TRADEABLE 22:06:26Z | 29001 | 29002.8 | 29001.9 | 1.8 | LIVE / READY |
| US500 | TRADEABLE 22:06:26Z | 7612 | 7612.6 | 7612.3 | 0.6 | LIVE / READY |
| US30 | TRADEABLE 22:06:26Z | 52382.2 | 52384.2 | 52383.2 | 2.0 | LIVE / READY |
| DE40 | TRADEABLE 22:06:26Z | 25397.1 | 25405.1 | 25401.1 | 8.0 | LIVE / READY |
| UK100 | TRADEABLE 22:06:26Z | 10619.9 | 10622.9 | 10621.4 | 3.0 | LIVE / READY |
| J225 | TRADEABLE 22:06:26Z | 63507.5 | 63517.5 | 63512.5 | 10.0 | LIVE / READY |
| CN50 | CLOSED | 14539 | 14549 | 14544 | 10.0 | CLOSED / STALE |
| BTC | TRADEABLE (24/7) | 76956.9 | 77006.9 | 76981.9 | 50.0 | LIVE / READY |

## Feature readiness (opening)

All TRADEABLE markets above: PRICE_LIVE, MOMENTUM_READY, VOLATILITY_READY, CROSS_ASSET_READY, LEGACY_READY (history).
CN50: NOT_READY / DEGRADED / CLOSED.

Legacy ScanLegacy runs only for GOLD when history READY. GOLD setup=`NO_SETUP`. Other markets remain `INSUFFICIENT_DATA` (not false `NO_SETUP`).

## Opportunity (opening / ~+3m)

| Rank | Market | Attention | Coverage | Tier | Setup | Eligibility |
|---|---|---|---|---|---|---|
| 1 | GOLD | 60 | 62.5 | TIER_B | NO_SETUP | ANALYSIS_ONLY |
| 2 | SILVER | 60 | 62.5 | TIER_B | INSUFFICIENT_DATA | ANALYSIS_ONLY |
| 3 | BTC | 55 | 62.5 | TIER_B | INSUFFICIENT_DATA | ANALYSIS_ONLY |
| 4 | CHINA_HK | 45 | 50 | TIER_C | INSUFFICIENT_DATA | ANALYSIS_ONLY |
| 5 | DE40 | 45 | 50 | TIER_C | INSUFFICIENT_DATA | ANALYSIS_ONLY |

No Tier A. These are not fixture ranks (old process was GOLD 90 / US100 60 / US500 60).
+15m / +30m / +1h / +2h snapshots: **PENDING** (soak running).

## GOLD Sunday Go

| Gate | Result |
|---|---|
| Capital DEMO | PASS (balance ~999.97, LIVE IMPOSSIBLE) |
| Account | PASS |
| GOLD TRADEABLE | PASS |
| Open positions before | 0 |
| Kill switch | off / healthy |
| Journal | writable |
| Prepare canary | OPEN/CONFIRM/READ/CLOSE size=0.01 BUY; positions 0 |
| Calibrate canary | same; VALUE_PER_POINT UNKNOWN; pnl=0.0000 |
| Positions after | 0 |
| Monetary | **BROKER_METADATA_ONLY** |
| Preflight | READY_WITH_WARNINGS (`gold_monetary` WARN) |
| Demo-week | **BLOCKED — no bypass** |

Canary 1: dealRef `o_22ffb69a-5e37-4f84-9511-01900d687171` open 4337.93 close 4337.41.
Canary 2: dealRef `o_d1fdd26a-ad4c-4b28-8f84-21bd4aa8aac8` open 4337.13 close 4336.52.
Existing prepare path refuses RUNTIME_VALIDATED from a single-trade close. UPL multi-sample is required and is not implemented as a new feature in this run.

SILVER / US100 / OIL / indices: TRADEABLE, **not calibrated**. No CalibrationQueue store exists in-repo; they stay queued by policy only.

## BTC micro (new runtime, ~3 min)

trades=6955 deltas=1595 gaps=0 drops=0 resyncs=1 book_synced=true book_age_ms≈52 p50≈198ms p95≈300ms quality=GOOD V1=FLOW_NEUTRAL Absorption raw=NOT_SUPPORTIVE (engine unchanged).

## Event timeline

Raw `AbsorptionStatus` still updates every tick on `/status`.
UI timeline: 5s dwell for MIXED/NOT_SUPPORTIVE; SUPPORTIVE / STRONGLY_SUPPORTIVE / UNAVAILABLE emit immediately.
Opening events included UNAVAILABLE → NOT_SUPPORTIVE (material) and later dwell-qualified MIXED/NOT_SUPPORTIVE holds (>5s), not 1Hz flicker.

## Safety

LIVE: IMPOSSIBLE / FAIL-CLOSED
Multi-market DEMO execution: OFF
Broker mismatches: none
HALT_NEW_ORDERS: not raised
`:8766` intelligence left running
`:8765` not started

## 2h checkpoint

Not yet due. Revisit `research/reports/OPENING_SESSION_2026_09_13.md` after 00:06:20Z.
