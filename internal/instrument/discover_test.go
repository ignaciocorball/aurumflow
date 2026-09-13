package instrument

import "testing"

func TestMapHitsDoesNotInvent(t *testing.T) {
	got := MapHits([]SearchHit{
		{Epic: "GOLD", Name: "Gold", Status: "CLOSED"},
		{Epic: "", Name: "Gold", Status: "CLOSED"},
		{Epic: "ZZ", Name: "Mystery", Status: "TRADEABLE"},
	})
	if len(got) != 1 || got[0].Epic != "GOLD" {
		t.Fatal(got)
	}
}
