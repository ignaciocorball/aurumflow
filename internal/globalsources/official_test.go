package globalsources

import (
	"testing"
	"time"

	"aurumflow/internal/worlddomain"
)

func TestFredLatestParsing(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	raw := []byte("DATE,WALCL\n2026-08-27,6700\n2026-09-03,6650\n")
	xs := ParseFredSeries("WALCL", raw, now)
	o, ok := Latest(xs, "FED_H41", "FED_ASSETS", now)
	if !ok || o.Value != 6650 || o.ObservedAt.Format("2006-01-02") != "2026-09-03" {
		t.Fatalf("%+v ok=%v", o, ok)
	}
	if o.SeriesID != "WALCL" {
		t.Fatal(o.SeriesID)
	}
}

func TestCboeOfficialParsing(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	raw := []byte("DATE,CALL,PUT,TOTAL,P/C Ratio\n09/11/2026,100,80,180,0.80\n")
	xs := ParseCboeOfficial(raw, "TOTAL_PC", now)
	o, ok := Latest(xs, "CBOE", "TOTAL_PC", now)
	if !ok || o.Value != 0.80 {
		t.Fatalf("%+v", xs)
	}
}

func TestICICurrentParsing(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	xs := ParseICI(mustRead(t, "ici.csv"), now)
	if _, ok := Latest(xs, "ICI", "EQUITY", now); !ok {
		t.Fatal("ici")
	}
}

func TestJPXBothFormats(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	v1 := ParseJPXAny(mustRead(t, "jpx_v1.csv"), now)
	if _, ok := Latest(v1, "JPX", "FOREIGN", now); !ok {
		t.Fatal("v1")
	}
	v2 := ParseJPXAny(mustRead(t, "jpx_v2.csv"), now)
	o, ok := Latest(v2, "JPX", "FOREIGN", now)
	if !ok || o.ParserVer != "post-April-2026" {
		t.Fatalf("%+v", v2)
	}
}

func TestRawHashReproducible(t *testing.T) {
	a := HashBytes([]byte("official"))
	b := HashBytes([]byte("official"))
	if a == "" || a != b {
		t.Fatal(a, b)
	}
}

func TestTICSLTAndFedHTML(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	tic := []byte("Grand Total\t99996\t2026-06\t39189956\t207076\t-1\n")
	xs := ParseTICSLTTable1(tic, now)
	o, ok := Latest(xs, "TIC", "NET_LT_SECURITIES", now)
	if !ok || o.Period != "2026-06" || o.AvailableAt.After(now) {
		t.Fatalf("%+v", xs)
	}
	html := []byte("<title>H.4.1 - September 10, 2026</title>Total assets 6,740,619 Reserve balances with Federal Reserve Banks 3,200,000 U.S. Treasury, General Account 800,000 Reverse repurchase agreements 250,000")
	fed := ParseFedH41HTML(html, now)
	if _, ok := Latest(fed, "FED_H41", "FED_ASSETS", now); !ok {
		t.Fatalf("%+v", fed)
	}
}

func TestTICReleaseChronology(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	raw := []byte("DATE,NET_LT,FOREIGN_PURCHASES_US,US_PURCHASES_FOREIGN\n2026-05,10,1,1\n2026-06,8,1,1\n")
	xs := ParseTICAny(raw, now)
	last, ok := Latest(xs, "TIC", "NET_LT_SECURITIES", now)
	if !ok || last.ObservedAt.After(now) || last.AvailableAt.After(now) {
		t.Fatalf("%+v", last)
	}
	if last.Period != "2026-06" {
		t.Fatal(last.Period)
	}
}

func TestCacheReloadStampsOfficial(t *testing.T) {
	dir := t.TempDir()
	oldRaw, oldNorm := RawRoot, NormalizedRoot
	RawRoot, NormalizedRoot = dir+"/raw", dir+"/norm"
	defer func() { RawRoot, NormalizedRoot = oldRaw, oldNorm }()
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	obs := StampOrigin(ParseFedH41(mustRead(t, "fed.csv"), now), worlddomain.OriginCache, "abc")
	if err := WriteNormalized("FED_H41", obs, CacheMeta{Provider: "FED_H41", Origin: "CACHE_OFFICIAL", Hash: "abc"}); err != nil {
		t.Fatal(err)
	}
	got, meta, err := LoadNormalized("FED_H41")
	if err != nil || len(got) == 0 || meta.Hash != "abc" {
		t.Fatal(err, len(got), meta)
	}
	if got[0].Origin != worlddomain.OriginCache {
		t.Fatal(got[0].Origin)
	}
}
