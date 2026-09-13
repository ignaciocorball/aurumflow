# ACTIVE SESSION SURFACE

Status: **PREOPEN_READY** — traditional markets were CLOSED during P8.2 engineering. Live Asia/Europe/US/commodity features are not invented.

Command:

```text
go run ./cmd/bot --intelligence-runtime --status-addr 127.0.0.1:8766 --soak-minutes 120
```

## Sunday 2026-09-13 last observed Capital tape (P8.1, CLOSED)

| Market | Mid | Status | Attention | Note |
|---|---|---|---|---|
| BTC | 77316 | TRADEABLE | 45 TIER_C | only live tape; micro depended on process |
| GOLD | 4349.04 | CLOSED | — | ratio vs SILVER ≈ 67.44 |
| SILVER | 64.49 | CLOSED | — | |
| OIL_CRUDE | 96.508 | CLOSED | — | physical EIA UNKNOWN |
| US100 | 29372.1 | CLOSED | — | |
| US500 | epic US500 | CLOSED | — | |
| US30 | 52539 | CLOSED | — | |
| DE40 | 25566 | CLOSED | — | |
| UK100 | 10665.6 | CLOSED | — | |
| J225 | 64695 | CLOSED | — | |
| CN50 | 14544 | CLOSED | — | |

Live US / Europe / Asia / precious / energy frames: **not computed as live** (quotes CLOSED).

WorldState layers: SlowContext (official) + LiveMarketFrame (queued) + MicrostructureFrame (BTC sensors when unified runtime is up).

WORLD_HASH_V1 unchanged.
