package core

import "testing"

func TestScanReadyDistinctFromStrategyReady(t *testing.T) {
	l := &Loop{}
	if l.ScanReady() {
		t.Fatal("zero LastEvalAt is not scan-ready")
	}
	l.lastDecision = "data_frozen:M15"
	if l.ScanReady() {
		t.Fatal("frozen scan must not be ready")
	}
	l.lastDecision = "idle"
	if l.ScanReady() {
		t.Fatal("still no LastEvalAt")
	}
	l.LastEvalAt = l.LastEvalAt.Add(1)
	if !l.ScanReady() {
		t.Fatal("composer-eval + not frozen should be scan-ready")
	}
}
