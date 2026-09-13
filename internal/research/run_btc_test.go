package research

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRunFromZips(t *testing.T) {
	dir := t.TempDir()
	zp := filepath.Join(dir, "d.zip")
	f, err := os.Create(zp)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("x.csv")
	// two minutes of trades
	_, _ = w.Write([]byte("1,100,1,1,1,1690000000000,false\n2,101,1,2,2,1690000060000,true\n3,102,1,3,3,1690000120000,false\n"))
	_ = zw.Close()
	_ = f.Close()
	res, err := RunFromZips(context.Background(), []string{zp}, "BTCUSDT")
	if err != nil {
		t.Fatal(err)
	}
	if res.TradeEvents < 2 {
		t.Fatalf("%+v", res)
	}
}
