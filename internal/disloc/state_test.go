package disloc

import "testing"

func TestStates(t *testing.T) {
	if Classify(Input{}) != Normal {
		t.Fatal("normal")
	}
	if Classify(Input{HighSalience: 3}) != Elevated {
		t.Fatal("elev")
	}
	if Classify(Input{HighSalience: 5, MedianSal: 65}) != Severe {
		t.Fatal("sev")
	}
}
