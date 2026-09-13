# 57 — P7 Global Capital Architecture

Hierarchy is now WORLD → REGION → ASSET CLASS → MARKET → MICROSTRUCTURE → DECISION → EXECUTION.

BTC is MICROSTRUCTURE_LAB + 24/7 SENSOR, not the primary global asset. GOLD remains the only DEMO execution pilot.

New packages: `worlddomain`, `globalsources`, `worldstate`, `crossasset`, `opportunity`, `scanner`, `orchestrator`, `sessions`.

No auto-execution from WorldState. DecisionOrchestrator is SHADOW-only and cannot embed ExecutionProvider.
