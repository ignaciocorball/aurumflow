# 01 — Repository Inventory

**Status of this document:** evidence-based listing of what exists on disk and in git as of 2026-09-12.

## Git snapshot (before audit writes)

```text
COMMAND: git status / git branch -a / git log --oneline -20
RESULT:
  Branch: feature/claude-code (tracking origin/feature/claude-code)
  Also:   main, remotes/origin/main, remotes/origin/feature/claude-code
  HEAD:   4250f7b feat(1): initial implementation
  main == feature/claude-code (no commits either direction)
  Uncommitted (pre-audit): .gitignore, cmd/bot/main.go, config/*, go.mod/sum,
                           internal/backtest/backtest.go, internal/core/loop.go
  Untracked: internal/telemetry/*.go
  Remote: https://github.com/ignaciocorball/aurumflow.git
  Stash: empty  Tags: none
```

**All remote branches reviewed.** There is only one commit on both `main` and `feature/claude-code`. Divergence is **working-tree only** (Firebase telemetry + gitignore expansion).

---

## Dual reality: git vs disk

| Layer | In git | On disk (gitignored / untracked) |
|-------|--------|----------------------------------|
| Go trading bot | Yes | Same + uncommitted telemetry |
| `config/example_config.json` | Yes (mode=live) | Plus `config/config.json`, `config/bk_config.json` with secrets |
| `scripts/` | **No** (`/scripts` ignored) | `build.ps1/sh`, `run.ps1/sh`, `backtest_runner.py`, `analytics_extractor.py`, `firebase_rtdb.py` |
| `bots/` | **No** | 7 instance folders: xauusd, us100, us500, ethusd, eth-defensive-core, eth-offensive-alpha, msft |
| `backtesting/` | **No** | GOLD, US100, US500, US30, ETHUSD, MSFT, SILVER, NDAQ, MNQ1 + HTML/CSV/PNG reports |
| `docs/` | **No** (was `/docs`) | `RTDB_SCHEMA_V1.md`, `RTDB_EXAMPLE_EXPORT.json` |
| `service-account.json` | **No** | Present locally (Firebase SA) |
| Docker / CI | None | None |
| MQL5 / MT5 | None | None |
| TypeScript frontend | None | Schema-only for a planned React dashboard |

---

## Languages and versions

| Language | Evidence | Version / notes |
|----------|----------|-----------------|
| Go | `go.mod` `module aurumflow`, `go 1.24.0`, `toolchain go1.24.3` | Runtime: `go version go1.24.3 windows/amd64` |
| Python | `scripts/*.py` (gitignored) | Used for backtest orchestration + Firebase analytics. `python` not on PATH in this audit shell |
| PowerShell / Bash | `scripts/*.ps1`, `scripts/*.sh` | Local launcher |
| JSON / JSONL / CSV | `lab/`, `backtesting/`, journals | Data + research artifacts |
| Markdown | `README.md`, local `docs/`, `bots/*/README.md` | |
| HTML / PNG | `backtesting/*/report.html`, charts | Research reports, not an app |

**No:** Java, C#, TypeScript app, Jupyter notebooks, MQL5, Dockerfiles, GitHub Actions.

---

## Tracked Go package tree

```text
cmd/bot/main.go                 entrypoint (live loop + backtest flags)
config/config.go                load/validate JSON|YAML + env overrides
config/example_config.json      committed template (mode=live)
pkg/models/models.go            Candle, Swing, structure, Fib, signal
internal/market/                Capital.com REST client
  client.go                     HTTP + CST + X-SECURITY-TOKEN + rate limit
  session.go                    create/ping/get/switch session, accounts
  prices.go                     OHLC (bid)
  markets.go                    search epic + market details
  positions.go                  GET positions only
internal/structure/             swings + HH/HL classification
internal/strategy/              fib, liquidity, intent/sweep, equilibrium, sessions, H1 filter, composer
internal/indicators/            RSI (SMA + Wilder), ATR
internal/marketstate/           TrendFromCandles, M5 timing
internal/core/                  state machine + 60s event loop
internal/risk/                  size, max trades, daily DD
internal/execution/             POST positions + confirm
internal/journal/               JSONL writer
internal/logger/                console + optional file + JSON env
internal/notifications/         Pushover, Telegram, Redis dedup
internal/backtest/              candle replay
internal/telemetry/             UNTRACKED: Firebase RTDB live publisher
lab/candles.json                sample Capital.com price dump
lab/candles.toon                alternate serialization of lab candles
```

---

## Entry points

| Entrypoint | How | Status |
|------------|-----|--------|
| Live / demo bot | `go run ./cmd/bot` or `scripts/run.ps1` | `IMPLEMENTED_UNVERIFIED` |
| Offline backtest | `./aurumflow --backtest path.json` | `IMPLEMENTED_UNVERIFIED` |
| Download+backtest | `--backtest-from` `--backtest-to` (needs Capital session) | `IMPLEMENTED_UNVERIFIED` |
| Multi-epic research runner | `scripts/backtest_runner.py` (gitignored) | `IMPLEMENTED_UNVERIFIED` |
| Analytics → Firebase | `scripts/analytics_extractor.py`, `scripts/firebase_rtdb.py` | `IMPLEMENTED_UNVERIFIED` |

Single binary. No workers, no HTTP server, no gRPC, no message bus.

---

## File-type census (workspace, including ignored)

Approximate disk census from PowerShell `Group-Object Extension`:

| Ext | Count | Role |
|-----|------:|------|
| `.json` | 1224 | candles, run configs, summaries |
| `.log` | 661 | bot run logs |
| `.csv` | 526 | trade exports / leaderboards |
| `.jsonl` | 499 | journals |
| `.png` | 93 | research charts |
| `.go` | 57 | product source + tests |
| `.html` | 29 | backtest reports |
| `.md` | 10+ | README + local docs |
| `.py` | 3 | research scripts |
| `.ps1` / `.sh` | 9 / 9 | launchers |

---

## Tests present (git)

16 `*_test.go` files. Packages **without** tests: `cmd/bot`, `config`, `execution`, `journal`, `logger`, `market`, `risk`, `telemetry`, `pkg/models`.

---

## Hardcoded / notable constants

| Item | Location | Note |
|------|----------|------|
| Demo host | `config/config.go` | `https://demo-api-capital.backend-capital.com` |
| Live host | `config/config.go` | `https://api-capital.backend-capital.com` |
| Loop interval | `internal/core/loop.go` | 60s |
| Session ping | `loop.go` | 5 minutes |
| HTTP rate limit | `market/client.go` | 5 req/s, burst 2 |
| Order POST spacing | `execution/executor.go` | 100 ms |
| Default min size | `core.NewLoop` | 0.1 / 0.1 (ignores market details) |
| Value per point default | `config.Validate` | 1.0 |
| ATR buckets 8–18 | `strategy/composer.go` | gold-tuned, reused on indices |
| Session hours UTC | `strategy/sessions.go` | Asia 00–08, London 08–17, NY 13–22 |

---

## TODO / FIXME / HACK / stubs

Ripgrep over `*.go,*.py,*.md` in the **tracked** tree: **no TODO/FIXME/HACK**.

Functional stubs instead:

- `UsePartial` / `PartialR` / `RunnerR` — config only, never executed (`STUB`)
- `BlockH4Range` — config + comment, not a hard reject (`DEAD_CODE` / unused)
- `InFibZone`, `InFibZoneExtended` — never called (`DEAD_CODE`)
- `PhaseExpansion` — constant never assigned (`DEAD_CODE`)
- Notification types `REGIME_CHANGE`, `ORDER_FILLED`, `POSITION_MODIFIED`, `PNL_MILESTONE` — never emitted (`DEAD_CODE`)
- `SessionResponse.StreamingHost` — parsed, unused (`PARTIAL`)

---

## Duplications / abandoned prototypes

| Item | Assessment |
|------|------------|
| Live loop vs `backtest.Run` | Parallel gating (score + H4/H1 + sessions). Drift risk. `KEEP_AND_REFACTOR` |
| `config/bk_config.json` | Local backup of live secrets. `DEPRECATE` operationally |
| `bots/*` vs root `config/config.json` | Same engine, per-epic params. Local ops layer |
| `lab/candles.toon` | Alternate format of lab dump. Research leftover |
| Planned React dashboard | Schema in local `docs/RTDB_SCHEMA_V1.md` only. `NOT_IMPLEMENTED` |

---

## Prerequisites to reproduce the git product

- Go 1.24.3+
- Capital.com API key + identifier + API password
- `config/config.json` (not in git) or env `AURUMFLOW_API_KEY` / `AURUMFLOW_IDENTIFIER` / `AURUMFLOW_PASSWORD`
- No Docker, no DB server required for the core bot
- Optional: Redis (notification global dedup), Firebase SA (telemetry), Pushover/Telegram tokens
