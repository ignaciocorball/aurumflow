# 19 — P1 DEMO Validation

## Safety / unit baseline (this machine)

```text
go test ./...     PASS   (117 tests; was 66)
go vet ./...      PASS
go build ./cmd/bot PASS
```

No Capital.com LIVE calls were made.

## DEMO credential search

Required for runtime canary (explicit DEMO, not LIVE reuse):

```text
AURUMFLOW_DEMO_API_KEY
AURUMFLOW_DEMO_IDENTIFIER
AURUMFLOW_DEMO_PASSWORD
optional: AURUMFLOW_DEMO_ACCOUNT_ID
```

or a dedicated `config/demo_config.json` (gitignored).

| Probe | Result |
|-------|--------|
| `AURUMFLOW_DEMO_*` env | unset |
| `config/demo_config.json` | missing |
| `config/config.json` | LIVE (ignored; not reused, not pointed at demo host) |

```text
DEMO_RUNTIME = BLOCKED_MISSING_DEMO_CREDENTIALS
```

P1 does **not** copy LIVE keys onto the DEMO host.

## Canary results

```text
DEMO AUTH:            NOT RUN
ACCOUNT:              NOT RUN
MARKET DATA:          NOT RUN
MARKET DETAILS:       NOT RUN
POSITIONS READ:       NOT RUN

Open / Confirm / Close canary: NOT RUN
Final open positions: N/A
```

When DEMO secrets exist:

```text
go run ./cmd/bot --canary-readonly
```

Then, only after that passes, a single controlled open/close can be issued from a DEMO `execution_mode=demo` process (not implemented as an automatic second canary in this freeze because credentials are absent).

## How to unblock runtime (operator)

1. Create a Capital.com **DEMO** API key (separate from LIVE).
2. Export `AURUMFLOW_DEMO_*` or write `config/demo_config.json` with `environment=demo`, `execution_mode=disabled`.
3. Run `--canary-readonly`.
4. Only then set `execution_mode=demo` for a single GOLD instance.
