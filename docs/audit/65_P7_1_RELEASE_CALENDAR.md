# P7.1 Release calendar

`ReleaseCalendar` tracks expected freshness for slow sources. It is not a news-trading trigger.

| Source | Cadence | Last actual (this run) | Next expected |
| --- | --- | --- | --- |
| FED H.4.1 | WEEKLY | 2026-09-10 | next Thursday |
| CFTC | WEEKLY | 2026-09-08 | next Thursday |
| ICI | WEEKLY | unknown (403) | next Thursday |
| TIC | MONTHLY | 2026-06 | **2026-09-16** (July data) |
| JPX | WEEKLY | 2026-09-01 | next Thursday |

TIC is not polled every 60 seconds. Official fetch runs on snapshot / 6-hour live tick. Cache-only WorldState refresh may run every 60s without re-downloading monthly files.
