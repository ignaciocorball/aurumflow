# 26 — P3 Radar MVP

```text
RADAR_MODE = SHADOW | OFF
AUTO_EXECUTE = not implemented
```

Radar may observe, score, classify, journal, and alert. It may not open, update, or close Capital positions.

`radar.MayMutateBroker` is always false.

## PressureSnapshot

Components are additive and visible:

```text
AggressiveFlow
Absorption          (POTENTIAL_ABSORPTION only)
BookImbalance
Liquidity
Persistence
Volatility
------------------
PressureScore       clamped -100..+100
Direction           from score
Confidence          separate (feed, sync, samples, freshness)
State               NO_TRADE | ABSORPTION | EXPANSION
```

Unsynced book forces `NO_TRADE` and reduces confidence.

State changes use hysteresis / minimum hold.

## Soak

Command:

```text
go run ./cmd/bot --shadow-soak --soak-minutes 60
```

Soaks in this environment:

```text
3m  REST snapshot only (WSS silent): events=1 synced=true mutations=0
1m  REST aggTrade+depth fallback: events=370 trades=310 snapshots=60
    synced=true cvd updates pressure updates mutations=0 no panic
```

WSS combined stream is implemented; this workstation appears to allow HTTPS REST to `fapi.binance.com` but not a live `fstream` event flow. REST poll keeps SHADOW features alive without credentials. Metrics: `journals/soak-metrics.json`.

Use `--shadow-soak --soak-minutes 60` for the longer Sunday soak.
