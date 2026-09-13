# 62 — P7 Global UI

Console default tab is GLOBAL (capital map, region cards, heatmap, opportunity radar, flows, session timeline, delayed institutional inspector).

MARKETS / MICRO / EXECUTION / RESEARCH keep the P6.1 observatory.

APIs (GET-only): `/api/world`, `/api/regions`, `/api/assets`, `/api/opportunities`, `/api/context/sources`, `/api/institutional`.

WorldState is fetched at 15s cadence, not stuffed into the 500ms SSE payload.
