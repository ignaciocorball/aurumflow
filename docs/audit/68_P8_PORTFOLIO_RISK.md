# P8 Portfolio risk

`PortfolioRiskManager` exists before any multi-market DEMO execution.

Conservative DEMO caps (not optimized):

- max open positions: 3
- max correlated positions: 1
- max aggregate risk: 300
- max risk per region: 150
- max risk per asset class: 150

Correlation groups: `US_EQUITY_INDEX` · `PRECIOUS_METALS` · `ENERGY` · `EUROPE_EQUITY` · `ASIA_EQUITY` · `CRYPTO`

US100 + US500 + US30 share one US equity group. Unknown monetary risk blocks a new order.
