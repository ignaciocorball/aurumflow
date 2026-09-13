package ops

import "testing"

func TestFormatNoFakeZero(t *testing.T) {
	if FormatPrice(false, 0, 2) != Missing || FormatPrice(true, 0, 2) != Missing {
		t.Fatal("fake zero")
	}
	if FormatPrice(true, 77304.4, 1) != "77,304.4" {
		t.Fatal(FormatPrice(true, 77304.4, 1))
	}
	if FormatPct(true, 0.0042) != "+0.42%" {
		t.Fatal(FormatPct(true, 0.0042))
	}
	if FormatMs(false, 0) != Missing || FormatMoney(true, 300) != "$300.00" {
		t.Fatal("money")
	}
}

func TestEmptyAndHealth(t *testing.T) {
	if EmptyState("CLOSED") != "MARKET CLOSED" || EmptyState("TRADEABLE") != "AWAITING TRADEABLE" {
		t.Fatal("empty")
	}
	if HealthTone("DEGRADED") != "DEGRADED" || HealthTone("HEALTHY") != "HEALTHY" || HealthTone("UNSYNCED") != "FAILED" {
		t.Fatal(HealthTone("x"))
	}
	if MilestoneCaption() != "NOT VALIDATION THRESHOLDS" {
		t.Fatal("milestone")
	}
	if GateMark(true) != "✓" || GateMark(false) != "○" {
		t.Fatal("gate")
	}
	if len(WhyLines("a\n\nb\nc\nd\ne\nf", 5)) != 5 {
		t.Fatal("why cap")
	}
}
