package ops

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) SetWorld(v any) {
	raw, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.world = raw
	s.worldAt = time.Now().UTC()
	s.mu.Unlock()
}

func (s *Server) WorldRaw() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.world) == 0 {
		return []byte(`{"status":"WAITING"}`)
	}
	return s.world
}

func (s *Server) apiWorld(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(s.WorldRaw())
}

func (s *Server) apiWorldSlice(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		var m map[string]any
		if json.Unmarshal(s.WorldRaw(), &m) != nil {
			writeJSON(w, map[string]any{})
			return
		}
		writeJSON(w, map[string]any{key: m[key], "as_of": m["AsOf"]})
	}
}

func (s *Server) apiSources(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	raw := append([]byte(nil), s.sources...)
	s.mu.RUnlock()
	if len(raw) == 0 {
		writeJSON(w, map[string]any{"sources": []any{}})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(raw)
}

func (s *Server) SetSources(v any) {
	raw, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.sources = raw
	s.mu.Unlock()
}

func (s *Server) apiInstitutional(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"label": "DELAYED_POSITIONING_CONTEXT",
		"note":  "13F is delayed. Not live order direction.",
		"managers": []any{},
	})
}
