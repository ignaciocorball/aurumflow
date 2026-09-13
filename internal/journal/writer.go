package journal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// JournalWriter writes journal events (JSONL). All methods are safe for concurrent use.
type JournalWriter interface {
	WriteSetupEvaluated(SetupEvaluated) error
	WriteSignalRejected(SignalRejected) error
	WriteSignalGenerated(SignalGenerated) error
	WriteOrderAttempt(OrderAttempt) error
	WriteOrderResult(OrderResult) error
	WritePositionClosed(PositionClosed) error
	WriteLifecycle(Lifecycle) error
	Close() error
}

// FileWriter writes one JSON line per event to a file.
type FileWriter struct {
	path string
	f    *os.File
	mu   sync.Mutex
}

// NewFileWriter creates a FileWriter that appends to path. Parent directory is created if needed.
func NewFileWriter(path string) (*FileWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileWriter{path: path, f: f}, nil
}

func (w *FileWriter) writeLine(obj interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.f.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	return w.f.Sync()
}

func (w *FileWriter) WriteSetupEvaluated(e SetupEvaluated) error {
	if e.Ts == "" {
		e.Ts = ts()
	}
	if e.Event == "" {
		e.Event = EventSetupEvaluated
	}
	return w.writeLine(e)
}

func (w *FileWriter) WriteSignalRejected(e SignalRejected) error {
	if e.Ts == "" {
		e.Ts = ts()
	}
	if e.Event == "" {
		e.Event = EventSignalRejected
	}
	return w.writeLine(e)
}

func (w *FileWriter) WriteSignalGenerated(e SignalGenerated) error {
	if e.Ts == "" {
		e.Ts = ts()
	}
	if e.Event == "" {
		e.Event = EventSignalGenerated
	}
	return w.writeLine(e)
}

func (w *FileWriter) WriteOrderAttempt(e OrderAttempt) error {
	if e.Ts == "" {
		e.Ts = ts()
	}
	if e.Event == "" {
		e.Event = EventOrderAttempt
	}
	return w.writeLine(e)
}

func (w *FileWriter) WriteOrderResult(e OrderResult) error {
	if e.Ts == "" {
		e.Ts = ts()
	}
	if e.Event == "" {
		e.Event = EventOrderResult
	}
	return w.writeLine(e)
}

func (w *FileWriter) WritePositionClosed(e PositionClosed) error {
	if e.Ts == "" {
		e.Ts = ts()
	}
	if e.Event == "" {
		e.Event = EventPositionClosed
	}
	return w.writeLine(e)
}

func (w *FileWriter) WriteLifecycle(e Lifecycle) error {
	if e.Ts == "" {
		e.Ts = ts()
	}
	return w.writeLine(e)
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}
