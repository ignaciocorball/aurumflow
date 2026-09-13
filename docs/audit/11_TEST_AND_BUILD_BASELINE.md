# 11 — Test and Build Baseline

**Policy:** measure, do not mass-fix. No production API calls.

## Environment

```text
COMMAND: go version
RESULT:  go version go1.24.3 windows/amd64
PASS
```

```text
COMMAND: git branch / log
RESULT:  feature/claude-code @ 4250f7b; main identical
PASS (repo readable)
```

## Unit tests

```text
COMMAND: go test ./... -count=1
RESULT:  exit 0  (~51s first run)
PASS packages:
  aurumflow/internal/backtest
  aurumflow/internal/core
  aurumflow/internal/indicators
  aurumflow/internal/marketstate
  aurumflow/internal/notifications
  aurumflow/internal/strategy
  aurumflow/internal/structure
NO TEST FILES:
  cmd/bot, config, execution, journal, logger, market, risk, telemetry, pkg/models
FAIL: none
```

```text
COMMAND: go test ./... -count=1 -json  (count Action=pass with Test)
RESULT:  pass_tests=66  fail_tests=0  skip=0
PASS
```

Coverage is **function-level unit tests**, not broker integration. `backtest.Run` test does not assert strategy quality.

## Vet / build

```text
COMMAND: go vet ./...
RESULT:  empty output, exit 0
PASS
```

```text
COMMAND: go build -o dist/aurumflow-audit.exe ./cmd/bot
RESULT:  exit 0
PASS
```

`dist/` is gitignored.

## Lint / typecheck / static analysis (non-Go)

| Tool | Result |
|------|--------|
| golangci-lint | Not configured / not run (`UNKNOWN`) |
| ESLint / tsc | N/A (no TS app) |
| Python lint | Scripts exist; `python` not on PATH (`BLOCKED`) |
| Docker build | No Dockerfile (`NOT_IMPLEMENTED`) |
| CI | No `.github/workflows` (`NOT_IMPLEMENTED`) |

## Integration / demo

```text
COMMAND: (none — Capital.com session)
RESULT:  SKIPPED
BLOCKED: all local configs mode=live with credentials; audit policy forbids LIVE calls
```

## Severity of gaps

| Gap | Severity | Impact |
|-----|----------|--------|
| No `internal/market` tests | P1 HIGH | Auth/orders can regress silently |
| No `internal/execution` tests | P1 HIGH | SL/TP payload unchecked |
| No `internal/risk` tests | P1 HIGH | Size/DD math unchecked |
| No `config.Validate` tests | P0 CRITICAL | Live-confirm bypass untested |
| Backtest test is a smoke test | P2 MEDIUM | Replay can drift from live |
| No CI | P2 MEDIUM | Baseline is laptop-only |

## Reproducible path (git product)

```text
install:  go 1.24.3+ ; go mod download
configure: copy example → config/config.json ; set DEMO host and env secrets
           (do NOT copy current local live files)
build:    go build -o aurumflow ./cmd/bot
          or (local) scripts/build.ps1 — gitignored
run:      AURUMFLOW_CONFIG=... AURUMFLOW_EPIC=GOLD ./aurumflow
test:     go test ./...
```

Scripts documented in README (`setup-and-run.ps1`) are **not in git**.

## Warnings

- First `go test` pulled modules including Firebase (uncommitted `go.mod` requires `firebase.google.com/go/v4`).
- Working tree must be considered when reproducing HEAD vs dirty tree.
