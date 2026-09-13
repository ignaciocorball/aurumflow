package okxswap

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aurumflow/internal/md"
)

func TestFetchSnapshotPublic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/market/books" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":"0","data":[{"seqId":"9","bids":[["100","2"]],"asks":[["101","3"]]}]}`))
	}))
	defer srv.Close()
	a := New()
	a.RESTBase = srv.URL
	a.HTTP = srv.Client()
	seq, bids, asks, err := a.FetchSnapshot(context.Background())
	if err != nil || seq != 9 || len(bids) != 1 || asks[0].Qty != 3 {
		t.Fatalf("%v %d %v %v", err, seq, bids, asks)
	}
	if a.Capabilities()&md.CapBookDelta == 0 {
		t.Fatal("caps")
	}
	_ = time.Second
}

func TestFetchInstrumentVerified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v5/public/instruments" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"code":"0","data":[{"instId":"BTC-USDT-SWAP","instType":"SWAP","state":"live"}]}`))
	}))
	defer srv.Close()
	a := New()
	a.RESTBase = srv.URL
	a.HTTP = srv.Client()
	id, err := a.FetchInstrument(context.Background())
	if err != nil || id != InstID {
		t.Fatal(err, id)
	}
}
