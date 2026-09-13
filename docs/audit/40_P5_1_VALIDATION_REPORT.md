# 40 — P5.1 Validation Report

Spec `FLOW_EXHAUSTION_V1` hash `f66e744e…f7399b5d` was frozen before holdout outcomes were computed.

Pressure formula unchanged. Threshold unchanged (`|P| >= 15`).

## EXTERNAL_HOLDOUT_1 (confirmatory)

```text
events     135,907,413
runtime    57.0s
events/sec 2,382,891
mem MB     35.0

Legacy baseline          n=959  15m mean=+0.00005  hit=0.504  mfe/mae=0.99
FLOW_CONTINUATION        n=41   15m mean=-0.00068  hit=0.317  mfe/mae=0.39
FLOW_EXHAUSTION_CONFIRM  n=50   15m mean=+0.00098  hit=0.660  mfe/mae=1.99
                         ci=[+0.00030,+0.00181]
Delta vs Legacy          mean=+0.00093  hit=+0.156  mfe/mae=+1.00
MAE                      -0.00090 vs -0.00099  (not catastrophic)

LONG exhaustion          n=22  mean=+0.00137
SHORT exhaustion         n=28  mean=+0.00068
Symmetry                 SYMMETRIC

EARLY                    n=20  mean=+0.00105  sign=+
MIDDLE                   n=15  mean=+0.00118  sign=+
LATE                     n=15  mean=+0.00068  sign=+
```

Decision on holdout: **SUPPORTED** (n=50 meets the floor; primary mean > baseline; 3/3 subperiod signs positive; MAE not worse).

n=50 is the minimum. Treat as a thin confirmatory sample, not a license to trade.

## POST_HOC_DISCOVERY (not validation)

```text
2026-06-15 → 2026-09-11
events 107,444,915
exhaustion n=56  15m mean=+0.00106  hit=0.750
legacy     n=928 15m mean=+0.00010  hit=0.541
```

Same sign as holdout. Not combined with holdout.

## Assets

```text
BTC_FLOW_EXHAUSTION_V1 = VALIDATED_EXTERNAL_HOLDOUT
GOLD_FLOW_EXHAUSTION   = UNVALIDATED
US100_FLOW_EXHAUSTION  = UNVALIDATED
```

Not promoted. RADAR_MODE remains SHADOW.
