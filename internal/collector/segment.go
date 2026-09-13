package collector

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"aurumflow/internal/catalog"
)

type SegmentMeta struct {
	Provider     string    `json:"provider"`
	Instrument   string    `json:"instrument"`
	Kind         string    `json:"kind"`
	FirstEvent   time.Time `json:"first_event_time"`
	LastEvent    time.Time `json:"last_event_time"`
	FirstSeq     int64     `json:"first_sequence"`
	LastSeq      int64     `json:"last_sequence"`
	Events       int       `json:"events"`
	Gaps         int       `json:"gaps"`
	Resyncs      int       `json:"resyncs"`
	Checksum     string    `json:"checksum"`
	CreatedAt    time.Time `json:"created_at"`
	Path         string    `json:"path"`
}

type SegmentedWriter struct {
	Root       string
	Provider   string
	Instrument string
	Kind       string
	mu         sync.Mutex
	hour       string
	inner      *RotatingWriter
	meta       SegmentMeta
	dir        string
}

func NewSegmented(root, provider, instrument, kind string) *SegmentedWriter {
	return &SegmentedWriter{Root: root, Provider: provider, Instrument: instrument, Kind: kind}
}

func (w *SegmentedWriter) WriteEvent(t time.Time, seq int64, v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	day := t.UTC().Format("2006/01/02")
	h := t.UTC().Format("15")
	dir := filepath.Join(w.Root, w.Provider, w.Instrument, day)
	key := day + h
	if w.hour != key || w.inner == nil {
		if err := w.flushLocked(); err != nil {
			return err
		}
		w.dir = dir
		w.inner = &RotatingWriter{Dir: dir, Name: h + "-" + w.Kind}
		w.meta = SegmentMeta{
			Provider: w.Provider, Instrument: w.Instrument, Kind: w.Kind,
			FirstEvent: t, FirstSeq: seq, CreatedAt: time.Now().UTC(),
		}
		w.hour = key
	}
	if err := w.inner.Write(v); err != nil {
		return err
	}
	if w.meta.FirstEvent.IsZero() {
		w.meta.FirstEvent = t
		w.meta.FirstSeq = seq
	}
	w.meta.LastEvent = t
	w.meta.LastSeq = seq
	w.meta.Events++
	return nil
}

func (w *SegmentedWriter) NoteGap() {
	w.mu.Lock()
	w.meta.Gaps++
	w.mu.Unlock()
}

func (w *SegmentedWriter) NoteResync() {
	w.mu.Lock()
	w.meta.Resyncs++
	w.mu.Unlock()
}

func (w *SegmentedWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.flushLocked()
}

func (w *SegmentedWriter) flushLocked() error {
	if w.inner == nil {
		return nil
	}
	_ = w.inner.Close()
	ents, _ := os.ReadDir(w.dir)
	var path string
	prefix := w.inner.Name
	for _, e := range ents {
		if len(e.Name()) >= len(prefix) && e.Name()[:len(prefix)] == prefix {
			path = filepath.Join(w.dir, e.Name())
		}
	}
	if path != "" {
		sum, _ := catalog.FileSHA256(path)
		if sum == "" {
			h := sha256.Sum256([]byte(path))
			sum = hex.EncodeToString(h[:])
		}
		w.meta.Checksum = sum
		w.meta.Path = path
		raw, _ := json.MarshalIndent(w.meta, "", "  ")
		_ = os.WriteFile(path+".meta.json", raw, 0o644)
	}
	w.inner = nil
	return nil
}
