# 19 — P1 DEMO Validation

## Safety / unit baseline (this machine)

```text
go test ./...     PASS   (152 tests; was 66 / 120)
go vet ./...      PASS
go build ./cmd/bot PASS
```

No Capital.com LIVE calls were made.

## DEMO vs LIVE semantics

```text
Capital.com DEMO runtime validates integration and lifecycle correctness.

It does NOT prove:
- LIVE execution quality
- LIVE liquidity
- LIVE slippage
- strategy profitability
- institutional market data quality
```

`Capital GOLD CFD != CME GC futures`.

## Readonly canary (2026-09-12)

Credentials were loaded in-process from local LIVE `config/config.json` (`api_key`, `identifier`, `password` only) into `AURUMFLOW_DEMO_*`. They were not printed, not copied into `config/demo_config.json`, and not persisted. LIVE `account_id` was not used.

```text
config/demo_config.json:
  environment=demo
  host=https://demo-api-capital.backend-capital.com
  execution_mode=DISABLED
  credentials empty (gitignored)

resolved before CreateSession:
  API ENVIRONMENT: demo
  API BASE URL: https://demo-api-capital.backend-capital.com
  EXECUTION MODE: DISABLED
```

```text
AUTH             PASS
PING             PASS
ACCOUNT          PASS
MARKET DETAILS   PASS
PRICES           PASS
POSITIONS READ   PASS
```

```text
DEMO ACCOUNTS FOUND: 1
SELECTED: id=30**************30 name=Account type=CFD currency=USD preferred=true status=ENABLED balance=1000.00
```

```text
GOLD
  provider: Capital.com
  epic: GOLD
  display name: Gold
  instrument type: COMMODITIES
  currency: USD
  market status: CLOSED (weekend at canary time)
  min deal size: 0.01
  max deal size: 50000
  size increment: 0.01
  lot size: 1
  value_per_point: unknown (not invented; InstrumentSpec.SizingComplete=false)
  last M15: 2026-09-11T20:45:00Z (20 candles; 24h window 404 while CLOSED)
  identity: Capital GOLD CFD != CME GC futures
```

```text
Open positions: 0
LIVE requests: 0
Mutating /positions: 0
```

OPEN → CONFIRM → CLOSE was **not** executed in P1.1 (GOLD CLOSED).

## Lifecycle canary (2026-09-13) — TRADEABLE BTCUSD

GOLD remained CLOSED. The canary was generalized (`--canary-lifecycle [--epic]`) and discovered Bitcoin via `/markets?searchTerm=Bitcoin`.

```text
DISCOVERY: BTCUSD TRADEABLE min=0.0001
OPEN → CONFIRM → READ (by epic/dealId) → CLOSE → CONFIRM CLOSE
FINAL open positions: 0
```

Confirm `dealId` can differ from the live position `dealId`. The second attempt resolved via `GetPositions` by epic. First ambiguous OPEN was not retried; flatten recovered the leftover.

See `22_P1_RUNTIME_PROOF.md`.

GOLD `value_per_point` remains unvalidated for strategy execution (`BROKER_METADATA_ONLY` from lotSize=1). `--demo-week` requires `RUNTIME_VALIDATED`.
