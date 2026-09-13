package stratrade

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type WeekState struct {
	Epic                     string    `json:"epic"`
	MaxTrades                int       `json:"max_trades"`
	CompletedStrategyTrades  int       `json:"completed_strategy_trades"`
	OpenStrategyTradeID      string    `json:"open_strategy_trade_id"`
	LastStrategyTradeID      string    `json:"last_strategy_trade_id"`
	LastSignalID             string    `json:"last_signal_id"`
	UpdatedAt                time.Time `json:"updated_at"`
	mu                       sync.Mutex
	path                     string
}

func WeekPath(dir, epic string) string {
	if dir == "" {
		dir = "journals"
	}
	return filepath.Join(dir, "strategy-week-"+epic+".json")
}

func LoadWeek(path, epic string, maxTrades int) (*WeekState, error) {
	if maxTrades <= 0 {
		maxTrades = 1
	}
	st := &WeekState{Epic: epic, MaxTrades: maxTrades, path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return st, err
	}
	var loaded WeekState
	if err := json.Unmarshal(b, &loaded); err != nil {
		return st, err
	}
	st.Epic = loaded.Epic
	if st.Epic == "" {
		st.Epic = epic
	}
	st.MaxTrades = loaded.MaxTrades
	if st.MaxTrades <= 0 {
		st.MaxTrades = maxTrades
	}
	st.CompletedStrategyTrades = loaded.CompletedStrategyTrades
	st.OpenStrategyTradeID = loaded.OpenStrategyTradeID
	st.LastStrategyTradeID = loaded.LastStrategyTradeID
	st.LastSignalID = loaded.LastSignalID
	st.UpdatedAt = loaded.UpdatedAt
	return st, nil
}

func (s *WeekState) BlocksNewStrategy() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.MaxTrades > 0 && s.CompletedStrategyTrades >= s.MaxTrades
}

func (s *WeekState) HasOpen() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.OpenStrategyTradeID != ""
}

func (s *WeekState) Snapshot() WeekState {
	if s == nil {
		return WeekState{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return WeekState{
		Epic:                    s.Epic,
		MaxTrades:               s.MaxTrades,
		CompletedStrategyTrades: s.CompletedStrategyTrades,
		OpenStrategyTradeID:     s.OpenStrategyTradeID,
		LastStrategyTradeID:     s.LastStrategyTradeID,
		LastSignalID:            s.LastSignalID,
		UpdatedAt:               s.UpdatedAt,
	}
}

func (s *WeekState) MarkOpen(id, signalID string) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OpenStrategyTradeID = id
	s.LastStrategyTradeID = id
	s.LastSignalID = signalID
	s.UpdatedAt = time.Now().UTC()
	return s.persistLocked()
}

func (s *WeekState) MarkClosed(id string) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.OpenStrategyTradeID != "" {
		s.CompletedStrategyTrades++
	}
	s.OpenStrategyTradeID = ""
	s.LastStrategyTradeID = id
	s.UpdatedAt = time.Now().UTC()
	return s.persistLocked()
}

func (s *WeekState) persistLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	payload := struct {
		Epic                    string    `json:"epic"`
		MaxTrades               int       `json:"max_trades"`
		CompletedStrategyTrades int       `json:"completed_strategy_trades"`
		OpenStrategyTradeID     string    `json:"open_strategy_trade_id"`
		LastStrategyTradeID     string    `json:"last_strategy_trade_id"`
		LastSignalID            string    `json:"last_signal_id"`
		UpdatedAt               time.Time `json:"updated_at"`
	}{
		Epic:                    s.Epic,
		MaxTrades:               s.MaxTrades,
		CompletedStrategyTrades: s.CompletedStrategyTrades,
		OpenStrategyTradeID:     s.OpenStrategyTradeID,
		LastStrategyTradeID:     s.LastStrategyTradeID,
		LastSignalID:            s.LastSignalID,
		UpdatedAt:               s.UpdatedAt,
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0644)
}

func RecoveryFlatOK(local, broker, pending int) bool {
	return local == 0 && broker == 0 && pending == 0
}

func MaxTradesSafeAfterRestart(beforeCompleted, afterCompleted, maxTrades int, wouldAllowSecond bool) bool {
	if beforeCompleted != afterCompleted {
		return false
	}
	if afterCompleted >= maxTrades && wouldAllowSecond {
		return false
	}
	return true
}
