# 04 — Strategy Inventory

There is **one strategy**, not a catalog of interchangeable strategies. It is a scored pipeline:

**Liquidity → Structure → Intent (sweep) → Equilibrium → RSI/ATR filters → H4/H1/session context → State machine.**

README name: AurumFlow institutional-style logic.

---

## Master strategy: AurumFlow Composer

```text
Nombre:           AurumFlow Signal Composer
Ubicación:        internal/strategy/composer.go
                  BuildComposerInput(), SignalComposerWithZones()
Instrumentos:     Any Capital.com epic (default gold). Local ops: GOLD, US100, US500, ETHUSD, MSFT, SILVER, NDAQ, US30, MNQ1
Timeframes:       M15 setup (required); H1 bias (optional); H4 regime score/readiness; M5 timing (optional)
Inputs:           []models.Candle + config (RSI/ATR/sweep/score/sessions)
Outputs:          *models.TradeSignal {Direction, Entry, StopLoss, TakeProfit, Confidence, Score} or reject
Estado:           IMPLEMENTED_UNVERIFIED (unit tests exist; live edge unverified)
Dependencias:     structure, indicators, marketstate, config
¿Tiene tests?:    Sí — composer_test.go, plus component tests
¿Tiene backtest?: Sí — internal/backtest.Run uses the same composer
¿Puede emitir señal?: Sí
¿Puede ejecutar trade?: Sí, if Loop state READY + risk + (not live-blocked)
¿Conectada con risk?: Sí — Loop calls Risk.ValidateSignal then PositionSize
¿Conectada con broker?: Sí — Loop → Executor.SendOrder
```

Scoring (composer):

| Condition | Points |
|-----------|--------|
| LiquidityEvent.Strength > 0 | +3 |
| Direction aligned with structure trend | +2 |
| Equilibrium touch | +2 |
| ATR in [min,max] | +1 |
| ATR in [12,18] | +1 extra (gold-tuned) |
| ATR in [8,12) | −1 |
| RSI BOOST in zone | +1 |
| RSI PENALTY in zone / against | +2 / −1 |
| RSI GATE | reject if out of zone |

Direction only if structure trend matches sweep side (BUY_SIDE+BULLISH or SELL_SIDE+BEARISH).

Entry = last close. SL/TP from Fib+ATR or ATR 1R/2R, then sanity gate min RR 1.5.

H4/session added **after** compose in `loop.go` / `backtest.go` (`finalScore = score + h4Score + sessionScore`). NY +1, London −1, H4 aligned +2 / RANGE −1 / opposed −2.

---

## Component inventory

### Liquidity analyzer

```text
Nombre: LiquidityAnalyzer
Ubicación: internal/strategy/liquidity.go
Instrumentos: candle-generic
Timeframes: applied on M15 (and any slice passed in)
Inputs: candles, ATR, lookback
Outputs: []models.LiquidityZone {Price, Type=BUY_LIQUIDITY|SELL_LIQUIDITY, Strength}
Estado: IMPLEMENTED_UNVERIFIED
Dependencias: models
Tests: indirect via composer; no dedicated liquidity_test.go
Backtest: yes (via BuildComposerInput)
Señal: no (feature)
Trade: no
Risk/broker: no
```

Detects long wicks vs body/ATR. Caps to last 10 zones. **Not** order-book liquidity.

### Intent / sweep detector

```text
Nombre: IntentDetector
Ubicación: internal/strategy/intent.go
Inputs: last candle, zones, ATR, SweepOptions
Outputs: sweepBuy, sweepSell, strength, SweepDiagnostics
Estado: IMPLEMENTED_UNVERIFIED
Tests: no dedicated file; HasSweep tested
```

Institutional candle rule: broke zone + wick ≥ minPenetration·ATR + wick/body ≥ minWickRatio + close back inside tolerance.

Fallback without zones: wick-only.

### Equilibrium

```text
Nombre: EquilibriumTouch / EquilibriumCalculator
Ubicación: internal/strategy/equilibrium.go
Inputs: last close, last swing high/low, tolerance 0.5%
Outputs: bool (price near midpoint ≈ 50% of last impulse)
Estado: IMPLEMENTED_UNVERIFIED
Tests: none dedicated
```

### Fibonacci

```text
Nombre: CalculateFib
Ubicación: internal/strategy/fibonacci.go
Inputs: swings
Outputs: models.Fibonacci levels 0.382/0.5/0.618/0.705 + ImpulseDirection
Estado: PARTIAL
Tests: fibonacci_test.go
Used for: SL/TP when FibValid and impulse direction matches
DEAD_CODE: InFibZone, InFibZoneExtended — never called
```

### Market structure / swings

```text
Nombre: DetectSwings + BuildStructure
Ubicación: internal/structure/swings.go, structure.go
Outputs: pivots; Trend BULLISH|BEARISH|RANGE via HH/HL or LH/LL
Estado: IMPLEMENTED_UNVERIFIED
Tests: swings_test.go, structure_test.go
```

`StructureState.Phase` = IMPULSE | PULLBACK. Constant `EXPANSION` never assigned (`DEAD_CODE`).

No named BOS / CHoCH / FVG / order block / Elliott functions exist (ripgrep: no matches).

### H1 filter

```text
Nombre: H1Allowed / Loop H1 block
Ubicación: internal/strategy/h1filter.go + loop.go
Estado: IMPLEMENTED_UNVERIFIED
Tests: h1filter_test.go
```

BUY only if H1 BULLISH; SELL only if H1 BEARISH. Optional RANGE block. Crypto override by extra score.

### H4 “filter”

```text
Nombre: UseH4Filter
Ubicación: config + loop fetch + score; backtest same
Estado: PARTIAL
```

Config comment says “BUY only when H4 BULLISH”. **Live/backtest implementation uses H4 as score**, except live **blocks evaluation** if fewer than 60 H4 candles. `BlockH4Range` is **not applied as a hard reject** (`DEAD_CODE` / unused flag).

### Sessions

```text
Nombre: CanTrade / GetSessionInfo / NextSessionStart
Ubicación: internal/strategy/sessions.go
Hours UTC: Asia 00–08, London 08–17, NY 13–22
Overlap label: LONDON+NY (13–17)
Estado: VERIFIED_WORKING (unit tests)
Tests: sessions_test.go
```

Loop can hard-skip London+NY overlap (`block_london_ny_overlap`, default true if unset). Example config sets it **false**.

### M5 refiner

```text
Nombre: M5TimingAndScore
Ubicación: internal/marketstate/engine.go
Estado: IMPLEMENTED_UNVERIFIED
Tests: engine_test.go
```

In READY only: opposed M5 trend → recycle to WAIT_PULLBACK. Degrades to M15-only if stale.

### Indicators

```text
Nombre: RSI / RSIWilder / ATR / ATRValid / InBuyZone / InSellZone
Ubicación: internal/indicators
Estado: VERIFIED_WORKING (unit tests)
Tests: rsi_test.go, atr_test.go
```

---

## Concept search (requested terms)

| Concept | Present? | Where | Status |
|---------|----------|-------|--------|
| Liquidity | Yes | candle zones, not book | `PARTIAL` |
| Sweep | Yes | IntentDetector | `IMPLEMENTED_UNVERIFIED` |
| BOS | No named | implied via HH/HL | `NOT_IMPLEMENTED` |
| CHoCH | No | — | `NOT_IMPLEMENTED` |
| FVG / imbalance | No | — | `NOT_IMPLEMENTED` |
| Fibonacci | Yes | SL/TP + unused zone helpers | `PARTIAL` |
| Elliott | No | — | `NOT_IMPLEMENTED` |
| Pivots | Yes | swing pivots | `IMPLEMENTED_UNVERIFIED` |
| Volume | Field only | copied from API, unused in score | `STUB` |
| ATR / volatility | Yes | filter + SL/TP + buckets | `VERIFIED_WORKING` |
| Sessions Asia/London/NY | Yes | sessions.go | `VERIFIED_WORKING` |
| Institutional | Marketing/README | candle heuristics | `PARTIAL` |
| Order flow | No | — | `NOT_IMPLEMENTED` |
| Spread | Optional gate | `api.max_spread` (0 = off) | `PARTIAL` |
| Position sizing | Yes | `risk.PositionSize` | `IMPLEMENTED_UNVERIFIED` |
| Drawdown | Yes | daily DD % | `PARTIAL` |
| Trailing | Config UsePartial unused; BE in backtest only | | `STUB` / `PARTIAL` |
| Stop loss / take profit | Yes on send | no update | `PARTIAL` |

---

## Future-state mapping (only with foundation)

| Future state | Existing analogue | Foundation? |
|--------------|-------------------|-------------|
| ACCUMULATION | None | Do not map |
| DISTRIBUTION | None | Do not map |
| ABSORPTION | None (no book) | Do not map |
| EXPANSION | Constant `PhaseExpansion` unused; `IMPULSE` when price beyond midpoint | Weak / do not equate |
| NO_TRADE | IDLE, OFF_HOURS, MARKET_CLOSED, DATA_FROZEN, rejects, COOLDOWN | Yes — operational, not regime |

---

## Can it emit and trade?

Yes, in process: composer → READY → risk → `POST /positions`.

There is no paper-trading / dry-run strategy adapter. Backtest emits simulated fills only.
