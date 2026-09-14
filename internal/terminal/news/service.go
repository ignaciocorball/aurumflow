package news

import (
	"net/http"
	"sync"
	"time"
)

type Fetcher interface {
	GetBytes(url, etag, modified string) ([]byte, http.Header, error)
}

type Service struct {
	mu      sync.Mutex
	store   *Store
	sources []Source
	client  Fetcher
	etag    map[string]string
	mod     map[string]string
	next    map[string]time.Time
	lastOK  time.Time
	status  string
	detail  string
}

func NewService(store *Store, client Fetcher) *Service {
	return &Service{
		store: store, sources: DefaultSources(), client: client,
		etag: map[string]string{}, mod: map[string]string{}, next: map[string]time.Time{},
		status: "OFFLINE", detail: "No poll yet",
	}
}

func (s *Service) Status() (status, detail string, last time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status, s.detail, s.lastOK
}

func (s *Service) Tick(now time.Time) bool {
	changed := false
	okAny := false
	fail := 0
	for _, src := range s.sources {
		s.mu.Lock()
		due := now.After(s.next[src.ID])
		etag, mod := s.etag[src.ID], s.mod[src.ID]
		s.mu.Unlock()
		if !due {
			continue
		}
		body, hdr, err := s.client.GetBytes(src.URL, etag, mod)
		s.mu.Lock()
		s.next[src.ID] = now.Add(src.Cadence)
		s.mu.Unlock()
		if err != nil {
			if err.Error() == "not modified" {
				okAny = true
				continue
			}
			fail++
			continue
		}
		okAny = true
		if hdr != nil {
			s.mu.Lock()
			if v := hdr.Get("ETag"); v != "" {
				s.etag[src.ID] = v
			}
			if v := hdr.Get("Last-Modified"); v != "" {
				s.mod[src.ID] = v
			}
			s.mu.Unlock()
		}
		items := ParseFeed(src.Name, body)
		if s.store.Ingest(items) > 0 {
			changed = true
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if okAny {
		s.lastOK = now
		if fail > 0 {
			s.status = "DEGRADED"
			s.detail = "Some sources failed"
		} else {
			s.status = "HEALTHY"
			s.detail = "Official RSS"
		}
	} else if fail > 0 {
		s.status = "DEGRADED"
		if s.lastOK.IsZero() {
			s.status = "OFFLINE"
		}
		s.detail = "Source fetch failed"
	}
	return changed
}

func (s *Service) Clusters(now time.Time) []Cluster {
	return s.store.Clusters(now)
}
