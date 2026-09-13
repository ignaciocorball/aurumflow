# 16 — Roadmap (derived from this codebase)

Not a generic platform wishlist. Each phase exits when **this repo** can prove the criterion.

---

## P0 — Discovery / Freeze (this iteration)

**Objective:** Know what AurumFlow actually is.

**Exit criteria**

| Criterion | Met? |
|-----------|------|
| Architecture documented | Yes — `02_CURRENT_ARCHITECTURE.md` |
| Capital.com audited | Yes at code level; runtime DEMO `UNKNOWN` |
| Execution audited | Yes |
| Risk audited | Yes |
| Test baseline | Yes — 66/66, vet, build |
| Secrets audited | Yes — local LIVE exposure documented |
| Strategies catalogued | Yes — single composer |
| Build reproducible | Yes for git Go tree; scripts not in git |

**Status:** complete with warnings (`17_P0_FREEZE_REPORT.md`).

---

## P1 — Execution Foundation (DEMO)

**Objective:** A platform that can be trusted on **Capital.com DEMO** without opening LIVE.

Derived gaps: live default, confirm bypass, no close, no modes, no market tests, journal off, wrong account balance.

**Exit criteria (adapt to reality)**

```text
Dedicated demo config (mode=demo, demo host) — no live keys in example
CreateSession + GetAccounts + GetPrices proven against DEMO (read-only first)
GetPositions proven
Optional: one documented min-size DEMO open+close (only after close API exists)
Risk gates unit-tested (size, max, DD, account id)
Kill switch: halt new orders; later flatten
Trade journal on by default in demo
AURUMFLOW_LIVE_CONFIRM required for any live host, regardless of mode field
EXECUTION_MODE in {DISABLED, DRY_RUN, DEMO} — LIVE still out of scope
Deterministic httptest suite for market + execution
Loop uses current account balance, not accounts[0]
```

**Not in P1:** Radar score, CME, options, dashboard rewrite, Kafka.

---

## P2 — Market Data Foundation

**Objective:** Separate intelligence from execution.

Add interfaces (names indicative, keep them small):

```text
MarketDataProvider      // candles + snapshot
ExecutionProvider       // open/close/update
HistoricalDataProvider  // files + Capital download
```

Capital.com becomes **one** implementation. Instrument record: epic, asset class, value_per_point, min size, CFD vs future flag.

Prepare slots for later CME/Nasdaq/options adapters — do not implement them until a feed is chosen.

---

## P3 — Institutional Radar MVP

**Objective:** Observe **NQ and GC** (or honest CFD proxies labeled as such) with:

```text
order flow, absorption, liquidity, session context, regime, alerts
```

Human decision only. Reuse: sessions, journal, notifications, candle structure as *secondary* features.

**Does not exist today.** Requires new data.

---

## P4 — Options + Cross Asset

QQQ/SPX options, GEX, dealer pressure, DXY, yields, VIX.

**Does not exist today.**

---

## P5 — Historical Replay / Research

Promote `internal/backtest` + local Python runner to event-level replay.

Measure precision/recall, MAE/MFE, false positives, regime slices.

Do not promote PF from current GOLD/US100 HTML reports to “edge.”

---

## P6 — Human Decision Console

Build the dashboard the RTDB schema already describes. Consume live status + alerts. No auto-trade.

---

## P7 — Controlled Automation

Only after measured edge:

```text
DISABLED → DRY_RUN → DEMO → (LIVE later, out of initial scope)
```

LIVE remains a conscious, separate program — never the example default.

---

## Suggested sequencing after this freeze

1. Human accepts P0 freeze.
2. P1 only: demo-hard, tests, close/update, modes, journal, kill switch.
3. Then decide whether Radar (P3) or instrument-correct CFD research (P2) is the next product bet.

Minimum path to a **DEMO bot that is honest**: P1.  
Minimum path to **Institutional Radar**: P1 → P2 → P3 (new data). The current composer is a feature source, not the Radar.
