# 51 — P5.4 Multi-Venue Provenance

Canonical identity is never collapsed:

```text
Binance BTCUSDT USD-M
  != OKX BTC-USDT-SWAP
  != Capital BTCUSD CFD
```

Relationships:

```text
CORRELATED_PROXY
never IDENTICAL
```

## Frozen V1 input

`FLOW_EXHAUSTION_V1` remains Binance USD-M `aggTrade` semantics.

```text
V1FlowProvider = binance_usdm_public
```

If L2 is OKX:

```text
L2Provider     = okx_swap_public
L2FlowProvider = okx_swap_public
relation       = CORRELATED_PROXY
```

OKX trades are same-venue auxiliary flow for OKX book response. They do not replace Binance V1.

## Clock

All events normalize to UTC. Store provider event time and local receive time. Latency and clock-skew estimates are descriptive. No future-information alignment.

## Basis (descriptive, not arb)

When both mids exist:

```text
basis = mid_OKX - mid_Binance
z     = past-only robust z-score
```

Purpose: detect when the OKX book is too dislocated to be a useful proxy for Binance V1.

## L2ProxyQuality

Operational thresholds only (not fit to returns):

```text
GOOD      synced, fresh, healthy, |z| modest
DEGRADED  aging book, elevated latency, moderate basis
UNUSABLE  unsynced, stale, provider down, extreme basis/skew
```
