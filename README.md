# AurumFlow

Automated XAUUSD (gold) trading bot in Golang using institutional-style logic: **Liquidity → Structure → Intent → Equilibrium**. Integrated with the [Capital.com](https://capital.com) REST API. Terminal-only output.

## Principles

- **No prediction:** The bot reacts to structural probabilities, not forecasts.
- **Risk over winrate:** Strict risk management (risk per trade, max trades, daily drawdown limit).
- **Engine order:** Liquidity → Structure → Intent → Equilibrium; RSI/ATR are filters only.

## Requirements

- Go 1.21+
- Capital.com account (demo recommended first)
- 2FA enabled; API key from **Settings → API integrations → Generate new key**

## Configuration

1. Copy and edit `config/config.json`:
   - `api.mode`: `"demo"` or `"live"` — sets `api_base_url` automatically if empty; use `"demo"` for demo account
   - `api.api_base_url`: Optional; overridden by `mode` when mode is set. Demo: `https://demo-api-capital.backend-capital.com`, Live: `https://api-capital.backend-capital.com`
   - `api.api_key`: Your API key (from platform)
   - `api.identifier`: Your login email
   - `api.password`: The **custom password** you set for the API key (not your account password)
   - `api.account_id`: Optional; account to operate. If set, the bot switches to this account after login (use the id from the accounts list printed at startup)
   - `api.max_spread`: Optional; skip orders if spread exceeds this (0 = disabled)
   - `strategy.rsi_buy_low/high`, `rsi_sell_low/high`: Optional RSI zones (defaults 38–42, 58–62)
   - `strategy.trading_sessions`: Array of allowed trading sessions: `["LONDON"]`, `["NY"]`, `["ASIA"]`, `["LONDON", "NY"]`, or `["ALL"]` (default: `["ALL"]`). Sessions: London 08:00-17:00 UTC, NY 13:00-22:00 UTC, Asia 00:00-08:00 UTC. The bot only scans for signals during allowed sessions.

2. Environment variables (secrets and overrides):
   - `AURUMFLOW_CONFIG`: Path to config file (default: `config/config.json`). You can also pass it via `.\scripts\run.ps1 -Config <path>` or `./scripts/run.sh --config <path>` (path relative to project root or absolute).
   - `AURUMFLOW_EPIC`: Instrument epic (e.g. XAUUSD); if unset, resolved via search "gold". You can also pass it via `.\scripts\run.ps1 -Epic <epic>` or `./scripts/run.sh --epic <epic>`.
   - `AURUMFLOW_API_KEY`, `AURUMFLOW_IDENTIFIER`, `AURUMFLOW_PASSWORD`: Override API credentials (leave empty in config when using env)
   - `AURUMFLOW_ACCOUNT_ID`: Override account to operate
   - `AURUMFLOW_LIVE_CONFIRM`: Set to `1` or `true` to allow sending orders in **live** mode (required for live; bot skips orders otherwise)
   - `AURUMFLOW_LOG_JSON`: Set to `1` or `true` for JSON-structured log lines
   - `AURUMFLOW_TELEGRAM_BOT_TOKEN`, `AURUMFLOW_TELEGRAM_CHAT_ID`: Override Telegram bot token and channel/chat id when using Telegram notifications (do not commit tokens; use env in production)

### Notifications (Pushover and Telegram)

Notifications are optional. When `notifications.enabled` is true, the bot can send to **Pushover** (existing) and/or **Telegram** (channel or group).

- **Pushover:** Set `notifications.provider` to `"pushover"` and configure `notifications.pushover` with token and user (or use `AURUMFLOW_PUSHOVER_TOKEN`, `AURUMFLOW_PUSHOVER_USER`).
- **Telegram (public channel or group):**
  1. Create a bot with [BotFather](https://t.me/BotFather); copy the **bot token**.
  2. Create a channel or group; add the bot as **admin** (for channels, the bot must post).
  3. Get the **chat id**: for a public channel use `@channelname`, or use a numeric id (e.g. from a forward or getUpdates).
  4. In config, set `notifications.telegram.enabled` to `true`, `bot_token` and `chat_id` (or set `AURUMFLOW_TELEGRAM_BOT_TOKEN` and `AURUMFLOW_TELEGRAM_CHAT_ID` in the environment; do not commit the token).
  5. Optional: `notifications.telegram.events` — list of event types to send (e.g. `["SIGNAL_GENERATED","ORDER_SENT","POSITION_CLOSED"]`). If empty, all events that pass the notification policy are sent.

Example config block:

```json
"notifications": {
  "enabled": true,
  "provider": "pushover",
  "pushover": { "token": "...", "user": "..." },
  "telegram": {
    "enabled": true,
    "bot_token": "",
    "chat_id": "",
    "parse_mode": "HTML",
    "events": [],
    "include_market_context": true,
    "fail_policy": "best_effort",
    "rate_limit_per_min": 0
  },
  "queue_cap": 100
}
```

Telegram messages are best-effort (do not block trading). Messages are truncated to 4096 characters and formatted with safe HTML for the channel. Sensitive data (balance, exact size, account id) is omitted from Telegram messages for public channels.

## Build and run

### Scripts (recommended): install deps, test, build to `dist/`, run agent

**Windows (PowerShell)** — from project root:

```powershell
.\scripts\setup-and-run.ps1
```

Or step by step: `.\scripts\build.ps1` (deps + test + build to `dist/`), then `.\scripts\run.ps1` (runs the Windows binary from `dist/`).

**Linux / macOS** — from project root:

```bash
chmod +x scripts/*.sh
./scripts/setup-and-run.sh
```

Or step by step: `./scripts/build.sh` (deps + test + build to `dist/`), then `./scripts/run.sh` (runs the binary for current OS from `dist/`).

The build scripts produce binaries in **`dist/`**:

- `aurumflow-windows-amd64.exe`, `aurumflow-windows-arm64.exe`
- `aurumflow-linux-amd64`, `aurumflow-linux-arm64`
- `aurumflow-darwin-amd64`, `aurumflow-darwin-arm64`

Run with arguments (e.g. backtest): `.\scripts\run.ps1 --backtest candles.json` or `./scripts/run.sh --backtest candles.json`.

Optional **epic** and **config** (e.g. best config from backtesting per epic):

- **Windows:** `.\scripts\run.ps1 -Epic US100 -Config backtesting\gold\<run_id>\config.json`
- **Linux / macOS:** `./scripts/run.sh --epic US100 --config backtesting/gold/<run_id>/config.json` (short: `-e`, `-c`)

Config path can be relative to the project root or absolute. If omitted, the bot uses `config/config.json` and resolves epic from env or default (gold).

### Manual build and run

```bash
go build -o aurumflow ./cmd/bot
./aurumflow
```

Or from project root (config path relative to cwd):

```bash
go run ./cmd/bot
```

Run backtest from a candles file (JSON or CSV) and exit:

```bash
./aurumflow --backtest path/to/candles.json
```

The bot will:

1. Load config and create a Capital.com session
2. List all accounts (id and balance) and optionally switch to `api.account_id` if set
3. Resolve the gold epic (if not set)
4. Fetch candles for H1, M15, M5 and log to terminal
5. Start the event loop (every 60s): refresh balance/positions, structure → strategy → composer → state machine → risk → execution
6. Send orders only when state is READY and risk allows; in **live** mode, set `AURUMFLOW_LIVE_CONFIRM=1` to allow orders
7. On 401 (session expired), the bot re-logs in automatically; when all positions close, state moves to COOLDOWN

Stop with Ctrl+C.

## API limits (Capital.com)

- Session: 10 minutes; the bot pings every 5 minutes to keep it alive.
- Rate: Max 10 req/s; **1 req per 0.1 s** for `POST /positions` and `POST /workingorders` (enforced in execution layer).
- Demo: 1000 position/order creations per hour.

## Success criteria (from spec)

- Winrate target: 45–55%
- Profit factor > 1.2 (goal 1.3+)
- Max drawdown < 12%
- Consistency and risk control over high winrate

## API documentation

- [Capital.com Public API](https://open-api.capital.com/) — REST and WebSocket reference

## Project layout

- `cmd/bot/main.go` — Entrypoint, config, session, event loop
- `config/` — Config load (JSON/YAML)
- `internal/market/` — Capital.com REST client (session, prices, positions, markets)
- `internal/structure/` — Swing detection, market structure (HH/HL/LH/LL)
- `internal/strategy/` — Fibonacci, liquidity, intent, equilibrium, signal composer
- `internal/indicators/` — RSI, ATR
- `internal/risk/` — Position size, max trades, daily DD
- `internal/execution/` — POST positions, confirm, rate limit, terminal log
- `internal/logger/` — Structured logging (info/warn/error; optional JSON via env)
- `internal/core/` — State machine, event loop
- `internal/backtest/` — Replay candles, metrics (winrate, profit factor, max DD)
- `pkg/models/` — Candle, Swing, TradeSignal, etc.

## Disclaimer

Trading CFDs carries high risk. Use demo first. This bot is for educational purposes; past performance does not guarantee future results.
