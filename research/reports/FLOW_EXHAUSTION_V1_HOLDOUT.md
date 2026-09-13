# FLOW_EXHAUSTION_V1

Role: EXTERNAL_HOLDOUT_1
Spec: FLOW_EXHAUSTION_V1
Hash: f66e744ec7f8e5785bc6d43b2fc8211941ac725bd463baa33782d955f7399b5d
Range: 2026-03-17T12:30:00Z → 2026-06-14T19:30:00Z (90 days)
Events: 135907413
Runtime: 57.034683s
Events/sec: 2382891
Mem MB: 35.0
Decision: SUPPORTED

## Legacy baseline
legacy n=959 hit=0.504 mean=0.00005 ci=[-0.00004,0.00015] median=0.00001 mfe=0.00098 mae=-0.00099 mfe/mae=0.99

## FLOW_NEUTRAL
neutral n=868 hit=0.503 mean=0.00003 ci=[-0.00006,0.00012] median=0.00001 mfe=0.00095 mae=-0.00097 mfe/mae=0.99

## FLOW_CONTINUATION_CONFIRM
continuation n=41 hit=0.317 mean=-0.00068 ci=[-0.00108,-0.00028] median=-0.00054 mfe=0.00064 mae=-0.00167 mfe/mae=0.39

## FLOW_EXHAUSTION_CONFIRM
exhaustion n=50 hit=0.660 mean=0.00098 ci=[0.00030,0.00181] median=0.00037 mfe=0.00179 mae=-0.00090 mfe/mae=1.99
Delta mean=0.00093 hit=0.156 mfe/mae=1.00
TimeToMFE=511s TimeToMAE=284s
exhaustion_n=50 late_aggressive_sell_into_long=22 late_aggressive_buy_into_short=28 (no participant identity)

## LONG/SHORT
exh_long n=22 hit=0.727 mean=0.00137 ci=[0.00029,0.00303] median=0.00051 mfe=0.00207 mae=-0.00092 mfe/mae=2.26
exh_short n=28 hit=0.607 mean=0.00068 ci=[0.00003,0.00148] median=0.00031 mfe=0.00156 mae=-0.00088 mfe/mae=1.77
leg_long n=465 hit=0.475 mean=0.00001 ci=[-0.00011,0.00013] median=-0.00002 mfe=0.00089 mae=-0.00099 mfe/mae=0.90
leg_short n=494 hit=0.530 mean=0.00009 ci=[-0.00002,0.00021] median=0.00005 mfe=0.00107 mae=-0.00100 mfe/mae=1.07
Symmetry: SYMMETRIC

## Subperiods
- EARLY 2026-03-17 sign=1
  exh n=20 hit=0.450 mean=0.00105 ci=[-0.00033,0.00246] median=-0.00009 mfe=0.00203 mae=-0.00112 mfe/mae=1.81
  legacy n=324 hit=0.534 mean=0.00012 ci=[-0.00007,0.00027] median=0.00007 mfe=0.00108 mae=-0.00106 mfe/mae=1.02
- MIDDLE 2026-04-16 sign=1
  exh n=15 hit=0.933 mean=0.00118 ci=[0.00056,0.00201] median=0.00082 mfe=0.00157 mae=-0.00055 mfe/mae=2.85
  legacy n=311 hit=0.469 mean=-0.00002 ci=[-0.00017,0.00010] median=-0.00005 mfe=0.00082 mae=-0.00089 mfe/mae=0.92
- LATE 2026-05-16 sign=1
  exh n=15 hit=0.667 mean=0.00068 ci=[0.00001,0.00153] median=0.00036 mfe=0.00169 mae=-0.00095 mfe/mae=1.79
  legacy n=324 hit=0.506 mean=0.00004 ci=[-0.00009,0.00016] median=0.00002 mfe=0.00104 mae=-0.00103 mfe/mae=1.01

## Magnitude (exhaustion only)
- 15-20 n=16 hit=0.750 mean=0.00044 ci=[0.00044,0.00044] median=0.00026 mfe=0.00102 mae=-0.00078 mfe/mae=1.30
- 20-40 n=27 hit=0.593 mean=0.00076 ci=[-0.00011,0.00155] median=0.00038 mfe=0.00157 mae=-0.00074 mfe/mae=2.13
- 40-60 n=7 hit=0.714 mean=0.00308 ci=[-0.00005,0.00711] median=0.00076 mfe=0.00439 mae=-0.00178 mfe/mae=2.46
- 60+ n=0 hit=0.000 mean=0.00000 ci=[0.00000,0.00000] median=0.00000 mfe=0.00000 mae=0.00000 mfe/mae=0.00

## Delay diagnostic (POST_HOC_DELAY_DIAGNOSTIC)
- t0 exh n=50 hit=0.660 mean=0.00098 ci=[0.00030,0.00181] median=0.00037 mfe=0.00179 mae=-0.00090 mfe/mae=1.99
- t0-1m exh n=32 hit=0.344 mean=-0.00036 ci=[-0.00036,-0.00036] median=-0.00041 mfe=0.00118 mae=-0.00165 mfe/mae=0.72
- t0-3m exh n=46 hit=0.652 mean=0.00062 ci=[0.00002,0.00164] median=0.00013 mfe=0.00146 mae=-0.00119 mfe/mae=1.23
- t0-5m exh n=35 hit=0.543 mean=0.00051 ci=[-0.00028,0.00183] median=0.00013 mfe=0.00166 mae=-0.00143 mfe/mae=1.16

## Secondary horizons
5m exh5 n=50 hit=0.640 mean=0.00031 ci=[-0.00004,0.00067] median=0.00011 mfe=0.00089 mae=-0.00061 mfe/mae=1.47
15m exh15 n=50 hit=0.660 mean=0.00098 ci=[0.00030,0.00181] median=0.00037 mfe=0.00179 mae=-0.00090 mfe/mae=1.99
30m exh30 n=50 hit=0.600 mean=0.00143 ci=[0.00034,0.00254] median=0.00052 mfe=0.00261 mae=-0.00122 mfe/mae=2.14
1h exh1h n=50 hit=0.600 mean=0.00112 ci=[-0.00005,0.00223] median=0.00062 mfe=0.00334 mae=-0.00175 mfe/mae=1.91
