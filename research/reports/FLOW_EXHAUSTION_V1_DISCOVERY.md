# FLOW_EXHAUSTION_V1

Role: POST_HOC_DISCOVERY
Spec: FLOW_EXHAUSTION_V1
Hash: f66e744ec7f8e5785bc6d43b2fc8211941ac725bd463baa33782d955f7399b5d
Range: 2026-06-15T16:00:00Z → 2026-09-11T23:00:00Z (89 days)
Events: 107444915
Runtime: 45.5575393s
Events/sec: 2358444
Mem MB: 30.7
Decision: POST_HOC_DIAGNOSTIC/SUPPORTED

## Legacy baseline
legacy n=928 hit=0.541 mean=0.00010 ci=[0.00002,0.00019] median=0.00008 mfe=0.00092 mae=-0.00092 mfe/mae=1.00

## FLOW_NEUTRAL
neutral n=839 hit=0.539 mean=0.00007 ci=[-0.00002,0.00015] median=0.00007 mfe=0.00086 mae=-0.00091 mfe/mae=0.95

## FLOW_CONTINUATION_CONFIRM
continuation n=33 hit=0.242 mean=-0.00073 ci=[-0.00112,-0.00018] median=-0.00064 mfe=0.00064 mae=-0.00176 mfe/mae=0.37

## FLOW_EXHAUSTION_CONFIRM
exhaustion n=56 hit=0.750 mean=0.00106 ci=[0.00054,0.00161] median=0.00063 mfe=0.00198 mae=-0.00058 mfe/mae=3.41
Delta mean=0.00096 hit=0.209 mfe/mae=2.41
TimeToMFE=590s TimeToMAE=289s
exhaustion_n=56 late_aggressive_sell_into_long=17 late_aggressive_buy_into_short=39 (no participant identity)

## LONG/SHORT
exh_long n=17 hit=0.765 mean=0.00064 ci=[0.00029,0.00117] median=0.00046 mfe=0.00158 mae=-0.00057 mfe/mae=2.78
exh_short n=39 hit=0.744 mean=0.00124 ci=[0.00060,0.00204] median=0.00077 mfe=0.00216 mae=-0.00059 mfe/mae=3.68
leg_long n=491 hit=0.525 mean=0.00005 ci=[-0.00006,0.00016] median=0.00005 mfe=0.00092 mae=-0.00094 mfe/mae=0.98
leg_short n=437 hit=0.558 mean=0.00015 ci=[0.00003,0.00027] median=0.00013 mfe=0.00092 mae=-0.00090 mfe/mae=1.03
Symmetry: SHORT_ONLY

## Subperiods
- EARLY 2026-06-15 sign=1
  exh n=15 hit=0.733 mean=0.00082 ci=[0.00027,0.00142] median=0.00073 mfe=0.00161 mae=-0.00060 mfe/mae=2.66
  legacy n=330 hit=0.539 mean=0.00004 ci=[-0.00008,0.00015] median=0.00010 mfe=0.00094 mae=-0.00101 mfe/mae=0.92
- MIDDLE 2026-07-15 sign=1
  exh n=18 hit=0.833 mean=0.00066 ci=[0.00034,0.00105] median=0.00051 mfe=0.00135 mae=-0.00056 mfe/mae=2.42
  legacy n=269 hit=0.569 mean=0.00018 ci=[0.00007,0.00028] median=0.00014 mfe=0.00079 mae=-0.00065 mfe/mae=1.21
- LATE 2026-08-13 sign=1
  exh n=23 hit=0.696 mean=0.00153 ci=[0.00035,0.00290] median=0.00074 mfe=0.00272 mae=-0.00058 mfe/mae=4.66
  legacy n=329 hit=0.520 mean=0.00009 ci=[-0.00009,0.00027] median=0.00004 mfe=0.00101 mae=-0.00104 mfe/mae=0.97

## Magnitude (exhaustion only)
- 15-20 n=24 hit=0.750 mean=0.00047 ci=[0.00014,0.00077] median=0.00053 mfe=0.00142 mae=-0.00071 mfe/mae=1.99
- 20-40 n=22 hit=0.773 mean=0.00114 ci=[0.00049,0.00215] median=0.00059 mfe=0.00189 mae=-0.00055 mfe/mae=3.43
- 40-60 n=10 hit=0.700 mean=0.00229 ci=[0.00091,0.00398] median=0.00165 mfe=0.00353 mae=-0.00033 mfe/mae=10.76
- 60+ n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00

## Delay diagnostic (POST_HOC_DELAY_DIAGNOSTIC)
- t0 exh n=56 hit=0.750 mean=0.00106 ci=[0.00054,0.00161] median=0.00063 mfe=0.00198 mae=-0.00058 mfe/mae=3.41
- t0-1m exh n=33 hit=0.697 mean=0.00006 ci=[-0.00043,0.00052] median=0.00022 mfe=0.00114 mae=-0.00105 mfe/mae=1.09
- t0-3m exh n=35 hit=0.629 mean=0.00080 ci=[0.00018,0.00148] median=0.00023 mfe=0.00172 mae=-0.00088 mfe/mae=1.95
- t0-5m exh n=33 hit=0.576 mean=0.00042 ci=[-0.00016,0.00138] median=0.00015 mfe=0.00131 mae=-0.00081 mfe/mae=1.62

## Secondary horizons
5m exh5 n=56 hit=0.643 mean=0.00072 ci=[0.00033,0.00115] median=0.00019 mfe=0.00115 mae=-0.00034 mfe/mae=3.39
15m exh15 n=56 hit=0.750 mean=0.00106 ci=[0.00054,0.00161] median=0.00063 mfe=0.00198 mae=-0.00058 mfe/mae=3.41
30m exh30 n=56 hit=0.661 mean=0.00124 ci=[0.00059,0.00193] median=0.00052 mfe=0.00275 mae=-0.00081 mfe/mae=3.39
1h exh1h n=56 hit=0.625 mean=0.00111 ci=[-0.00004,0.00249] median=0.00055 mfe=0.00419 mae=-0.00178 mfe/mae=2.36
