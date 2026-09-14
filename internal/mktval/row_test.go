package mktval

import "testing"

func TestAttentionCannotPromote(t *testing.T) {
	r := Row{Market: "US100", M15: 400, Signals: 3, AttentionNote: "90"}
	st := DecideStatus(r, Stats{N: 3, Expectancy: 2, PF: 9, MaxDD: 1, PositiveThirds: 3}, true, true)
	if st == ShadowValidated {
		t.Fatal("tiny sample must not promote even with high attention note")
	}
}
