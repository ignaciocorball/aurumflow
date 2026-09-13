package killswitch

import (
	"os"
	"strings"
	"sync"
)

const (
	EnvName     = "AURUMFLOW_KILL_SWITCH"
	DefaultFile = ".aurumflow.kill"
	Reason      = "KILL_SWITCH_ACTIVE"
)

// Switch is a simple HALT_NEW_ORDERS gate (config + env + sentinel file).
type Switch struct {
	mu       sync.RWMutex
	halted   bool
	filePath string
}

func New(enabled bool, filePath string) *Switch {
	if strings.TrimSpace(filePath) == "" {
		filePath = DefaultFile
	}
	s := &Switch{halted: enabled, filePath: filePath}
	if envHalt() {
		s.halted = true
	}
	return s
}

func envHalt() bool {
	v := strings.TrimSpace(os.Getenv(EnvName))
	return v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "halt")
}

func (s *Switch) HaltNewOrders() bool {
	if s == nil {
		return envHalt()
	}
	s.mu.RLock()
	halted := s.halted
	file := s.filePath
	s.mu.RUnlock()
	if halted || envHalt() {
		return true
	}
	if file != "" {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}
	return false
}

func (s *Switch) Halt() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.halted = true
	s.mu.Unlock()
}

// HaltPersist turns HALT_NEW_ORDERS on and writes the sentinel file.
func (s *Switch) HaltPersist() error {
	if s == nil {
		return os.WriteFile(DefaultFile, []byte("HALT_NEW_ORDERS\n"), 0644)
	}
	s.Halt()
	return os.WriteFile(s.FilePath(), []byte("HALT_NEW_ORDERS\n"), 0644)
}

func (s *Switch) Resume() {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.halted = false
	s.mu.Unlock()
}

func (s *Switch) FilePath() string {
	if s == nil {
		return DefaultFile
	}
	return s.filePath
}
