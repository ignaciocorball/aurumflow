package research

import "testing"

func TestVelAnswerThresholds(t *testing.T) {
	exh := map[string]VelDist{"1s": {N: 50, Mean: 8}, "5s": {N: 50, Mean: 6}}
	cont := map[string]VelDist{"1s": {N: 30, Mean: 4}}
	neu := map[string]VelDist{"1s": {N: 100, Mean: 2}, "5s": {N: 100, Mean: 2}}
	if velAnswer(exh, cont, neu, map[string]float64{"1s": 1}) != "YES" {
		t.Fatal("expected YES")
	}
	small := map[string]VelDist{"1s": {N: 5, Mean: 9}}
	if velAnswer(small, cont, neu, nil) != "INCONCLUSIVE" {
		t.Fatal("n")
	}
}
