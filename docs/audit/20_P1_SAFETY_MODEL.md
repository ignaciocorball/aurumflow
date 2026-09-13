# 20 — P1 Safety Model

## Invariant

> LIVE FAILS CLOSED.

A P1 binary must not authenticate to Capital.com LIVE, fetch LIVE data, or send LIVE orders — regardless of leftover JSON, empty fields, typos, or old scripts.

## Two axes

| Axis | Values | Role |
|------|--------|------|
| `api.environment` / legacy `api.mode` | `demo` only | Broker environment |
| `api.execution_mode` | `DISABLED` `DRY_RUN` `DEMO` | Permission to mutate |

`live` on either axis is a **fatal parse**, not a degraded mode.

## Control points (defense in depth)

1. **Config** `ApplyP1BrokerLock` — missing/empty/invalid/LIVE environment → refuse. LIVE URL → refuse. Arbitrary host → refuse. Then pin `BaseURL` to `DemoAPIHost`.
2. **Env** `AURUMFLOW_API_ENVIRONMENT=live` or `AURUMFLOW_EXECUTION_MODE=LIVE` → refuse.
3. **HTTP** `Client.Do` → `assertNotLive` before any request.
4. **Entrypoint** `NewDemoClient()` for the operational bot; banner; second LIVE check.
5. **Scripts** `bots/*/run.ps1` and `scripts/run.ps1` only launch the binary. They cannot bypass the lock.

`httptest.Server.URL` is injected via `market.NewClient(url)` in tests only. Operational `Load` never accepts localhost or custom hosts.

## Execution modes

| Mode | Market data | Signals | Size | Journal | POST/PUT/DELETE |
|------|-------------|---------|------|---------|-----------------|
| DISABLED | yes | yes | no order | skip event | no |
| DRY_RUN | yes | yes | yes | `would_have_sent` | no (`DryRunProvider`) |
| DEMO | yes | yes | yes | full lifecycle | yes, demo host only |

## Kill switch

`HALT_NEW_ORDERS` if any of:

- `kill_switch.enabled=true`
- `AURUMFLOW_KILL_SWITCH=1`
- sentinel file exists (default `.aurumflow.kill`)

Checked immediately before broker mutation. Analysis continues.

## Flatten

`--flatten` requires DEMO environment + DEMO execution mode + `--flatten-confirm` or interactive `FLATTEN-DEMO`. Uses `NewDemoClient` only.

## Account / sizing

- No `accounts[0]` fallback.
- Missing `account_id` uses session `currentAccountId` only.
- Size < broker min → reject (does not upsize).
- Incomplete `InstrumentSpec` → do not trade.

## What old LIVE configs do

They remain on disk. `config.Load` prints `STARTUP REFUSED: LIVE broker environment is disabled in P1.` and exits before `CreateSession`.
