# FLOW EXHAUSTION MECHANISM

EXPLORATORY_MECHANISM_RESEARCH only. Not a new validation number.

Runtime 4m17.7886858s events=243352328 ev/s=943999 mem=107.6MB

## DISCOVERY (2026-06-15→2026-09-11)
Legacy=928 Exh=56 Cont=33 Neu=839 Central=YES
### EXHAUSTION
- flow_magnitude n=56 mean=4.55441 median=1.38621 p25=0.69370 p75=5.07640
- flow_imbalance n=56 mean=0.03508 median=0.07572 p25=-0.16391 p75=0.27456
- cvd n=56 mean=23.71627 median=-50.20200 p25=-202.22600 p75=130.22400
- cvd_slope n=56 mean=0.22284 median=0.08401 p25=-0.20637 p75=0.58406
- trade_velocity n=56 mean=0.03310 median=0.03333 p25=0.03333 p75=0.03333
- price_displacement n=56 mean=0.00081 median=0.00063 p25=0.00003 p75=0.00141
- flow_efficiency n=56 mean=-3.07923 median=-0.58126 p25=-3.02356 p75=0.62994
- impact_failure n=56 mean=3.07923 median=0.61882 p25=-0.49850 p75=3.03480
### CONTINUATION
- flow_magnitude n=33 mean=3.66945 median=2.56517 p25=1.10548 p75=4.62685
- flow_imbalance n=33 mean=-0.07369 median=-0.12827 p25=-0.34345 p75=0.19648
- cvd n=33 mean=-161.77482 median=-150.42000 p25=-400.29800 p75=71.82600
- cvd_slope n=33 mean=0.02331 median=-0.10351 p25=-0.49924 p75=0.49376
- trade_velocity n=33 mean=0.03333 median=0.03333 p25=0.03333 p75=0.03333
- price_displacement n=33 mean=0.00090 median=0.00054 p25=0.00031 p75=0.00123
- flow_efficiency n=33 mean=-1.75609 median=-1.16943 p25=-2.76333 p75=-0.06356
- impact_failure n=33 mean=1.75609 median=1.16943 p25=0.06356 p75=2.76333
### NEUTRAL
- flow_magnitude n=839 mean=1.54065 median=0.66056 p25=0.35316 p75=1.08399
- flow_imbalance n=839 mean=-0.00571 median=-0.00266 p25=-0.16978 p75=0.16096
- cvd n=839 mean=-48.66097 median=-33.27400 p25=-155.80600 p75=77.50700
- cvd_slope n=839 mean=0.00689 median=-0.00117 p25=-0.12459 p75=0.11802
- trade_velocity n=839 mean=0.03329 median=0.03333 p25=0.03333 p75=0.03333
- price_displacement n=839 mean=0.00031 median=0.00023 p25=-0.00006 p75=0.00061
- flow_efficiency n=839 mean=-0.86746 median=-0.10295 p25=-1.14286 p75=0.52677
- impact_failure n=839 mean=0.86746 median=0.10295 p25=-0.52677 p75=1.14286
Matched stratified nearest unused control (dir, UTC hour, vol tertile, score bucket); no future outcomes exh=52 ctl=52 effects=map[cvd:35.11669230769315 flow_efficiency:-2.834044640743964 flow_magnitude:3.2924769853466245 impact_failure:2.834044640743964 trade_velocity:-0.00012820512820512815]
PrePost {Note:PRE = mean |imbalance| t-15..t0; POST = mean price_norm * legacy at +15 (descriptive) PreAggMean:0.34730709067174675 PostDispLegacy:-0.000671996023482912}

## EXTERNAL_HOLDOUT (2026-03-17→2026-06-14)
Legacy=959 Exh=50 Cont=41 Neu=868 Central=YES
### EXHAUSTION
- flow_magnitude n=50 mean=4.44376 median=1.81040 p25=0.60449 p75=3.90711
- flow_imbalance n=50 mean=-0.01569 median=-0.02182 p25=-0.27438 p75=0.25715
- cvd n=50 mean=-73.23152 median=-28.05400 p25=-221.74800 p75=214.43000
- cvd_slope n=50 mean=-0.38965 median=-0.04220 p25=-0.52327 p75=0.34673
- trade_velocity n=50 mean=0.03293 median=0.03333 p25=0.03333 p75=0.03333
- price_displacement n=50 mean=0.00116 median=0.00063 p25=0.00038 p75=0.00125
- flow_efficiency n=50 mean=-2.85561 median=-0.91358 p25=-2.56894 p75=0.17450
- impact_failure n=50 mean=2.85561 median=1.02838 p25=-0.17450 p75=2.56894
### CONTINUATION
- flow_magnitude n=41 mean=5.35393 median=2.57024 p25=0.58745 p75=7.71486
- flow_imbalance n=41 mean=0.00395 median=-0.03031 p25=-0.29528 p75=0.29467
- cvd n=41 mean=-39.60590 median=-12.34800 p25=-271.51700 p75=143.38500
- cvd_slope n=41 mean=0.08905 median=-0.04883 p25=-0.31754 p75=0.55325
- trade_velocity n=41 mean=0.03333 median=0.03333 p25=0.03333 p75=0.03333
- price_displacement n=41 mean=0.00119 median=0.00094 p25=0.00030 p75=0.00156
- flow_efficiency n=41 mean=-3.69748 median=-1.57640 p25=-5.48810 p75=-0.08263
- impact_failure n=41 mean=3.69748 median=1.57640 p25=0.08263 p75=5.48810
### NEUTRAL
- flow_magnitude n=868 mean=1.41296 median=0.66207 p25=0.33877 p75=0.95287
- flow_imbalance n=868 mean=0.00650 median=0.01168 p25=-0.16172 p75=0.18084
- cvd n=868 mean=-17.60620 median=-20.03500 p25=-159.45200 p75=115.59300
- cvd_slope n=868 mean=0.01768 median=0.00756 p25=-0.13141 p75=0.14007
- trade_velocity n=868 mean=0.03328 median=0.03333 p25=0.03333 p75=0.03333
- price_displacement n=868 mean=0.00033 median=0.00026 p25=-0.00011 p75=0.00067
- flow_efficiency n=868 mean=-0.80271 median=-0.15902 p25=-1.10346 p75=0.51647
- impact_failure n=868 mean=0.80271 median=0.16345 p25=-0.51420 p75=1.10392
Matched stratified nearest unused control (dir, UTC hour, vol tertile, score bucket); no future outcomes exh=48 ctl=48 effects=map[cvd:-28.531687500001965 flow_efficiency:-1.6159484089500407 flow_magnitude:2.5173223959210436 impact_failure:1.6159484089500407 trade_velocity:0]
PrePost {Note:PRE = mean |imbalance| t-15..t0; POST = mean price_norm * legacy at +15 (descriptive) PreAggMean:0.3306946198779857 PostDispLegacy:0.0002238318182019429}

## Consistency
- flow_magnitude: CONSISTENT
- flow_efficiency: CONSISTENT
- impact_failure: CONSISTENT
- cvd: INCONSISTENT
- trade_velocity: CONSISTENT

OVERALL central question: YES

## Interpretation (not a new validation)

YES means Exhaustion prints carry more pre-signal FlowMagnitude and higher ImpactFailure than Neutral / matched Legacy prints on both datasets.

It does **not** mean raw 5m price displacement is smaller than Neutral. Raw DispVsFlow is slightly larger (0.00081 vs 0.00031 discovery; 0.00116 vs 0.00033 holdout). Displacement simply fails to scale with the extra aggression.

CVD level is INCONSISTENT (discovery mean +23.7, holdout −73.2). Trade velocity on the 5m engine window is an artifact of 1-minute lumped tape (~0.033 = 2 prints / 60s). Path-level radar velocity (hundreds–thousands/min) is the usable intensity series.

Continuation is also a high-flow region, especially on the holdout (flow_magnitude 5.35 vs Exhaustion 4.44). The clean contrast is Exhaustion vs Neutral, not Exhaustion vs Continuation.

LONG/SHORT membership (from frozen V1 counts, not re-pooled):
- DISCOVERY exhaustion: LONG 17 / SHORT 39
- HOLDOUT exhaustion: LONG 22 / SHORT 28 (outcome-symmetric on the confirmatory study)

PRE path |imbalance| ≈ 0.33–0.35. Post +15m unsigned price_norm is mixed in the path helper (not direction-adjusted). Directional +15m behavior remains the original V1 labels, which were not recomputed here.

These numbers are EXPLORATORY_MECHANISM_RESEARCH. They do not replace or update VALIDATED_EXTERNAL_HOLDOUT.
