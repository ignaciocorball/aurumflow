package notifications

import (
	"sync"
	"time"
)

// Store holds in-memory state for dedup, throttle, and aggregation. Thread-safe.
type Store struct {
	mu              sync.Mutex
	lastSendByKey   map[string]time.Time
	lastSendByType  map[string]time.Time
	unknownReasons  map[string]time.Time // REJ_UNKNOWN:* -> last sent; cap 1 per dedup window
	aggregateRejects []aggregateReject   // SIGNAL_REJECTED bucket by reason
	lastSweepType   string              // for SWEEP_CONFIRMED type change detection
	dedupWindow     time.Duration
	aggregateWindow time.Duration
}

type aggregateReject struct {
	ReasonCode string
	Count      int
	LastDir    string
	LastScore  int
	LastConf   float64
	At         time.Time
}

// NewStore creates a store with the given windows.
func NewStore(dedupWindow, aggregateWindow time.Duration) *Store {
	return &Store{
		lastSendByKey:   make(map[string]time.Time),
		lastSendByType:  make(map[string]time.Time),
		unknownReasons:  make(map[string]time.Time),
		aggregateRejects: nil,
		dedupWindow:     dedupWindow,
		aggregateWindow: aggregateWindow,
	}
}

// ShouldDedup returns true if we already sent this DedupKey within the dedup window.
func (s *Store) ShouldDedup(dedupKey string, ev NotifEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	// Prune old keys
	for k, t := range s.lastSendByKey {
		if now.Sub(t) > s.dedupWindow {
			delete(s.lastSendByKey, k)
		}
	}
	if t, ok := s.lastSendByKey[dedupKey]; ok && now.Sub(t) <= s.dedupWindow {
		return true
	}
	// REJ_UNKNOWN:*: cap to 1 per dedup window for unknown codes
	if ev.Type == TypeSignalRejected {
		if reasonCode, ok := ev.Payload["reason_code"].(string); ok && len(reasonCode) > len(RejUnknown) && reasonCode[:len(RejUnknown)] == RejUnknown {
			if t, ok := s.unknownReasons[reasonCode]; ok && now.Sub(t) <= s.dedupWindow {
				return true
			}
		}
	}
	return false
}

// RecordSend records that we sent this dedup key and event type (for throttle).
func (s *Store) RecordSend(dedupKey, evType string, ev NotifEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	s.lastSendByKey[dedupKey] = now
	s.lastSendByType[evType] = now
	if ev.Type == TypeSignalRejected {
		if reasonCode, ok := ev.Payload["reason_code"].(string); ok && len(reasonCode) > len(RejUnknown) && reasonCode[:len(RejUnknown)] == RejUnknown {
			s.unknownReasons[reasonCode] = now
		}
	}
}

// Throttled returns true if we sent this event type recently within minInterval.
func (s *Store) Throttled(evType string, minInterval time.Duration) bool {
	if minInterval <= 0 {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if t, ok := s.lastSendByType[evType]; ok && now.Sub(t) < minInterval {
		return true
	}
	return false
}

// AddReject adds a SIGNAL_REJECTED to the aggregate bucket.
func (s *Store) AddReject(ev NotifEvent) {
	reasonCode, _ := ev.Payload["reason_code"].(string)
	if reasonCode == "" {
		reasonCode = RejNoSignal
	}
	dir, _ := ev.Payload["direction"].(string)
	score, _ := ev.Payload["score"].(int)
	conf, _ := ev.Payload["confidence"].(float64)
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	found := false
	for i := range s.aggregateRejects {
		if s.aggregateRejects[i].ReasonCode == reasonCode {
			s.aggregateRejects[i].Count++
			s.aggregateRejects[i].LastDir = dir
			s.aggregateRejects[i].LastScore = score
			s.aggregateRejects[i].LastConf = conf
			s.aggregateRejects[i].At = now
			found = true
			break
		}
	}
	if !found {
		s.aggregateRejects = append(s.aggregateRejects, aggregateReject{
			ReasonCode: reasonCode,
			Count:      1,
			LastDir:    dir,
			LastScore:  score,
			LastConf:   conf,
			At:         now,
		})
	}
}

// FlushRejects returns and clears the current aggregate bucket (for sending one summary).
func (s *Store) FlushRejects() (total int, byReason map[string]int, lastDir string, lastScore int, lastConf float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	byReason = make(map[string]int)
	for _, a := range s.aggregateRejects {
		total += a.Count
		byReason[a.ReasonCode] = a.Count
		lastDir = a.LastDir
		lastScore = a.LastScore
		lastConf = a.LastConf
	}
	s.aggregateRejects = nil
	return total, byReason, lastDir, lastScore, lastConf
}

// RejectCount returns current aggregate count without flushing.
func (s *Store) RejectCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, a := range s.aggregateRejects {
		n += a.Count
	}
	return n
}

// SweepTypeChanged updates last sweep type and returns true if type changed BUY_SIDE <-> SELL_SIDE.
func (s *Store) SweepTypeChanged(sweepType string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := false
	if s.lastSweepType != "" && sweepType != "" && s.lastSweepType != sweepType {
		if (s.lastSweepType == "BUY_SIDE" && sweepType == "SELL_SIDE") || (s.lastSweepType == "SELL_SIDE" && sweepType == "BUY_SIDE") {
			changed = true
		}
	}
	s.lastSweepType = sweepType
	return changed
}
