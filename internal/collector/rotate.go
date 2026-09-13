package collector

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RotatingWriter struct {
	Dir   string
	Name  string
	mu    sync.Mutex
	hour  string
	f     *os.File
	gz    *gzip.Writer
	Bytes int64
	Rows  int64
}

func (w *RotatingWriter) Write(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	h := time.Now().UTC().Format("2006010215")
	if w.hour != h || w.f == nil {
		if err := w.rotate(h); err != nil {
			return err
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	n, err := w.gz.Write(b)
	w.Bytes += int64(n)
	w.Rows++
	return err
}

func (w *RotatingWriter) rotate(h string) error {
	_ = w.closeLocked()
	if err := os.MkdirAll(w.Dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(w.Dir, w.Name+"-"+h+".jsonl.gz")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.f = f
	w.gz = gzip.NewWriter(f)
	w.hour = h
	return nil
}

func (w *RotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeLocked()
}

func (w *RotatingWriter) closeLocked() error {
	if w.gz != nil {
		_ = w.gz.Close()
		w.gz = nil
	}
	if w.f != nil {
		err := w.f.Close()
		w.f = nil
		return err
	}
	return nil
}
