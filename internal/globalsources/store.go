package globalsources

import (
	"sort"
	"sync"
	"time"

	"aurumflow/internal/worlddomain"
)

type Store struct {
	mu   sync.RWMutex
	obs  []worlddomain.ContextObservation
}

func (s *Store) Add(xs ...worlddomain.ContextObservation) {
	s.mu.Lock()
	s.obs = append(s.obs, xs...)
	s.mu.Unlock()
}

func (s *Store) All() []worlddomain.ContextObservation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]worlddomain.ContextObservation, len(s.obs))
	copy(out, s.obs)
	return out
}

func (s *Store) Usable(at time.Time) []worlddomain.ContextObservation {
	all := s.All()
	var out []worlddomain.ContextObservation
	for _, o := range all {
		if o.UsableAt(at) {
			out = append(out, o)
		}
	}
	return out
}

func Latest(obs []worlddomain.ContextObservation, source, metric string, at time.Time) (worlddomain.ContextObservation, bool) {
	var best worlddomain.ContextObservation
	ok := false
	for _, o := range obs {
		if o.Source != source || o.Metric != metric || !o.UsableAt(at) {
			continue
		}
		if !ok || o.AvailableAt.After(best.AvailableAt) || (o.AvailableAt.Equal(best.AvailableAt) && o.ObservedAt.After(best.ObservedAt)) {
			best, ok = o, true
		}
	}
	return best, ok
}

func Series(obs []worlddomain.ContextObservation, source, metric string, at time.Time) []worlddomain.ContextObservation {
	var out []worlddomain.ContextObservation
	for _, o := range obs {
		if o.Source == source && o.Metric == metric && o.UsableAt(at) {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ObservedAt.Before(out[j].ObservedAt) })
	return out
}
