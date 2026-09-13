# 15 — Security and Secrets

**No secret values are printed in this audit.**

## Git index

```text
COMMAND: git ls-files
RESULT: config/config.json NOT tracked
        service-account.json NOT tracked
        .env NOT present
        example_config.json tracked with empty api_key/identifier/password
PASS (no secrets in tracked files at HEAD)
```

Uncommitted `.gitignore` adds `/config/bk_config.json`, `/docs`, `service-account.json`.

This audit added exceptions so `docs/audit/**` can be tracked.

## Local secret inventory (presence only)

| Location | Variable / field | Provider | Configured? |
|----------|------------------|----------|-------------|
| `config/config.json` | `api.api_key` | Capital.com | Yes |
| same | `api.identifier` | Capital.com login | Yes |
| same | `api.password` | Capital.com **API** password | Yes |
| same | `api.account_id` | Capital.com | Yes |
| same | `api.mode` / `api_base_url` | Capital.com **LIVE** | Yes |
| same | `notifications.pushover.token` / `user` | Pushover | Yes |
| same | `notifications.telegram.bot_token` | Telegram | Yes |
| `config/bk_config.json` | same API fields | Capital.com LIVE | Yes |
| `bots/*/config.json` (7 bots) | API key/identifier/password/account | Capital.com LIVE | Yes |
| `service-account.json` | GCP service account JSON | Firebase / Google | Yes (file ~2385 bytes) |
| `.env` | — | — | Missing |
| Process env `AURUMFLOW_*` | API / Telegram / Pushover | — | All unset in audit shell |
| `FIREBASE_SERVICE_ACCOUNT_JSON` / `GOOGLE_APPLICATION_CREDENTIALS` | Firebase | — | Unset |

Bot READMEs claim credentials are **not** in `config.json` and should come from env. **Inspection contradicts the READMEs.**

## Env contract (code)

| Env | Purpose | In git? |
|-----|---------|---------|
| `AURUMFLOW_CONFIG` | Config path | Name only |
| `AURUMFLOW_EPIC` | Instrument | Name only |
| `AURUMFLOW_API_KEY` | Capital key | Override if config empty |
| `AURUMFLOW_IDENTIFIER` | Login | same |
| `AURUMFLOW_PASSWORD` | API password | same |
| `AURUMFLOW_ACCOUNT_ID` | Account switch | same |
| `AURUMFLOW_LIVE_CONFIRM` | Allow LIVE orders | **Required only if mode=live** |
| `AURUMFLOW_LOG_JSON` / `LOG_DIR` / `LOG_COLOR` | Logging | |
| `AURUMFLOW_TELEGRAM_*` / `AURUMFLOW_PUSHOVER_*` | Notif secrets | |
| `AURUMFLOW_BLOCK_ON_DATA_STALE` | Exit on M5 stale at startup | |
| `FIREBASE_SERVICE_ACCOUNT_JSON` | SA JSON string | |
| `GOOGLE_APPLICATION_CREDENTIALS` | SA path | |

`ApplyEnvOverrides` fills secrets **only when config field is empty**. A file with live keys wins over a safer env.

## Live trading exposure

| Vector | Status |
|--------|--------|
| Example template | LIVE host |
| Local default config | LIVE + secrets |
| Seven bot instances | LIVE + secrets |
| Confirm flag | Unset now — blocks **orders** if mode stays live; does **not** block login |
| Mode bypass | Empty mode + live URL → orders without confirm (`BROKEN`) |
| Multi-bot | Parallel LIVE processes possible via `bots/*/run.ps1` |

**Live trading exposure: HIGH on this workstation. This audit did not touch the broker.**

## Other findings

| Finding | Severity | Notes |
|---------|----------|-------|
| Telegram to public channel | P2 | README omits size/balance — good; token still powerful |
| Firebase SA on disk | P1 | Full project credentials if file is a standard SA |
| Logs may contain account ids / balances | P2 | `logger.Info("account[%d] id=%s balance=...")` |
| Journal may contain deal refs | P3 | Expected |
| No secret scanner / pre-commit | P2 | Single commit repo |
| `/scripts` ignored | P2 | Launchers that set epic/config not reviewable in git |

## Recommendations (P1, not done)

1. Rotate Capital.com API keys if these files were ever copied or backed up unsafely.
2. Demo-only committed example.
3. Env-only secrets; refuse to start if secrets are inline and `mode=live`.
4. Host allow-list: refuse non-demo unless `LIVE_CONFIRM` **and** `mode=live`.
5. Keep `service-account.json` ignored (already).
