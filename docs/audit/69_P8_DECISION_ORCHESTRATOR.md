# P8 Decision orchestrator

Still SHADOW. No broker access.

```
WorldState → Opportunity → Market analysis → Legacy → CrossAsset → Micro(optional)
        → DecisionProposal → ExecutionEligibility → PortfolioRisk
```

Proposal decision: `WATCH` · `SETUP` · `DEMO_CANDIDATE` · `BLOCKED`

Attention ≠ setup ≠ demo candidate ≠ order.

Microstructure unavailable does not hide a market. BTC L2 is not used as the book of GOLD / US100 / OIL.

`DemoExecutionOrchestrator` exists with `MULTI_MARKET_DEMO_EXECUTION=OFF`. It cannot mutate the broker.
