package terminal

import (
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

type Server struct {
	http *http.Server
	addr string
	hub  *Hub
}

func NewServer(addr string, hub *Hub) *Server {
	if addr == "" {
		addr = "127.0.0.1:8770"
	}
	s := &Server{addr: addr, hub: hub}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.healthz)
	mux.HandleFunc("/ui/snapshot", s.snapshot)
	mux.HandleFunc("/ui/events", s.events)
	mux.HandleFunc("/", s.spa)
	s.http = &http.Server{Addr: addr, Handler: rejectMutations(mux), ReadHeaderTimeout: 5 * time.Second}
	return s
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

func (s *Server) Shutdown() error {
	if s.http == nil {
		return nil
	}
	return s.http.Close()
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{"ok": true, "mutation": BrokerMutationCapability()})
}

func (s *Server) snapshot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.hub.Snapshot())
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no stream", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch := s.hub.Subscribe()
	defer s.hub.Unsubscribe(ch)
	keep := time.NewTicker(15 * time.Second)
	defer keep.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-keep.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			fl.Flush()
		case ev, ok := <-ch:
			if !ok {
				return
			}
			raw, _ := json.Marshal(ev)
			_, _ = w.Write([]byte("event: " + ev.Type + "\n"))
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(raw)
			_, _ = w.Write([]byte("\n\n"))
			fl.Flush()
		}
	}
}

func (s *Server) spa(w http.ResponseWriter, r *http.Request) {
	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "" || strings.HasSuffix(p, "/") {
		p = "index.html"
	}
	b, err := fs.ReadFile(WebFS(), path.Join("webdist", p))
	if err != nil {
		b, err = fs.ReadFile(WebFS(), "webdist/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		p = "index.html"
	}
	w.Header().Set("Content-Type", contentType(p))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(b)
}

func contentType(p string) string {
	switch path.Ext(p) {
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".json":
		return "application/json"
	case ".woff2":
		return "font/woff2"
	default:
		return "text/html; charset=utf-8"
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}
