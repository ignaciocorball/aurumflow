package risk

import (
	"fmt"
	"math"
	"sync"
	"time"

	"aurumflow/pkg/models"
)

// Manager enforces risk per trade, max trades, and daily drawdown limit.
type Manager struct {
	RiskPerTrade       float64
	MaxTrades          int
	DailyDrawdownLimit float64
	Balance            float64
	DailyStartBalance  float64
	OpenCount          int
	mu                 sync.Mutex
	dailyReset         time.Time
}

// NewManager creates a risk manager with config.
func NewManager(riskPerTrade float64, maxTrades int, dailyDrawdownLimit float64, balance float64) *Manager {
	return &Manager{
		RiskPerTrade:       riskPerTrade,
		MaxTrades:          maxTrades,
		DailyDrawdownLimit: dailyDrawdownLimit,
		Balance:            balance,
		DailyStartBalance:  balance,
		dailyReset:         time.Now().UTC().Truncate(24 * time.Hour),
	}
}

// UpdateBalance sets current balance (e.g. from API).
func (m *Manager) UpdateBalance(balance float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Balance = balance
}

// SetOpenCount sets the number of open positions (from API).
func (m *Manager) SetOpenCount(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OpenCount = n
}

// GetBalance returns current balance (thread-safe).
func (m *Manager) GetBalance() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Balance
}

// GetOpenCount returns current open positions count (thread-safe).
func (m *Manager) GetOpenCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.OpenCount
}

// ResetDailyIfNewDay resets daily start balance at start of new day.
func (m *Manager) ResetDailyIfNewDay(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	today := now.UTC().Truncate(24 * time.Hour)
	if today.After(m.dailyReset) {
		m.dailyReset = today
		m.DailyStartBalance = m.Balance
	}
}

// DailyDrawdownPct returns current daily drawdown as percentage.
func (m *Manager) DailyDrawdownPct() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.DailyStartBalance <= 0 {
		return 0
	}
	dd := (m.DailyStartBalance - m.Balance) / m.DailyStartBalance * 100
	if dd < 0 {
		return 0
	}
	return dd
}

// CanOpenTrade returns true if we can open a new trade (under max trades and daily DD limit).
func (m *Manager) CanOpenTrade() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.OpenCount >= m.MaxTrades {
		return false
	}
	if m.DailyStartBalance <= 0 {
		return true
	}
	dd := (m.DailyStartBalance - m.Balance) / m.DailyStartBalance * 100
	return dd < m.DailyDrawdownLimit
}

// PositionSize computes size from balance, risk percent, stop distance, and value per point.
// stopDistance is in price units (e.g. 10.0). valuePerPoint = money per 1.0 price move per 1.0 size; if <= 0 uses 1.0.
// riskAmount = balance * (riskPercent/100); size = riskAmount / (stopDistance * valuePerPoint).
func PositionSize(balance, riskPercent, stopDistance, minSize, sizeStep, valuePerPoint float64) (float64, error) {
	if balance <= 0 || riskPercent <= 0 || stopDistance <= 0 {
		return 0, fmt.Errorf("invalid inputs: balance=%f risk%%=%f stopDist=%f", balance, riskPercent, stopDistance)
	}
	if valuePerPoint <= 0 {
		valuePerPoint = 1.0
	}
	riskAmount := balance * (riskPercent / 100)
	size := riskAmount / (math.Abs(stopDistance) * valuePerPoint)
	if minSize > 0 && size < minSize {
		size = minSize
	}
	if sizeStep > 0 {
		size = math.Floor(size/sizeStep) * sizeStep
		if size < minSize {
			size = minSize
		}
	}
	return size, nil
}

// ValidateSignal checks if we can act on the signal (risk limits).
func (m *Manager) ValidateSignal(signal *models.TradeSignal) (bool, error) {
	if signal == nil {
		return false, fmt.Errorf("nil signal")
	}
	if !m.CanOpenTrade() {
		return false, fmt.Errorf("risk: max trades or daily DD limit reached")
	}
	return true, nil
}

// StopDistance returns absolute distance from entry to stop for the signal.
func StopDistance(signal *models.TradeSignal) float64 {
	if signal == nil {
		return 0
	}
	dist := signal.Entry - signal.StopLoss
	if dist < 0 {
		dist = -dist
	}
	return dist
}
