# P1 — Local corpus inventory (gitignored)

Do not add these trees to git without a secrets review.

| Path | Classification | Notes |
|------|----------------|-------|
| `config/config.json` | CONTAINS_SECRETS | LIVE Capital + notif tokens |
| `config/bk_config.json` | CONTAINS_SECRETS | LIVE backup |
| `bots/*/config.json` | CONTAINS_SECRETS | LIVE; seven instances |
| `bots/*/README.md` | SAFE_TO_VERSION | After stripping any pasted secrets (currently instructional) |
| `bots/*/run.ps1` `run.sh` | SAFE_TO_VERSION | Launchers; cannot bypass P1 lock |
| `scripts/build.ps1` `build.sh` `run.ps1` `run.sh` | SAFE_TO_VERSION | Build/run wrappers |
| `scripts/backtest_runner.py` | SAFE_TO_VERSION | Research orchestrator |
| `scripts/analytics_extractor.py` | SAFE_TO_VERSION | |
| `scripts/firebase_rtdb.py` | SAFE_TO_VERSION | No SA in file; reads env/path |
| `service-account.json` | CONTAINS_SECRETS | Firebase SA |
| `backtesting/**/*.json` candles | RESEARCH_DATA | Capital OHLC dumps |
| `backtesting/**/config.json` | UNKNOWN | May copy strategy only; inspect before versioning |
| `backtesting/**/*.csv` `*.jsonl` `*.html` `*.png` | GENERATED_DATA | Reports |
| `logs/` | GENERATED_DATA | May contain account ids |
| `journals/` | GENERATED_DATA | Trade journal |
| `docs/RTDB_SCHEMA_V1.md` | SAFE_TO_VERSION | Schema, no secrets |
| `docs/RTDB_EXAMPLE_EXPORT.json` | UNKNOWN | Review before versioning |
| `docs.zip` | UNKNOWN | Do not commit blindly |
| `lab/candles.json` | RESEARCH_DATA | Already in git |

Classification is inventory-only. Nothing was deleted.
