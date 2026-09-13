package finra

import "testing"

func TestMissingNotZero(t *testing.T) {
	_, _, _, _, _, _, miss := Feature(nil, "AAPL", "2026-08-07")
	if !miss {
		t.Fatal("missing")
	}
	weeks := []Week{
		ParseWeek("AAPL", "2026-08-01", 10, 5, true),
		ParseWeek("AAPL", "2026-08-08", 20, 5, true),
	}
	ats, _, tot, ch, _, _, miss := Feature(weeks, "AAPL", "2026-08-08")
	if miss || ats != 20 || tot != 25 || ch != 10 {
		t.Fatalf("%v %v %v %v", ats, tot, ch, miss)
	}
}
