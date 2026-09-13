package secctx

import "testing"

func TestWatchlistRequiresResolvedCIK(t *testing.T) {
	if n := ResolveWatchlist(nil); len(n) != 0 {
		t.Fatal(n)
	}
	got := ResolveWatchlist(map[string]string{"blackrock": "1364742", "vanguard": "guess"})
	if len(got) != 1 || got[0].CIK != "0001364742" {
		t.Fatalf("%+v", got)
	}
}
