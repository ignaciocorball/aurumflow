package notifications

import (
	"testing"
	"time"
)

func TestStore_ShouldDedup(t *testing.T) {
	store := NewStore(100*time.Millisecond, 200*time.Millisecond)
	ev := NotifEvent{Type: TypeHeartbeat, Payload: map[string]any{}}
	key := DedupKeyHash("HEARTBEAT", "ETHUSD")
	if store.ShouldDedup(key, ev) {
		t.Error("first time should not dedup")
	}
	store.RecordSend(key, TypeHeartbeat, ev)
	if !store.ShouldDedup(key, ev) {
		t.Error("second time within window should dedup")
	}
	time.Sleep(150 * time.Millisecond)
	if store.ShouldDedup(key, ev) {
		t.Error("after window should not dedup")
	}
}

func TestStore_Throttled(t *testing.T) {
	store := NewStore(time.Second, time.Second)
	if store.Throttled(TypeHeartbeat, 100*time.Millisecond) {
		t.Error("first time should not be throttled")
	}
	store.RecordSend("k", TypeHeartbeat, NotifEvent{})
	if !store.Throttled(TypeHeartbeat, 100*time.Millisecond) {
		t.Error("just sent should be throttled")
	}
	time.Sleep(150 * time.Millisecond)
	if store.Throttled(TypeHeartbeat, 100*time.Millisecond) {
		t.Error("after interval should not be throttled")
	}
}

func TestStore_AddReject_FlushRejects(t *testing.T) {
	store := NewStore(time.Second, time.Second)
	ev := NotifEvent{
		Type: TypeSignalRejected,
		Payload: map[string]any{
			"reason_code": RejH1Filter,
			"direction":   "BUY",
			"score":       8,
			"confidence":  0.8,
		},
	}
	store.AddReject(ev)
	store.AddReject(ev)
	total, byReason, lastDir, lastScore, lastConf := store.FlushRejects()
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if byReason[RejH1Filter] != 2 {
		t.Errorf("byReason[%s] = %d, want 2", RejH1Filter, byReason[RejH1Filter])
	}
	if lastDir != "BUY" || lastScore != 8 || lastConf != 0.8 {
		t.Errorf("last dir=%s score=%d conf=%.2f", lastDir, lastScore, lastConf)
	}
	total2, _, _, _, _ := store.FlushRejects()
	if total2 != 0 {
		t.Errorf("after flush total = %d, want 0", total2)
	}
}

func TestStore_SweepTypeChanged(t *testing.T) {
	store := NewStore(time.Second, time.Second)
	if store.SweepTypeChanged("BUY_SIDE") {
		t.Error("first type should not be 'changed'")
	}
	if store.SweepTypeChanged("BUY_SIDE") {
		t.Error("same type should not be changed")
	}
	if !store.SweepTypeChanged("SELL_SIDE") {
		t.Error("BUY_SIDE -> SELL_SIDE should be changed")
	}
}
