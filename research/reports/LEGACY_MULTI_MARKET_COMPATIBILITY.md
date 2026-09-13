# LEGACY MULTI-MARKET COMPATIBILITY

COMPATIBLE ≠ VALIDATED. No strategy retune.

| Market | Compatible | Validated | Blocker |
|---|---|---|---|
| GOLD | LEGACY_COMPATIBLE | no | ATR 8–18 / London+NY / GOLD-scale stops. Needs lifecycle evidence. |
| SILVER | LEGACY_NEEDS_CONFIG | no | ATR and pip/tick GOLD-fitted |
| OIL_CRUDE | LEGACY_NEEDS_CONFIG | no | same; physical oil UNKNOWN |
| US100 | LEGACY_NEEDS_CONFIG | no | index ATR ≠ GOLD buckets; US cash session |
| US500 | LEGACY_NEEDS_CONFIG | no | same |
| US30 | LEGACY_NEEDS_CONFIG | no | same |
| DE40 | LEGACY_NEEDS_CONFIG | no | Europe session + index ATR |
| UK100 | LEGACY_NEEDS_CONFIG | no | same |
| J225 | LEGACY_NEEDS_CONFIG | no | Asia hours vs London/NY hard filter |
| CN50 | LEGACY_NEEDS_CONFIG | no | same |
| BTC | LEGACY_UNSUPPORTED | no | microstructure lab; no Capital BTC execution in P8.2 |

Runtime: GOLD may reach `NO_SETUP` only after history READY and ScanLegacy actually runs. All other markets remain `INSUFFICIENT_DATA`.

Instrument mechanics (precision, session, min history, spread gate) live in `internal/mktcfg`. Not fitted to returns.
