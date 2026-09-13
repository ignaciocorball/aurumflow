# P8.1 Legacy multi-instrument audit

Do not change Legacy. Compatible ≠ validated.

Hidden GOLD assumptions in `internal/strategy`:

- ATR buckets 8–12 / 12–18 from GOLD backtest
- Fib ± 0.2/0.5 ATR stops; 1R/2R ATR fallback
- London 08:00–17:00 UTC and NY 13:00–22:00 UTC session filters
- pip/tick treated as GOLD-scale points

| Market | Compat | Note |
|---|---|---|
| GOLD | LEGACY_COMPATIBLE | only instrument with runtime history |
| SILVER, OIL_CRUDE | LEGACY_NEEDS_CONFIG | ATR/pip scale differ |
| US100, US500, US30 | LEGACY_NEEDS_CONFIG | index points ≠ GOLD ATR buckets |
| DE40, UK100 | LEGACY_NEEDS_CONFIG | session + scale |
| J225, CN50 | LEGACY_NEEDS_CONFIG | Asia hours |
| BTC | LEGACY_NOT_VALIDATED | microstructure lab, not Legacy venue |

Missing candles must emit `INSUFFICIENT_DATA`, not `NO_SETUP` / Legacy NONE.
