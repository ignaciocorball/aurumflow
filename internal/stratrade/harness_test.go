package stratrade

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTradeIdentityDeterministic(t *testing.T) {
	a := TradeID("SIG-1", "REF-A", "DEAL-9", "GOLD")
	b := TradeID("SIG-1", "REF-A", "DEAL-9", "GOLD")
	c := TradeID("SIG-1", "REF-B", "DEAL-9", "GOLD")
	if a == "" || a != b {
		t.Fatal(a, b)
	}
	if a == c {
		t.Fatal("identity must include dealReference")
	}
	if !IsStrategyOrigin(OriginLegacy) || IsStrategyOrigin(OriginCalibration) {
		t.Fatal("origin filter")
	}
}

func TestPreSignalImmutable(t *testing.T) {
	dir := t.TempDir()
	r := NewRecorder(dir, "GOLD-test")
	s := FreezePreSignal(PreSignalSnapshot{
		SignalID: "S1", StrategyTradeID: "GOLD-test", Timestamp: time.Now().UTC(),
		Instrument: "GOLD", Direction: "BUY", EntryCandidate: 4300, StopLoss: 4290, TakeProfit: 4320,
	})
	if !s.Immutable || s.Origin != OriginLegacy {
		t.Fatal(s)
	}
	if err := r.PersistPreSignal(s); err != nil {
		t.Fatal(err)
	}
	if err := r.PersistPreSignal(s); err == nil {
		t.Fatal("immutable overwrite allowed")
	}
}

func TestOrderedLifecycleAndOneSignalOneOrder(t *testing.T) {
	dir := t.TempDir()
	r := NewRecorder(dir, "GOLD-life")
	_ = r.PersistPreSignal(FreezePreSignal(PreSignalSnapshot{SignalID: "S", StrategyTradeID: "GOLD-life", Timestamp: time.Now().UTC()}))
	for _, ev := range RequiredLifecycle {
		if ev == EvRiskRejected {
			continue
		}
		if err := r.Append(Event{Event: ev}); err != nil {
			t.Fatal(err)
		}
	}
	if !LifecycleComplete(r.Events()) || !LifecycleOrdered(r.Events()) {
		t.Fatal(r.Events())
	}
	if !OneSignalOneOrder(1) || OneSignalOneOrder(2) {
		t.Fatal("one signal one order")
	}
}

func TestProtectiveAndMonetaryReconciliation(t *testing.T) {
	if !ProtectiveOK(4290, 4320, 4290, 4320, 0.05) {
		t.Fatal("exact match")
	}
	if ProtectiveOK(4290, 4320, 4280, 4320, 0.05) {
		t.Fatal("SL mismatch must fail")
	}
	risk := ExpectedRiskUSD(0.01, 10, 1.00)
	if risk != 0.10 {
		t.Fatalf("expected risk %v", risk)
	}
	upl := ExpectedUPL("BUY", 0.01, 4300, 4310, 1.00)
	if upl != 0.10 {
		t.Fatalf("upl %v", upl)
	}
	sell := ExpectedUPL("SELL", 0.01, 4300, 4290, 1.00)
	if sell != 0.10 {
		t.Fatalf("sell upl %v", sell)
	}
	if abs := EntrySlippage("BUY", 4300.20, 4300.50); abs < 0.299 || abs > 0.301 {
		t.Fatalf("buy slip %v", abs)
	}
	if abs := EntrySlippage("SELL", 4300.20, 4299.90); abs < 0.299 || abs > 0.301 {
		t.Fatalf("sell slip %v", abs)
	}
}

func TestTradeVsOperationalOutcome(t *testing.T) {
	if TradeOutcome(-1.2, 0.01) != TradeLoss {
		t.Fatal("loss")
	}
	if OperationalOutcome(false, false, false, false) != OpsClean {
		t.Fatal("clean losing trade must still be operationally clean")
	}
	if OperationalOutcome(true, false, false, false) != OpsFailed {
		t.Fatal("mismatch is failed even if pnl positive")
	}
}

func TestMaxTradesPersistsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	path := WeekPath(dir, "GOLD")
	w, err := LoadWeek(path, "GOLD", 1)
	if err != nil {
		t.Fatal(err)
	}
	if w.BlocksNewStrategy() {
		t.Fatal("fresh week must allow first trade")
	}
	if err := w.MarkOpen("T1", "S1"); err != nil {
		t.Fatal(err)
	}
	if err := w.MarkClosed("T1"); err != nil {
		t.Fatal(err)
	}
	if !w.BlocksNewStrategy() {
		t.Fatal("completed maxTrades=1 must block")
	}
	reloaded, err := LoadWeek(path, "GOLD", 1)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.CompletedStrategyTrades != 1 || !reloaded.BlocksNewStrategy() {
		t.Fatalf("%+v", reloaded.Snapshot())
	}
	if !MaxTradesSafeAfterRestart(1, reloaded.CompletedStrategyTrades, 1, false) {
		t.Fatal("restart must not reset completed count")
	}
	if MaxTradesSafeAfterRestart(1, 0, 1, true) {
		t.Fatal("reset that allows a second trade is unsafe")
	}
}

func TestRecoveryWhileFlat(t *testing.T) {
	if !RecoveryFlatOK(0, 0, 0) {
		t.Fatal("flat")
	}
	if RecoveryFlatOK(1, 0, 0) || RecoveryFlatOK(0, 1, 0) || RecoveryFlatOK(0, 0, 1) {
		t.Fatal("not flat")
	}
}

func TestOffHoursObservabilityAndNextSession(t *testing.T) {
	// Sunday 22:30 UTC is after NY 22:00; next eligible LONDON 08:00 Monday.
	now := time.Date(2026, 9, 13, 22, 30, 0, 0, time.UTC)
	r := ObserveSession(now, []string{"LONDON", "NY"}, "TRADEABLE")
	if r.BrokerMarket != "TRADEABLE" {
		t.Fatal(r.BrokerMarket)
	}
	if r.StrategySession != "OFF_HOURS" {
		t.Fatal(r.StrategySession)
	}
	if r.StrategyReady {
		t.Fatal("must not be ready off hours")
	}
	if r.NextSession != "LONDON" || r.NextSessionAt.Hour() != 8 {
		t.Fatalf("next=%s at=%s", r.NextSession, r.NextSessionAt)
	}
	if r.StrategyWaiting != "STRATEGY WAITING FOR SESSION" {
		t.Fatal(r.StrategyWaiting)
	}
	london := time.Date(2026, 9, 14, 10, 0, 0, 0, time.UTC)
	on := ObserveSession(london, []string{"LONDON", "NY"}, "TRADEABLE")
	if !on.StrategyReady || on.StrategySession != "LONDON" {
		t.Fatalf("%+v", on)
	}
	implicitAll := ObserveSession(now, nil, "TRADEABLE")
	if !implicitAll.StrategyReady || implicitAll.NextSession == "" {
		t.Fatalf("empty sessions follow existing CanTrade ALL semantics: %+v", implicitAll)
	}
}

func TestAuditReconstruct(t *testing.T) {
	dir := t.TempDir()
	id := "GOLD-audit"
	r := NewRecorder(dir, id)
	_ = r.PersistPreSignal(FreezePreSignal(PreSignalSnapshot{SignalID: "S", StrategyTradeID: id, Direction: "BUY", Timestamp: time.Now().UTC()}))
	_ = r.Append(Event{Event: EvSignalObserved})
	_ = r.Finish(FinalRecord{ExitReason: "UNKNOWN", TradeOutcome: TradeBreakeven, OperationalOutcome: OpsWarning})
	a, err := Reconstruct(dir, id)
	if err != nil {
		t.Fatal(err)
	}
	if a.PreSignal == nil || a.Final == nil || len(a.Events) != 1 {
		t.Fatalf("%+v", a)
	}
	txt := FormatAudit(a)
	if txt == "" {
		t.Fatal("empty audit")
	}
}

func TestPromotionRequiresRecovery(t *testing.T) {
	sc := EmptyScorecard()
	sc = EvaluateTrust(sc, true, false, false, true, true, true, false)
	if MayPromoteDemoTrusted(true, true, false, false, true, true, true, false) {
		t.Fatal("must not promote without recovery PASS")
	}
	sc = EvaluateTrust(sc, true, false, false, true, true, true, true)
	if !MayPromoteDemoTrusted(true, true, false, false, true, true, true, true) {
		t.Fatal("all gates should promote")
	}
}

func TestPackageDirCreated(t *testing.T) {
	dir := t.TempDir()
	r := NewRecorder(filepath.Join(dir, "strategy-trades"), "X")
	_ = r.PersistPreSignal(FreezePreSignal(PreSignalSnapshot{SignalID: "S", StrategyTradeID: "X", Timestamp: time.Now().UTC()}))
	if _, err := os.Stat(filepath.Join(dir, "strategy-trades", "X", "pre_signal.json")); err != nil {
		t.Fatal(err)
	}
}
