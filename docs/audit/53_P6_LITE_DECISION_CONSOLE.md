# 53 — P6-lite Decision Console

One HTML page on the existing Go status server. No React. No mutation buttons.

```text
URL:  http://127.0.0.1:8766/
Bind: 127.0.0.1
```

Ports:

```text
8765  GOLD DEMO
8766  SHADOW intelligence + console
```

## Safety bar

Always visible:

```text
CAPITAL DEMO
LIVE IMPOSSIBLE / FAIL-CLOSED
EXECUTION MODE
KILL SWITCH
OPEN POSITIONS
DEMO BALANCE
RADAR / EXHAUSTION / ABSORPTION = SHADOW
```

No BUY / SELL / CLOSE / FLATTEN. Non-GET methods return 405. No CST, X-SECURITY-TOKEN, API keys, or passwords.

## APIs

```text
/status
/api/status
/api/gold
/api/intelligence
/api/book
/api/prospective
/api/research
/api/stream   SSE ~500ms aggregated snapshots
```

UI consumes aggregated status, not raw book deltas.

## Why panel

Deterministic prose from feature values. No LLM.
