# FREE RESEARCH BASELINE

Generated: 2026-09-13T04:11:32Z
Commit: 9564dde

Asset: Binance USD-M BTCUSDT (NOT Capital BTCUSD)
Range: 2026-06-15T00:00:00Z → 2026-09-11T23:58:00Z (89 days)
Trade events: 107444915
Radar snapshots: 128071
Runtime: 1m38.4894856s
Events/sec: 1090928
Mem MB: 34.1
Horizon: 15m0s

## A Legacy
Signals: 928 long=491 short=437
all n=928 hit=0.541 mean=0.00010 ci=[0.00002,0.00019] median=0.00008 mfe=0.00092 mae=-0.00092 mfe/mae=1.00
long n=491 hit=0.525 mean=0.00005 ci=[-0.00006,0.00016] median=0.00005 mfe=0.00092 mae=-0.00094 mfe/mae=0.98
short n=437 hit=0.558 mean=0.00015 ci=[0.00003,0.00027] median=0.00013 mfe=0.00092 mae=-0.00090 mfe/mae=1.03

## B Radar
Expansion snaps: 5674
radar_dir n=12410 hit=0.471 mean=-0.00005 ci=[-0.00009,-0.00001] median=-0.00008 mfe=0.00129 mae=-0.00131 mfe/mae=0.99

## C Fusion
Aligned: 33
Contradicted: 56
legacy_aligned n=33 hit=0.242 mean=-0.00073 ci=[-0.00112,-0.00018] median=-0.00064 mfe=0.00064 mae=-0.00176 mfe/mae=0.37
legacy_contradicted n=56 hit=0.750 mean=0.00106 ci=[0.00054,0.00161] median=0.00063 mfe=0.00198 mae=-0.00058 mfe/mae=3.41

## Horizons (1m candles; RESOLUTION_LIMITED: labels use 1m synthesized candles; +10s/+30s are not independent of +1m)
### 1m0s
legacy n=928 hit=0.444 mean=-0.00002 ci=[-0.00006,0.00002] median=-0.00001 mfe=0.00017 mae=-0.00018 mfe/mae=0.90
aligned n=33 hit=0.333 mean=-0.00003 ci=[-0.00026,0.00018] median=-0.00014 mfe=0.00021 mae=-0.00024 mfe/mae=0.86
contradicted n=56 hit=0.589 mean=0.00032 ci=[0.00005,0.00063] median=0.00006 mfe=0.00043 mae=-0.00011 mfe/mae=3.88
radar n=12410 hit=0.467 mean=-0.00001 ci=[-0.00002,0.00001] median=-0.00000 mfe=0.00023 mae=-0.00024 mfe/mae=0.96
### 5m0s
legacy n=928 hit=0.440 mean=-0.00007 ci=[-0.00014,-0.00000] median=-0.00008 mfe=0.00049 mae=-0.00057 mfe/mae=0.86
aligned n=33 hit=0.212 mean=-0.00065 ci=[-0.00095,-0.00029] median=-0.00070 mfe=0.00042 mae=-0.00097 mfe/mae=0.44
contradicted n=56 hit=0.643 mean=0.00072 ci=[0.00033,0.00115] median=0.00019 mfe=0.00115 mae=-0.00034 mfe/mae=3.39
radar n=12410 hit=0.474 mean=-0.00003 ci=[-0.00006,-0.00000] median=-0.00004 mfe=0.00069 mae=-0.00070 mfe/mae=0.98
### 15m0s
legacy n=928 hit=0.541 mean=0.00010 ci=[0.00002,0.00019] median=0.00008 mfe=0.00092 mae=-0.00092 mfe/mae=1.00
aligned n=33 hit=0.242 mean=-0.00073 ci=[-0.00112,-0.00018] median=-0.00064 mfe=0.00064 mae=-0.00176 mfe/mae=0.37
contradicted n=56 hit=0.750 mean=0.00106 ci=[0.00054,0.00161] median=0.00063 mfe=0.00198 mae=-0.00058 mfe/mae=3.41
radar n=12410 hit=0.471 mean=-0.00005 ci=[-0.00009,-0.00001] median=-0.00008 mfe=0.00129 mae=-0.00131 mfe/mae=0.99
### 30m0s
legacy n=928 hit=0.522 mean=0.00012 ci=[-0.00004,0.00027] median=0.00007 mfe=0.00155 mae=-0.00136 mfe/mae=1.14
aligned n=33 hit=0.333 mean=-0.00027 ci=[-0.00084,0.00041] median=-0.00052 mfe=0.00125 mae=-0.00201 mfe/mae=0.62
contradicted n=56 hit=0.661 mean=0.00124 ci=[0.00059,0.00193] median=0.00052 mfe=0.00275 mae=-0.00081 mfe/mae=3.39
radar n=12410 hit=0.477 mean=-0.00006 ci=[-0.00012,-0.00001] median=-0.00009 mfe=0.00188 mae=-0.00190 mfe/mae=0.99
### 1h0m0s
legacy n=928 hit=0.516 mean=0.00015 ci=[-0.00015,0.00042] median=0.00008 mfe=0.00255 mae=-0.00220 mfe/mae=1.16
aligned n=33 hit=0.545 mean=0.00004 ci=[-0.00105,0.00100] median=0.00011 mfe=0.00228 mae=-0.00252 mfe/mae=0.91
contradicted n=56 hit=0.625 mean=0.00111 ci=[-0.00004,0.00249] median=0.00055 mfe=0.00419 mae=-0.00178 mfe/mae=2.36
radar n=12410 hit=0.483 mean=-0.00007 ci=[-0.00013,0.00001] median=-0.00011 mfe=0.00275 mae=-0.00275 mfe/mae=1.00

## Radar filter
All: all n=928 hit=0.541 mean=0.00010 ci=[0.00002,0.00019] median=0.00008 mfe=0.00092 mae=-0.00092 mfe/mae=1.00
Excl contradictions: excl n=872 hit=0.528 mean=0.00004 ci=[-0.00004,0.00011] median=0.00005 mfe=0.00085 mae=-0.00094 mfe/mae=0.91
Improves? INCONCLUSIVE

## Pressure
- 0-20 n=873 hit=0.541 mean=0.00007 ci=[-0.00003,0.00016] median=0.00008 mfe=0.00087 mae=-0.00092 mfe/mae=0.95
- 20-40 n=39 hit=0.564 mean=0.00041 ci=[-0.00018,0.00103] median=0.00022 mfe=0.00144 mae=-0.00102 mfe/mae=1.40
- 40-60 n=16 hit=0.500 mean=0.00107 ci=[0.00107,0.00107] median=-0.00004 mfe=0.00223 mae=-0.00080 mfe/mae=2.80
- 60-80 n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00
- 80-101 n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00

## Confidence
- 0-40 n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00
- 40-60 n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00
- 60-80 n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00
- 80-101 n=928 hit=0.541 mean=0.00010 ci=[0.00002,0.00019] median=0.00008 mfe=0.00092 mae=-0.00092 mfe/mae=1.00

## States
- NO_TRADE n=888 hit=0.541 mean=0.00006 ci=[-0.00001,0.00013] median=0.00008 mfe=0.00087 mae=-0.00092 mfe/mae=0.95
- ABSORPTION n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00
- EXPANSION n=40 hit=0.550 mean=0.00091 ci=[0.00029,0.00162] median=0.00021 mfe=0.00202 mae=-0.00090 mfe/mae=2.23

## Walk-forward
Train 2026-06-15 → 2026-08-13
Valid 2026-08-13 → 2026-08-28
OOS 2026-08-28 → 2026-09-11
Legacy OOS legacy_oos n=176 hit=0.489 mean=0.00008 ci=[-0.00018,0.00031] median=-0.00001 mfe=0.00093 mae=-0.00097 mfe/mae=0.95
Radar OOS radar_oos n=2352 hit=0.481 mean=-0.00007 ci=[-0.00015,0.00001] median=-0.00007 mfe=0.00122 mae=-0.00126 mfe/mae=0.96
Fusion aligned OOS fusion_oos n=10 hit=0.400 mean=-0.00027 ci=[-0.00088,0.00049] median=-0.00040 mfe=0.00099 mae=-0.00136 mfe/mae=0.72

## Hypotheses
H1 Radar adds information: NOT SUPPORTED
H2 Pressure magnitude orders outcomes: NOT SUPPORTED
H3 Fusion improves Legacy OOS: NOT SUPPORTED

Costs: ESTIMATED_COST not applied (event-study forward returns, not executable PnL).
