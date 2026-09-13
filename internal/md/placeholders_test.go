package md

import "testing"

func TestPaidFeedsNotConnected(t *testing.T) {
	st := PaidFeedStatuses()
	for _, k := range []string{"CME_NQ", "CME_GC", "NASDAQ_TOTALVIEW", "OPTIONS"} {
		if st[k] != StatusNotConnected {
			t.Fatalf("%s=%s", k, st[k])
		}
	}
	if CME_NQ.Capabilities() != 0 {
		t.Fatal("placeholders must not fake capabilities")
	}
}
