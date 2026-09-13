# P8.2 — Legacy multi-market audit

COMPATIBLE ≠ VALIDATED. Strategy code was **not** retuned.

## Implicit GOLD assumptions (do not change)

| Area | Location | Finding |
|---|---|---|
| ATR buckets | `internal/strategy/composer.go` | +1 for ATR 12–18, −1 for 8–12 — GOLD-scale points |
| Sessions | `internal/strategy/sessions.go` | London 08–17 / NY 13–22 UTC hard filter |
| Scan path | `research.ScanLegacy(m15, h1, h4, cfg)` | requires M15+H1+H4 |
| Stops | composer Fib+ATR | GOLD pip/tick scale |
| Spread | backtest `SpreadPoints` | GOLD-fitted defaults |
| Crypto flag | `CryptoResearchConfig` | BTC research only; not Capital execution |

## Registry

| Market | Status | Validated | Blocker |
|---|---|---|---|
| GOLD | LEGACY_COMPATIBLE | no | Compatible ≠ validated. Needs Sunday lifecycle. |
| SILVER | LEGACY_NEEDS_CONFIG | no | ATR / pip scale GOLD-fitted |
| OIL_CRUDE | LEGACY_NEEDS_CONFIG | no | same + physical EIA UNKNOWN |
| US100 / US500 / US30 | LEGACY_NEEDS_CONFIG | no | index point ATR; US cash session |
| DE40 / UK100 | LEGACY_NEEDS_CONFIG | no | Europe session + index ATR |
| J225 / CN50 | LEGACY_NEEDS_CONFIG | no | Asia hours vs London/NY hard filter |
| BTC | LEGACY_UNSUPPORTED | no | microstructure lab; no Capital BTC execution in P8.2 |

## Runtime distinction

- `INSUFFICIENT_DATA` — Legacy did not evaluate (missing history, or market needs config).
- `NO_SETUP` — GOLD history READY and ScanLegacy ran with required inputs and found no setup.

Non-GOLD markets stay `INSUFFICIENT_DATA` until instrument mechanics are configured. Config lives in `internal/mktcfg` (precision, session, min history, spread gate). **Not** fitted to future returns.

## Why not mass-enable

Each market still needs broker monetary semantics, Legacy technical compatibility, risk calculation, and position lifecycle — one at a time.
