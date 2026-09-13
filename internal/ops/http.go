package ops

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	mu     sync.RWMutex
	status Status
	http   *http.Server
	addr   string
}

func NewServer(addr string, initial Status) *Server {
	if addr == "" {
		addr = "127.0.0.1:8765"
	}
	s := &Server{status: initial, addr: addr}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthz)
	mux.HandleFunc("/status", s.statusHandler)
	s.http = &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return s
}

func (s *Server) Set(st Status) {
	s.mu.Lock()
	s.status = st
	s.mu.Unlock()
}

func (s *Server) Get() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) statusHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.Get())
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.addr = ln.Addr().String()
	s.http.Addr = s.addr
	return s.http.Serve(ln)
}

func (s *Server) Addr() string { return s.addr }

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
