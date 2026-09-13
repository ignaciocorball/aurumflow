package binancehist

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseAggAndAggressor(t *testing.T) {
	tr, err := ParseAggRow([]string{"26129", "0.01633102", "4.70443515", "27781", "27781", "1498793709153", "true"})
	if err != nil || tr.ID != 26129 || !tr.BuyerMaker || Aggressor(tr.BuyerMaker) != "aggressive_sell" {
		t.Fatalf("%+v %v", tr, err)
	}
	if _, err := ParseAggRow([]string{"1", "0", "1", "1", "1", "1", "false"}); err == nil {
		t.Fatal("zero price")
	}
}

func TestChecksumAndDays(t *testing.T) {
	if !ParseChecksumLine("abc123  file.zip", "abc123") {
		t.Fatal("checksum")
	}
	ds := DaysInclusive(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
	if len(ds) != 3 || ds[0] != "2026-08-01" || ds[2] != "2026-08-03" {
		t.Fatal(ds)
	}
}

func TestIterZipCSVAndQuality(t *testing.T) {
	dir := t.TempDir()
	zp := filepath.Join(dir, "t.zip")
	f, err := os.Create(zp)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.Create("x.csv")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("agg_trade_id,price,quantity,first_trade_id,last_trade_id,transact_time,is_buyer_maker\n1,10,1,1,1,1000,false\n2,11,1,2,2,1001,true\n2,11,1,2,2,1001,true\n"))
	_ = zw.Close()
	_ = f.Close()
	q, err := (&Archive{}).ScanDay(zp)
	if err != nil || q.Rows != 3 || q.Duplicates < 1 || q.Invalid != 0 || q.Headers != 1 {
		t.Fatalf("header must not be invalid: %+v %v", q, err)
	}
}

func TestHeaderNotInvalid(t *testing.T) {
	if !IsHeaderRow([]string{"agg_trade_id", "price", "quantity", "first_trade_id", "last_trade_id", "transact_time", "is_buyer_maker"}) {
		t.Fatal("underscore header")
	}
	if IsHeaderRow([]string{"26129", "0.01", "1", "1", "1", "1498793709153", "true"}) {
		t.Fatal("data row")
	}
	_ = context.Background()
}

func TestResumeSkipExisting(t *testing.T) {
	dir := t.TempDir()
	a := New(dir, "BTCUSDT")
	p := a.DayPath("2026-08-01")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := a.DownloadDay(context.Background(), "2026-08-01")
	if err != nil || got != p {
		t.Fatal(err, got)
	}
}
