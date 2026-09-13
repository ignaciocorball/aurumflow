# 09 — Observability

## What exists

| Concern | Exists | Status | Evidence |
|---------|--------|--------|----------|
| Structured logging | Optional JSON via `AURUMFLOW_LOG_JSON` | `PARTIAL` | `internal/logger/logger.go` — `{time,level,msg}` only; default is colored text |
| File logs | Yes | `IMPLEMENTED_UNVERIFIED` | `InitFile` → `logs/aurumflow_YYYY-MM-DD_THH-MM-SS.log` |
| Metrics (Prometheus etc.) | No | `NOT_IMPLEMENTED` | |
| Distributed traces | No | `NOT_IMPLEMENTED` | otel appears only as Firebase **indirect** deps in `go.mod` |
| Event / trade / signal journal | Yes (opt-in) | `PARTIAL` | `internal/journal` JSONL; **off** unless `logging.journal_enabled` — missing in example and local root config |
| Execution logs | Yes | `PARTIAL` | logger + journal OrderAttempt/Result |
| Order IDs | Yes | `IMPLEMENTED_UNVERIFIED` | `dealReference`, `dealId` on confirm |
| Broker responses | Errors as strings | `PARTIAL` | `api error %d: %s` body in `Client.Do`; not persisted as raw JSON |
| PnL | Backtest yes; live weak | `PARTIAL` | Live close: `pnl_pct=0`, `exit_price=0` |
| Latency | HTTP timeout only | `NOT_IMPLEMENTED` | no RTT histogram |
| Slippage | Confirm vs signal entry | `PARTIAL` | journal if confirm succeeds |
| Spread | Optional pre-trade | `PARTIAL` | `max_spread` often 0 |
| Errors | Logs + some notifs | `PARTIAL` | |
| Alerts | Pushover / Telegram | `IMPLEMENTED_UNVERIFIED` | `internal/notifications` |
| Live status stream | Firebase RTDB (uncommitted) | `IMPLEMENTED_UNVERIFIED` | `internal/telemetry` + local `docs/RTDB_SCHEMA_V1.md` |

## Can we reconstruct why the bot decided?

**Partially, and only if journal is enabled.**

A complete reconstruction needs, per tick:

| Fact | Logged if journal on | Always on stdout | Missing |
|------|----------------------|------------------|---------|
| Epic, session, state | SetupEvaluated | Yes | |
| Liquidity / sweep / eq | SetupEvaluated | Yes | zone prices unless sweep_debug_log |
| Score + reject reasons | SetupEvaluated / SignalRejected | Yes | |
| H1 / H4 trend | SetupEvaluated | Sometimes | |
| Signal levels + RR | SignalGenerated | Yes | |
| Size / spread | OrderAttempt | Size in success log | |
| Broker accept / fill | OrderResult | dealRef | raw broker body |
| Why position closed | PositionClosed | “all positions closed” | **exit price, reason, PnL live** |
| Why *not* READY | SignalRejected STATE_NOT_READY | Yes | |

If `journal_enabled` is false (current example + local root config), reconstruction depends on ephemeral console/file logs. File logs are unstructured lines.

Notification payload **omits** balance/size/account on Telegram (README) — good for public channels, worse for forensics.

## Notification taxonomy vs emission

| Type | Defined | Emitted from loop/main |
|------|---------|------------------------|
| BOT_START / BOT_STOP | Yes | Yes |
| HEARTBEAT | Yes | Yes |
| DATA_STALE | Yes | Yes |
| SWEEP_CONFIRMED | Yes | If `sweep_debug_log` |
| SIGNAL_GENERATED / REJECTED | Yes | Yes |
| ORDER_SENT / REJECTED | Yes | Yes |
| POSITION_OPENED / CLOSED | Yes | Yes (closed incomplete) |
| DAILY_DD_UPDATE / TRADING_HALTED / MAX_TRADES | Yes | Heartbeat path |
| REGIME_CHANGE | Yes | **Never** (`DEAD_CODE`) |
| ORDER_FILLED | Yes | **Never** (`DEAD_CODE`) |
| POSITION_MODIFIED | Yes | **Never** (`DEAD_CODE`) |
| PNL_MILESTONE | Yes | **Never** (`DEAD_CODE`) |
| API_ERROR | Yes | Not seen in loop (401 handled as relogin) |

## Firebase live schema (local docs, not git)

`docs/RTDB_SCHEMA_V1.md` describes `v1/live/instances/{id}/{meta,status,positions}` and a large `v1/backtests` tree for a **planned React dashboard**. No React app is in this repository.

Uncommitted `telemetry.FirebasePublisher` writes `v1/live/instances/{instanceId}/meta|status|positions/open`.

## Reuse verdict

`KEEP` logger + journal types.  
`KEEP_AND_REFACTOR` notifications.  
`WRAP` telemetry.  
P1 must default `journal_enabled=true` for any DEMO run.
