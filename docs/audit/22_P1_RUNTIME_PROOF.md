# 22 — P1 Runtime Proof

## P1 FREEZE

```text
P1 FREEZE: READY
CAPITAL DEMO EXECUTION PLANE: VERIFIED_WORKING
```

Proven on Capital.com DEMO with a TRADEABLE instrument because GOLD was CLOSED.

## Safety

```text
API ENVIRONMENT: demo
API BASE URL: https://demo-api-capital.backend-capital.com
EXECUTION MODE: DEMO (canary only)
LIVE requests: 0
```

The canary never infers LIVE. Mutation requires `APIEnvironment=DEMO` and `ExecutionMode=DEMO`.

## Discovery

Search order: Bitcoin, then Ethereum. No hardcoded epic.

```text
search Bitcoin → TRADEABLE
epic: BTCUSD
name: Bitcoin / USD (broker catalog)
type: CRYPTOCURRENCIES
currency: USD
marketStatus: TRADEABLE
minDealSize: 0.0001
maxDealSize: known
minSizeIncrement: 0.0001
lotSize: 1
contractSize: 0
scalingFactor: 0
```

Identity: Capital BTCUSD CFD ≠ Binance BTCUSDT.

## Controlled mutation

```text
reason: CANARY_INFRASTRUCTURE_TEST
size: broker minimum 0.0001
positions opened: 1
retries of ambiguous OPEN: 0
```

First attempt (2026-09-13T02:56:32Z):

```text
OPEN submitted dealRef=o_13623086-...
CONFIRM ACCEPTED dealId=...e55...
READ/CLOSE by confirm dealId failed (broker dealId mismatch)
HALT_NEW_ORDERS
DEMO flatten recovered 1 BTCUSD position
```

Second attempt (2026-09-13T02:57:26Z) after resolve-by-epic:

```text
OPEN: PASS dealRef=o_49885ba0-57ea-4fb6-98a6-fca5a3579df2
CONFIRM: PASS confirm dealId=00000000-6273-0eef-0448-411e0055311e
         size=0.0001 level=77299.40 status=OPEN
READ: PASS actual dealId=00000000-6273-0ef2-0448-411e0055311e
      BUY 0.0001 @ 77299.40
CLOSE: PASS dealRef=p_00000000-6273-0ef2-0448-411e0055311e
       close level=77249.40
FINAL: canary_present=false epic_open=0
```

Re-check 2026-09-13T03:04Z:

```text
POSITIONS: PASS open=0
DEMO balance ≈ 999.98 virtual USD
GOLD: still CLOSED
```

## Evidence captured

```text
epic BTCUSD
instrument type CRYPTOCURRENCIES
currency USD
market status TRADEABLE
bid/offer/spread observed at canary time
lotSize 1
contractSize 0
scalingFactor 0
min/max/increment from dealing rules
dealReference + dealId (confirm vs position id may differ)
requested size 0.0001
confirmed size 0.0001
open level 77299.40
close level 77249.40
```

No credentials or tokens were journaled.

## GOLD

GOLD remains the Sunday execution instrument. It was not substituted for the infrastructure proof.

```text
GOLD status: CLOSED (weekend)
monetary status: BROKER_METADATA_ONLY (lotSize=1 inferred MPU=1; not RUNTIME_VALIDATED)
Capital GOLD CFD != CME GC
```
