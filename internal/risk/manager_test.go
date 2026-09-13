package risk

import (
	"math"
	"testing"
	"time"

	"aurumflow/internal/market"
)

func spec(min, max, step, vpp float64) market.InstrumentSpec {
	return market.InstrumentSpec{
		Epic: "GOLD", MinDealSize: min, MaxDealSize: max, SizeStep: step, ValuePerPoint: vpp,
	}
}

func TestComputeSize_Normal(t *testing.T) {
	// balance 10000, 1% risk = 100; stop 10; vpp 1 → size 10
	sz, err := ComputeSize(10000, 1, 10, spec(0.1, 50, 0.1, 1))
	if err != nil || sz != 10 {
		t.Fatalf("sz=%v err=%v", sz, err)
	}
}

func TestComputeSize_ZeroAndNegativeBalance(t *testing.T) {
	if _, err := ComputeSize(0, 1, 10, spec(0.1, 50, 0.1, 1)); err == nil {
		t.Fatal("zero balance")
	}
	if _, err := ComputeSize(-1, 1, 10, spec(0.1, 50, 0.1, 1)); err == nil {
		t.Fatal("neg balance")
	}
}

func TestComputeSize_StopDistance(t *testing.T) {
	if _, err := ComputeSize(10000, 1, 0, spec(0.1, 50, 0.1, 1)); err == nil {
		t.Fatal("zero stop")
	}
	if _, err := ComputeSize(10000, 1, 1e-12, spec(0.1, 50, 0.1, 1)); err == nil {
		t.Fatal("tiny stop should still compute or fail safely")
	}
	sz, err := ComputeSize(10000, 1, 10000, spec(0.1, 50, 0.1, 1))
	if err == nil {
		t.Fatalf("huge stop should fail min size, sz=%v", sz)
	}
	if _, err := ComputeSize(10000, 1, math.NaN(), spec(0.1, 50, 0.1, 1)); err == nil {
		t.Fatal("nan stop")
	}
	if _, err := ComputeSize(10000, 1, math.Inf(1), spec(0.1, 50, 0.1, 1)); err == nil {
		t.Fatal("inf stop")
	}
}

func TestComputeSize_MinSizeRejects(t *testing.T) {
	// risk amount too small to reach min 1.0
	_, err := ComputeSize(100, 0.1, 50, spec(1, 10, 0.1, 1))
	if err == nil || err.Error()[:len(ReasonMinSizeExceedsRisk)] != ReasonMinSizeExceedsRisk {
		t.Fatalf("want min-size reject, got %v", err)
	}
}

func TestComputeSize_MaxSizeRejects(t *testing.T) {
	_, err := ComputeSize(1e9, 50, 1, spec(0.1, 1, 0.1, 1))
	if err == nil {
		t.Fatal("expected max reject")
	}
}

func TestComputeSize_StepRounding(t *testing.T) {
	sz, err := ComputeSize(10000, 1, 30, spec(0.1, 50, 0.1, 1))
	if err != nil {
		t.Fatal(err)
	}
	// 100/30 = 3.333 → 3.3
	if math.Abs(sz-3.3) > 1e-9 {
		t.Fatalf("sz=%v", sz)
	}
}

func TestComputeSize_IncompleteSpec(t *testing.T) {
	if _, err := ComputeSize(10000, 1, 10, market.InstrumentSpec{Epic: "X"}); err == nil {
		t.Fatal("incomplete")
	}
}

func TestComputeSize_NaNBalance(t *testing.T) {
	if _, err := ComputeSize(math.NaN(), 1, 10, spec(0.1, 50, 0.1, 1)); err == nil {
		t.Fatal("nan")
	}
}

func TestDailyDrawdownThresholds(t *testing.T) {
	m := NewManager(1, 3, 5, 10000)
	m.UpdateBalance(10000)
	if !m.CanOpenTrade() {
		t.Fatal("full balance")
	}
	m.UpdateBalance(9600) // 4% < 5
	if !m.CanOpenTrade() {
		t.Fatal("below")
	}
	m.UpdateBalance(9500) // 5% not < 5
	if m.CanOpenTrade() {
		t.Fatal("at limit must block")
	}
	m.UpdateBalance(9000)
	if m.CanOpenTrade() {
		t.Fatal("above")
	}
}

func TestMaxOpenTrades(t *testing.T) {
	m := NewManager(1, 2, 50, 10000)
	m.SetOpenCount(2)
	if m.CanOpenTrade() {
		t.Fatal("max trades")
	}
}

func TestSelectDemoAccount_UniqueAndPreferred(t *testing.T) {
	one := []market.AccountInfo{{AccountID: "A", AccountType: "CFD", Balance: market.Balance{Balance: 1}}}
	got, err := SelectDemoAccount(one, "")
	if err != nil || got.AccountID != "A" {
		t.Fatalf("single: %v %#v", err, got)
	}
	many := []market.AccountInfo{
		{AccountID: "A", AccountType: "SPREADBET", Preferred: false},
		{AccountID: "B", AccountType: "CFD", Preferred: true, Balance: market.Balance{Balance: 9}},
		{AccountID: "C", AccountType: "CFD", Preferred: false},
	}
	got, err = SelectDemoAccount(many, "A")
	if err != nil || got.AccountID != "B" {
		t.Fatalf("preferred cfd: %v %#v", err, got)
	}
	ambiguous := []market.AccountInfo{
		{AccountID: "A", AccountType: "CFD", Preferred: false},
		{AccountID: "B", AccountType: "CFD", Preferred: false},
	}
	got, err = SelectDemoAccount(ambiguous, "B")
	if err != nil || got.AccountID != "B" {
		t.Fatalf("session current: %v %#v", err, got)
	}
	if _, err := SelectDemoAccount(ambiguous, ""); err == nil {
		t.Fatal("ambiguous without current must refuse")
	}
}

func TestSelectAccount_NoFallback(t *testing.T) {
	accs := []market.AccountInfo{{AccountID: "A", Balance: market.Balance{Balance: 1}}, {AccountID: "B", Balance: market.Balance{Balance: 2}}}
	if _, err := SelectAccount(accs, "", ""); err == nil {
		t.Fatal("empty must refuse accounts[0]")
	}
	got, err := SelectAccount(accs, "B", "A")
	if err != nil || got.AccountID != "B" || got.Balance.Balance != 2 {
		t.Fatalf("%v %#v", err, got)
	}
	if _, err := SelectAccount(accs, "Z", ""); err == nil {
		t.Fatal("missing account")
	}
}

func TestResetDaily(t *testing.T) {
	m := NewManager(1, 3, 5, 10000)
	m.UpdateBalance(8000)
	m.ResetDailyIfNewDay(time.Now().UTC().Add(25 * time.Hour))
	if m.DailyDrawdownPct() != 0 {
		t.Fatalf("dd=%v", m.DailyDrawdownPct())
	}
}

func TestValidateSignalNil(t *testing.T) {
	m := NewManager(1, 1, 5, 10000)
	if ok, err := m.ValidateSignal(nil); ok || err == nil {
		t.Fatal("nil")
	}
}
