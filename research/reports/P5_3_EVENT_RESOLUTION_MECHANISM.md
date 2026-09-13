# P5.3 Event-resolution mechanism

EXPLORATORY_MECHANISM_ONLY. V1 membership unchanged.

Runtime 5m48.5788851s events=243352328 ev/s=698127 mem=4.4MB

## DISCOVERY (2026-06-15→2026-09-11) conclusion=YES
### EXHAUSTION
- 1s n=56 mean=26.6964 median=4.0000 p25=2.0000 p75=11.0000
- 5s n=56 mean=23.3321 median=7.4000 p25=3.0000 p75=25.0000
- 15s n=56 mean=17.2714 median=8.9333 p25=3.4000 p75=18.4667
- 30s n=56 mean=18.0298 median=9.9667 p25=4.3667 p75=21.0333
- 1m n=56 mean=18.7935 median=9.4000 p25=5.0500 p75=18.2667
- 5m n=56 mean=20.0098 median=9.3433 p25=5.2100 p75=18.7367
### CONTINUATION
- 1s n=33 mean=7.6970 median=3.0000 p25=2.0000 p75=6.0000
- 5s n=33 mean=8.6242 median=4.2000 p25=2.4000 p75=6.6000
- 15s n=33 mean=11.6283 median=4.3333 p25=2.7333 p75=10.4000
- 30s n=33 mean=11.3818 median=6.0667 p25=3.3333 p75=14.8000
- 1m n=33 mean=13.2712 median=7.8500 p25=4.2000 p75=19.3667
- 5m n=33 mean=12.6542 median=11.2233 p25=6.4333 p75=15.3067
### NEUTRAL
- 1s n=839 mean=9.2038 median=3.0000 p25=2.0000 p75=6.0000
- 5s n=839 mean=9.2074 median=4.2000 p25=2.2000 p75=11.6000
- 15s n=839 mean=9.5107 median=6.1333 p25=2.8667 p75=11.4000
- 30s n=839 mean=9.3544 median=6.1000 p25=3.4000 p75=11.5000
- 1m n=839 mean=10.1516 median=6.6333 p25=3.8000 p75=11.8833
- 5m n=839 mean=11.2175 median=7.5933 p25=4.6700 p75=13.2633
Matched tps effects map[1m:5.221383647798742 1s:19.30188679245283 30s:7.106289308176102 5s:15.864150943396227]

## EXTERNAL_HOLDOUT (2026-03-17→2026-06-14) conclusion=YES
### EXHAUSTION
- 1s n=50 mean=17.4600 median=5.0000 p25=3.0000 p75=21.0000
- 5s n=50 mean=27.7320 median=12.4000 p25=4.0000 p75=33.6000
- 15s n=50 mean=22.1653 median=12.6667 p25=5.0000 p75=27.8667
- 30s n=50 mean=19.4520 median=10.7333 p25=5.6333 p75=21.1000
- 1m n=50 mean=18.6400 median=11.0167 p25=5.5000 p75=20.8500
- 5m n=50 mean=22.1847 median=10.7500 p25=6.4533 p75=20.6433
### CONTINUATION
- 1s n=41 mean=32.5122 median=5.0000 p25=3.0000 p75=15.0000
- 5s n=41 mean=32.4000 median=7.2000 p25=4.2000 p75=23.2000
- 15s n=41 mean=22.6715 median=10.4000 p25=5.7333 p75=22.6000
- 30s n=41 mean=19.1935 median=10.1000 p25=6.5667 p75=18.5667
- 1m n=41 mean=19.2337 median=10.4667 p25=6.7167 p75=16.7667
- 5m n=41 mean=18.0789 median=9.5100 p25=6.0233 p75=18.0867
### NEUTRAL
- 1s n=868 mean=13.0323 median=4.0000 p25=2.0000 p75=7.0000
- 5s n=868 mean=14.6696 median=5.4000 p25=2.6000 p75=15.4000
- 15s n=868 mean=13.0829 median=7.4000 p25=3.6000 p75=15.0000
- 30s n=868 mean=12.2756 median=8.2000 p25=4.5667 p75=13.7000
- 1m n=868 mean=12.3724 median=8.6500 p25=5.1333 p75=14.8000
- 5m n=868 mean=13.8834 median=9.5000 p25=5.7733 p75=16.5167
Matched tps effects map[1m:1.3457446808510638 1s:5.617021276595745 30s:3.815602836879434 5s:3.2510638297872356]

## Interpretation

The 1-minute lump artifact is gone. Short-window trades/sec is now a real number (not 0.033).

DISCOVERY: Exhaustion 1s mean 26.7 vs Continuation 7.7 vs Neutral 9.2. Matched 1s effect +19.3. YES vs both.

EXTERNAL HOLDOUT: Exhaustion 1s 17.5 vs Neutral 13.0 (YES) and matched 1s +5.6 (YES). Continuation 1s mean 32.5 is hotter than Exhaustion — short-window velocity does **not** isolate Exhaustion from Continuation on the holdout. That is the same P5.2 lesson: intensity alone is shared with Continuation.

Label: EXPLORATORY_MECHANISM_ONLY. V1 unchanged. PRE_2026_03_17 untouched.

