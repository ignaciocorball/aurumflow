package instrument

import "testing"

func TestResolveUS500UniqueOnly(t *testing.T) {
	_, ok := ResolveUS500([]CatalogHit{
		{Epic: "SPX1", Name: "S&P 500", InstrumentType: "INDICES", Currency: "USD", Country: "US"},
		{Epic: "US500", Name: "US500", InstrumentType: "INDICES", Currency: "USD", Country: "US"},
	})
	if ok {
		t.Fatal("must not pick first of two")
	}
	m, ok := ResolveUS500([]CatalogHit{
		{Epic: "US500", Name: "US500", InstrumentType: "INDICES", Currency: "USD", Country: "US"},
		{Epic: "US500", Name: "US 500", InstrumentType: "INDICES", Currency: "USD", Country: "US"},
		{Epic: "SPX200", Name: "S&P 200", InstrumentType: "INDICES", Currency: "USD", Country: "US"},
	})
	if !ok || m.Epic != "US500" {
		t.Fatal("unique US500 after epic dedup")
	}
	m, ok = ResolveUS500([]CatalogHit{
		{Epic: "FOO", Name: "Something", InstrumentType: "SHARES", Currency: "EUR"},
	})
	if ok || m.Epic != "" {
		t.Fatal("unresolved invented")
	}
}

func TestMapHitsLeavesUS500UnresolvedWhenAmbiguous(t *testing.T) {
	got := MapHits([]SearchHit{
		{Epic: "GOLD", Name: "Gold", Status: "CLOSED"},
		{Epic: "SP1", Name: "S&P 500", InstrumentType: "INDICES", Currency: "USD"},
		{Epic: "SP2", Name: "S&P 500", InstrumentType: "INDICES", Currency: "USD"},
	})
	var us Mapping
	for _, m := range got {
		if m.Canonical == "US500" {
			us = m
		}
	}
	if us.Epic != "" || !us.Unresolved {
		t.Fatalf("%+v", got)
	}
}

func TestCanonicalGuessDE40NotEuropeBlob(t *testing.T) {
	id, _, _ := GuessCanonical("Germany 40", "DE40")
	if id != "DE40" {
		t.Fatal(id)
	}
}

func TestIndexEpicsNotEquityShares(t *testing.T) {
	if id, _, _ := GuessCanonical("Nasdaq", "NDAQ"); id == "US100" {
		t.Fatal("NDAQ share must not become US100")
	}
	if id, _, _ := GuessCanonical("Dow Inc.", "DOW"); id == "US30" {
		t.Fatal("DOW share must not become US30")
	}
	got := MapHits([]SearchHit{
		{Epic: "NDAQ", Name: "Nasdaq", InstrumentType: "SHARES"},
		{Epic: "US100", Name: "US 100", InstrumentType: "INDICES"},
		{Epic: "DOW", Name: "Dow Inc.", InstrumentType: "SHARES"},
		{Epic: "US30", Name: "US 30", InstrumentType: "INDICES"},
	})
	m := map[string]string{}
	for _, x := range got {
		m[x.Canonical] = x.Epic
	}
	if m["US100"] != "US100" || m["US30"] != "US30" {
		t.Fatal(m)
	}
}

func TestExactUS500EpicWins(t *testing.T) {
	got := MapHits([]SearchHit{
		{Epic: "US500", Name: "US 500", InstrumentType: "INDICES", Status: "CLOSED"},
		{Epic: "SP1", Name: "S&P 500", InstrumentType: "INDICES", Status: "CLOSED"},
	})
	var us Mapping
	for _, m := range got {
		if m.Canonical == "US500" {
			us = m
		}
	}
	if us.Epic != "US500" || us.Unresolved {
		t.Fatal(us)
	}
}
