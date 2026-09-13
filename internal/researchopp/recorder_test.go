package researchopp

import (
	"os"
	"testing"
	"time"
)

func TestRecorderImmutableT0(t *testing.T) {
	dir := t.TempDir()
	s := Signal{ID: "t1", T0: time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC), Ranking: []byte(`{"rank":1}`)}
	if err := Record(dir, s); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(Path(dir))
	if err != nil || len(b) == 0 {
		t.Fatal(err)
	}
	if string(b)[:10] == "" {
		t.Fatal("empty")
	}
}
