package salience

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
)

func TestDeterministicNoLeakage(t *testing.T) {
	bars := make([]Bar, 30)
	for i := range bars {
		bars[i] = Bar{Close: 100 + float64(i)*0.01, High: 100.2 + float64(i)*0.01, Low: 99.8 + float64(i)*0.01}
	}
	a := Score(Input{Market: "J225", Bars: bars, CrossAbsRet: []float64{0.001, 0.002, 0.01}})
	b := Score(Input{Market: "J225", Bars: bars, CrossAbsRet: []float64{0.001, 0.002, 0.01}})
	if a.Score != b.Score || a.Label != "MARKET SALIENCE" {
		t.Fatalf("%+v %+v", a, b)
	}
	quiet := make([]Bar, 30)
	for i := range quiet {
		quiet[i] = Bar{Close: 100, High: 100.01, Low: 99.99}
	}
	q := Score(Input{Market: "GOLD", Bars: quiet, CrossAbsRet: []float64{0.001}})
	if a.Score <= q.Score {
		t.Fatalf("moving market must outrank flat: j225=%.1f gold=%.1f", a.Score, q.Score)
	}
}

func TestInsufficientAndPenalty(t *testing.T) {
	if Score(Input{Market: "X", Bars: []Bar{{Close: 1}}}).Reason != "INSUFFICIENT_HISTORY" {
		t.Fatal("short")
	}
	bars := make([]Bar, 20)
	for i := range bars {
		bars[i] = Bar{Close: 10 + float64(i), High: 11 + float64(i), Low: 9 + float64(i)}
	}
	ok := Score(Input{Market: "X", Bars: bars})
	pen := Score(Input{Market: "X", Bars: bars, DQPenalty: 1})
	if pen.Score != 0 || ok.Score <= 0 {
		t.Fatalf("penalty ok=%.1f pen=%.1f", ok.Score, pen.Score)
	}
	if DQPenalty("HEALTHY") != 0 || DQPenalty("STALE") != 0.35 || DQPenalty("CLOSED") != 1 {
		t.Fatal("dq")
	}
}

func TestDoesNotAffectExecutionConstants(t *testing.T) {
	if Spec != "MARKET_SALIENCE_V0" {
		t.Fatal(Spec)
	}
	raw, err := os.ReadFile("../../research/specs/MARKET_SALIENCE_V0.md")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	got := "sha256:" + hex.EncodeToString(sum[:])
	if got != SpecHash {
		t.Fatalf("spec mutated without version bump: %s want %s", got, SpecHash)
	}
}
