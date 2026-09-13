package globalsources

import (
	"os"
	"testing"
	"time"
)

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestParsersProducePresentValues(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	bis := ParseBIS(mustRead(t, "bis.csv"), now)
	if _, ok := Latest(bis, "BIS", "GLOBAL_CREDIT_USD", now); !ok {
		t.Fatalf("bis %+v", bis)
	}
	fed := ParseFedH41(mustRead(t, "fed.csv"), now)
	if _, ok := Latest(fed, "FED_H41", "FED_ASSETS", now); !ok {
		t.Fatal("fed")
	}
	cls, chg := FedLiquidityClass(fed, "FED_ASSETS", now)
	if cls == "" || len(chg) == 0 {
		t.Fatal(cls, chg)
	}
	if _, ok := Latest(ParseECB(mustRead(t, "ecb.csv"), now), "ECB", "POLICY_RATE", now); !ok {
		t.Fatal("ecb")
	}
	if _, ok := Latest(ParseTIC(mustRead(t, "tic.csv"), now), "TIC", "NET_LT_SECURITIES", now); !ok {
		t.Fatal("tic")
	}
	ici := ParseICI(mustRead(t, "ici.csv"), now)
	eq, ok := Latest(ici, "ICI", "EQUITY", now)
	if !ok {
		t.Fatal("ici")
	}
	dir, note := ICIFlowLabel("EQUITY", eq.Value, eq.Present)
	if note != "fund/ETF capital-flow proxy" || dir == "" {
		t.Fatal(dir, note)
	}
	if _, ok := Latest(ParseJPX(mustRead(t, "jpx.csv"), now), "JPX", "FOREIGN", now); !ok {
		t.Fatal("jpx")
	}
	if _, ok := Latest(ParseCboe(mustRead(t, "cboe.csv"), now), "CBOE", "TOTAL_PC", now); !ok {
		t.Fatal("cboe")
	}
	if _, ok := Latest(ParseWGC(mustRead(t, "wgc.csv"), now), "WGC", "GOLD_ETF_FLOWS", now); !ok {
		t.Fatal("wgc")
	}
	if _, ok := Latest(ParseIShares(mustRead(t, "ishares.csv"), now), "ISHARES", "IVV_SHARES", now); !ok {
		t.Fatal("ishares")
	}
	eia := ParseEIA(mustRead(t, "eia.csv"), now)
	st, why := OilPhysicalState(eia, now)
	if st != "TIGHTENING" || len(why) == 0 {
		t.Fatal(st, why)
	}
	if HKEXStatus() != "PENDING_PUBLIC_STRUCTURED_SOURCE" {
		t.Fatal(HKEXStatus())
	}
	if (EventContextProvider{}).Status() != "INTERFACE_ONLY" {
		t.Fatal("event")
	}
	reg := NewRegistry()
	if _, ok := reg.Get("BIS"); !ok {
		t.Fatal("registry")
	}
}

func TestFreshnessAndNoLookaheadStore(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	s := &Store{}
	s.Add(ParseBIS(mustRead(t, "bis.csv"), now)...)
	if len(s.Usable(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))) != 0 {
		t.Fatal("lookahead")
	}
	if len(s.Usable(now)) == 0 {
		t.Fatal("usable empty")
	}
}
