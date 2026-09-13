package binanceusdm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"aurumflow/internal/book"
)

func TestFetchSnapshotAndApply(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/depth" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"lastUpdateId":10,"bids":[["100","2"]],"asks":[["101","3"]]}`))
	}))
	defer srv.Close()
	a := New()
	a.RESTBase = srv.URL
	a.HTTP = srv.Client()
	snap, err := a.FetchSnapshot(context.Background())
	if err != nil || snap.LastUpdateID != 10 {
		t.Fatal(err, snap)
	}
	a.applySnapshot(snap)
	bid, ask, ok := a.Book.BestBidAsk()
	if !ok || bid != 100 || ask != 101 {
		t.Fatalf("%v %v %v", bid, ask, ok)
	}
	if !a.Book.Synced {
		t.Fatal("synced")
	}
	_ = book.Level{}
}

func TestFetchAggTrades(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fapi/v1/aggTrades" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"a":9,"p":"1.5","q":"2","m":true,"T":1}]`))
	}))
	defer srv.Close()
	a := New()
	a.RESTBase = srv.URL
	a.HTTP = srv.Client()
	trs, err := a.FetchAggTrades(context.Background())
	if err != nil || len(trs) != 1 || trs[0].ID != 9 || !trs[0].BuyerMaker {
		t.Fatalf("%v %+v", err, trs)
	}
}
