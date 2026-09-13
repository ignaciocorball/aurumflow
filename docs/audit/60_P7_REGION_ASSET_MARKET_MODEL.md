# 60 — P7 Region / Asset / Market Model

Regions: GLOBAL, UNITED_STATES, EUROPE, JAPAN, CHINA_HONG_KONG, ASIA_EX_JAPAN.

Asset classes keep precious metals and energy separate from a generic commodity bucket.

Canonical markets: GOLD, SILVER, OIL, US100, US500, US30, EUROPE, UK, JAPAN, HK, CHINA, BTC.

Capital epics are persisted only after `SearchMarkets` returns them. Eligibility stays ANALYSIS_ONLY or DEMO_NOT_CALIBRATED. P7 never auto-promotes DEMO_ELIGIBLE.
