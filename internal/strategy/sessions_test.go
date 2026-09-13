package strategy

import (
	"testing"
	"time"
)

func TestCanTrade(t *testing.T) {
	// Test London session (08:00-17:00 UTC)
	londonTime := time.Date(2026, 2, 6, 10, 0, 0, 0, time.UTC) // 10:00 UTC = London session
	if !CanTrade(londonTime, []string{"LONDON"}) {
		t.Error("should allow trading during London session")
	}
	if CanTrade(londonTime, []string{"NY"}) {
		t.Error("should not allow trading during London if only NY is allowed")
	}
	if !CanTrade(londonTime, []string{"ALL"}) {
		t.Error("should allow trading when ALL is specified")
	}
	if !CanTrade(londonTime, []string{"LONDON", "NY"}) {
		t.Error("should allow trading when London is in allowed list")
	}

	// Test NY session (13:00-22:00 UTC)
	nyTime := time.Date(2026, 2, 6, 15, 0, 0, 0, time.UTC) // 15:00 UTC = NY session
	if !CanTrade(nyTime, []string{"NY"}) {
		t.Error("should allow trading during NY session")
	}
	if !CanTrade(nyTime, []string{"LONDON", "NY"}) {
		t.Error("should allow trading when NY is in allowed list")
	}

	// Test overlap (London + NY)
	overlapTime := time.Date(2026, 2, 6, 14, 0, 0, 0, time.UTC) // 14:00 UTC = both sessions
	if !CanTrade(overlapTime, []string{"LONDON"}) {
		t.Error("should allow trading during overlap if London is allowed")
	}
	if !CanTrade(overlapTime, []string{"NY"}) {
		t.Error("should allow trading during overlap if NY is allowed")
	}

	// Test off-hours (outside all sessions)
	offHoursTime := time.Date(2026, 2, 6, 23, 0, 0, 0, time.UTC) // 23:00 UTC = off-hours
	if CanTrade(offHoursTime, []string{"LONDON", "NY"}) {
		t.Error("should not allow trading outside sessions")
	}
	if !CanTrade(offHoursTime, []string{"ALL"}) {
		t.Error("should allow trading when ALL is specified even off-hours")
	}

	// Test Asia session (00:00-08:00 UTC)
	asiaTime := time.Date(2026, 2, 6, 3, 0, 0, 0, time.UTC) // 03:00 UTC = Asia session
	if !CanTrade(asiaTime, []string{"ASIA"}) {
		t.Error("should allow trading during Asia session")
	}
}

func TestGetCurrentSession(t *testing.T) {
	londonTime := time.Date(2026, 2, 6, 10, 0, 0, 0, time.UTC)
	sessions := GetCurrentSession(londonTime)
	if len(sessions) == 0 || sessions[0] != SessionLondon {
		t.Errorf("expected London session, got %v", sessions)
	}

	nyTime := time.Date(2026, 2, 6, 15, 0, 0, 0, time.UTC)
	sessions = GetCurrentSession(nyTime)
	foundNY := false
	for _, s := range sessions {
		if s == SessionNY {
			foundNY = true
			break
		}
	}
	if !foundNY {
		t.Errorf("expected NY session, got %v", sessions)
	}

	overlapTime := time.Date(2026, 2, 6, 14, 0, 0, 0, time.UTC)
	sessions = GetCurrentSession(overlapTime)
	if len(sessions) < 2 {
		t.Errorf("expected overlap (London+NY), got %v", sessions)
	}
}

func TestGetSessionInfo(t *testing.T) {
	londonTime := time.Date(2026, 2, 6, 10, 0, 0, 0, time.UTC)
	info := GetSessionInfo(londonTime)
	if info != SessionLondon {
		t.Errorf("expected 'LONDON', got '%s'", info)
	}

	overlapTime := time.Date(2026, 2, 6, 14, 0, 0, 0, time.UTC)
	info = GetSessionInfo(overlapTime)
	if info != SessionLondonNY {
		t.Errorf("expected SessionLondonNY (%s), got '%s'", SessionLondonNY, info)
	}

	offHoursTime := time.Date(2026, 2, 6, 23, 0, 0, 0, time.UTC)
	info = GetSessionInfo(offHoursTime)
	if info != "OFF_HOURS" {
		t.Errorf("expected 'OFF_HOURS', got '%s'", info)
	}
}

func TestNextSessionStart(t *testing.T) {
	// (1) OFF_HOURS 23:00 UTC, allowed [ASIA, LONDON, NY] → ASIA tomorrow 00:00
	now := time.Date(2026, 2, 6, 23, 0, 0, 0, time.UTC)
	session, at, ok := NextSessionStart(now, []string{"ASIA", "LONDON", "NY"})
	if !ok || session != SessionAsia {
		t.Errorf("expected ASIA, ok=true, got session=%q ok=%v", session, ok)
	}
	expectAt := time.Date(2026, 2, 7, 0, 0, 0, 0, time.UTC)
	if !at.Equal(expectAt) {
		t.Errorf("expected next at %v, got %v", expectAt, at)
	}

	// (2) OFF_HOURS 22:00 UTC, allowed [ASIA, LONDON, NY] → ASIA 00:00 next day
	now = time.Date(2026, 2, 6, 22, 0, 0, 0, time.UTC)
	session, at, ok = NextSessionStart(now, []string{"ASIA", "LONDON", "NY"})
	if !ok || session != SessionAsia {
		t.Errorf("expected ASIA, ok=true, got session=%q ok=%v", session, ok)
	}
	expectAt = time.Date(2026, 2, 7, 0, 0, 0, 0, time.UTC)
	if !at.Equal(expectAt) {
		t.Errorf("expected next at %v, got %v", expectAt, at)
	}

	// (3) allowed ["ALL"] → same as all sessions
	now = time.Date(2026, 2, 6, 23, 0, 0, 0, time.UTC)
	session, at, ok = NextSessionStart(now, []string{"ALL"})
	if !ok || session != SessionAsia {
		t.Errorf("expected ASIA with ALL, got session=%q ok=%v", session, ok)
	}

	// (4) allowed only ["NY"], now 10:00 UTC (off-hours before London) → NY 13:00 same day
	now = time.Date(2026, 2, 6, 10, 0, 0, 0, time.UTC)
	session, at, ok = NextSessionStart(now, []string{"NY"})
	if !ok || session != SessionNY {
		t.Errorf("expected NY, ok=true, got session=%q ok=%v", session, ok)
	}
	expectAt = time.Date(2026, 2, 6, 13, 0, 0, 0, time.UTC)
	if !at.Equal(expectAt) {
		t.Errorf("expected next at %v, got %v", expectAt, at)
	}

	// (5) now already in session (London 10:00) → next session NY 13:00 same day
	now = time.Date(2026, 2, 6, 10, 0, 0, 0, time.UTC)
	session, at, ok = NextSessionStart(now, []string{"LONDON", "NY"})
	if !ok || session != SessionNY {
		t.Errorf("expected NY as next when in London, got session=%q ok=%v", session, ok)
	}
	expectAt = time.Date(2026, 2, 6, 13, 0, 0, 0, time.UTC)
	if !at.Equal(expectAt) {
		t.Errorf("expected next at %v, got %v", expectAt, at)
	}

	// allowed empty → ok=false
	_, _, ok = NextSessionStart(time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC), nil)
	if ok {
		t.Error("expected ok=false when allowed is nil")
	}
	_, _, ok = NextSessionStart(time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC), []string{})
	if ok {
		t.Error("expected ok=false when allowed is empty")
	}
}

func TestEvaluateSessionEmptyMeansAll(t *testing.T) {
	off := time.Date(2026, 9, 13, 22, 30, 0, 0, time.UTC)
	ev := EvaluateSession(off, nil)
	if ev.ClockSession != "OFF_HOURS" || ev.ConfigPolicy != PolicyAll || !ev.Eligible || ev.Reason != ReasonEmptyMeansAll {
		t.Fatalf("%+v", ev)
	}
	label, _ := NextEligibleLabel(ev, off, nil)
	if label != NextEligibleNA {
		t.Fatal(label)
	}
	ev = EvaluateSession(off, []string{"ALL"})
	if !ev.Eligible || ev.Reason != ReasonPolicyAll {
		t.Fatalf("%+v", ev)
	}
}

func TestEvaluateSessionExplicitAndBoundaries(t *testing.T) {
	asia := time.Date(2026, 9, 14, 3, 0, 0, 0, time.UTC)
	if ev := EvaluateSession(asia, []string{"ASIA"}); !ev.Eligible || ev.ClockSession != SessionAsia {
		t.Fatalf("%+v", ev)
	}
	londonOpen := time.Date(2026, 9, 14, 8, 0, 0, 0, time.UTC)
	if ev := EvaluateSession(londonOpen, []string{"LONDON"}); !ev.Eligible {
		t.Fatal("london open")
	}
	londonEnd := time.Date(2026, 9, 14, 17, 0, 0, 0, time.UTC)
	if ev := EvaluateSession(londonEnd, []string{"LONDON"}); ev.Eligible {
		t.Fatal("london end exclusive")
	}
	ny := time.Date(2026, 9, 14, 21, 59, 0, 0, time.UTC)
	if ev := EvaluateSession(ny, []string{"NY"}); !ev.Eligible {
		t.Fatal("ny")
	}
	multi := EvaluateSession(time.Date(2026, 9, 14, 14, 0, 0, 0, time.UTC), []string{"LONDON", "NY"})
	if !multi.Eligible {
		t.Fatal(multi)
	}
	rollover := time.Date(2026, 9, 13, 23, 30, 0, 0, time.UTC)
	if ev := EvaluateSession(rollover, []string{"LONDON", "NY"}); ev.Eligible {
		t.Fatal("restricted off hours")
	}
}
