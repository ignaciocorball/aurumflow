package collector

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRotateWrite(t *testing.T) {
	dir := t.TempDir()
	w := &RotatingWriter{Dir: dir, Name: "ev"}
	if err := w.Write(map[string]any{"n": 1}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	ents, _ := os.ReadDir(dir)
	if len(ents) == 0 {
		t.Fatal("no file")
	}
	f, err := os.Open(filepath.Join(dir, ents[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(gz)
	if len(b) == 0 {
		t.Fatal("empty")
	}
}
