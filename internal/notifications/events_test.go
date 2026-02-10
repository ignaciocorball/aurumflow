package notifications

import (
	"testing"
	"time"
)

func TestNormalizeReason(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"  H1_FILTER  ", "H1_FILTER"},
		{"h1 filter", "H1_FILTER"},
		{"score_after_context", "SCORE_AFTER_CONTEXT"},
		{"", ""},
	}
	for _, tt := range tests {
		got := NormalizeReason(tt.in)
		if got != tt.want {
			t.Errorf("NormalizeReason(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestReasonToCode(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"H1_FILTER", RejH1Filter},
		{"H1_RANGE_BLOCKED", RejH1RangeBlocked},
		{"NO_SIGNAL", RejNoSignal},
		{"score_after_context", RejScoreAfterContext},
		{"M5_TIMING_REJECT", RejM5TimingReject},
		{"unknown_reason_xyz", RejUnknown + ":UNKNOWN_REASON_XYZ"},
		{"", RejNoSignal},
	}
	for _, tt := range tests {
		got := ReasonToCode(tt.in)
		if got != tt.want {
			t.Errorf("ReasonToCode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDedupKeyHash(t *testing.T) {
	a := DedupKeyHash("HEARTBEAT", "ETHUSD", "NY")
	b := DedupKeyHash("HEARTBEAT", "ETHUSD", "NY")
	if a != b {
		t.Errorf("DedupKeyHash should be deterministic: %q != %q", a, b)
	}
	c := DedupKeyHash("HEARTBEAT", "ETHUSD", "LONDON")
	if a == c {
		t.Errorf("DedupKeyHash should differ for different inputs")
	}
}

func TestNotifEvent_ZeroTimestamp(t *testing.T) {
	ev := NotifEvent{Type: TypeHeartbeat, Category: CategoryHealth}
	if !ev.Timestamp.IsZero() {
		t.Errorf("expected zero timestamp")
	}
	ev.Timestamp = time.Now().UTC()
	if ev.Timestamp.IsZero() {
		t.Errorf("expected non-zero after set")
	}
}
