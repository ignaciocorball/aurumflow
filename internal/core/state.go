package core

import (
	"sync"
	"time"
)

// BotState represents the current state of the bot.
type BotState string

const (
	StateIDLE          BotState = "IDLE"
	StateSeekLiquidity BotState = "SEEK_LIQUIDITY"
	StateWaitSweep     BotState = "WAIT_SWEEP"
	StateWaitPullback  BotState = "WAIT_PULLBACK"
	StateReady         BotState = "READY"
	StateInTrade       BotState = "IN_TRADE"
	StateCooldown      BotState = "COOLDOWN"
)

// StateMachine gates when the execution layer is allowed to send orders.
// TTL is counted in new M15 candles (lastCandleTime change). setupDirection is set when entering WAIT_SWEEP.
type StateMachine struct {
	state            BotState
	lastCandleTime   time.Time
	stateCandleCount int
	setupDirection   string // "BUY", "SELL", or ""
	enteredAt        time.Time
	mu               sync.RWMutex
}

// NewStateMachine returns a state machine starting at IDLE.
func NewStateMachine() *StateMachine {
	return &StateMachine{state: StateIDLE}
}

// State returns the current state.
func (sm *StateMachine) State() BotState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

// Set sets the state (internal use or after transitions).
func (sm *StateMachine) Set(s BotState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state = s
}

// CanSendOrder returns true only when state is READY and risk allows (caller checks risk).
func (sm *StateMachine) CanSendOrder() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state == StateReady
}

// TransitionParams holds parameters for Transition (TTL and invalidation).
type TransitionParams struct {
	LastCandleTime             time.Time
	StructureTrend             string // "BULLISH", "BEARISH", "RANGE"
	SetupDirection             string // "BUY" or "SELL" when entering WAIT_SWEEP
	StateTTLCandles            int
	InvalidateOnOppositeStructure bool
}

// resetToIdle clears TTL/setup state and sets state to IDLE.
func (sm *StateMachine) resetToIdle() {
	sm.state = StateIDLE
	sm.stateCandleCount = 0
	sm.setupDirection = ""
}

// Transition updates state based on strategy outputs.
// Liquidity -> SEEK_LIQUIDITY; Sweep -> WAIT_SWEEP; then WAIT_PULLBACK; at equilibrium -> READY.
// TTL: if stateCandleCount >= StateTTLCandles in SEEK_LIQUIDITY/WAIT_SWEEP/WAIT_PULLBACK/READY, reset to IDLE.
// Invalidation: if structure trend opposes setupDirection (e.g. setup BUY but trend BEARISH), reset to IDLE.
func (sm *StateMachine) Transition(hasLiquidity, hasSweep, atEquilibrium bool, inTrade bool, params TransitionParams) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if inTrade {
		sm.state = StateInTrade
		return
	}

	// Count new M15 candle: only increment when last candle time actually changed
	if !params.LastCandleTime.IsZero() && params.LastCandleTime != sm.lastCandleTime {
		sm.stateCandleCount++
		sm.lastCandleTime = params.LastCandleTime
	}

	ttlCandles := params.StateTTLCandles
	if ttlCandles <= 0 {
		ttlCandles = 12
	}
	invalidate := params.InvalidateOnOppositeStructure

	// TTL: reset if we've been in transient states too long
	switch sm.state {
	case StateSeekLiquidity, StateWaitSweep, StateWaitPullback, StateReady:
		if sm.stateCandleCount >= ttlCandles {
			sm.resetToIdle()
			return
		}
		// Invalidation: opposite structure
		if invalidate && sm.setupDirection != "" && params.StructureTrend != "" {
			if (sm.setupDirection == "BUY" && (params.StructureTrend == "BEARISH" || params.StructureTrend == "RANGE")) ||
				(sm.setupDirection == "SELL" && (params.StructureTrend == "BULLISH" || params.StructureTrend == "RANGE")) {
				sm.resetToIdle()
				return
			}
		}
	}

	switch sm.state {
	case StateIDLE:
		if hasLiquidity {
			sm.state = StateSeekLiquidity
			sm.stateCandleCount = 0
			sm.lastCandleTime = params.LastCandleTime
			sm.enteredAt = params.LastCandleTime
		}
	case StateSeekLiquidity:
		if hasSweep {
			sm.state = StateWaitSweep
			sm.setupDirection = params.SetupDirection
		}
	case StateWaitSweep:
		sm.state = StateWaitPullback
	case StateWaitPullback:
		if atEquilibrium {
			sm.state = StateReady
		}
	case StateReady:
		// Stay READY until we send or timeout; caller will set IN_TRADE or COOLDOWN
	case StateInTrade:
		// Caller sets COOLDOWN when trade closes
	case StateCooldown:
		sm.resetToIdle()
	}
}

// ToCooldown sets state to COOLDOWN (after trade or reject).
func (sm *StateMachine) ToCooldown() {
	sm.Set(StateCooldown)
}

// ToInTrade sets state to IN_TRADE after sending order.
func (sm *StateMachine) ToInTrade() {
	sm.Set(StateInTrade)
}

// ToReady sets state to READY (e.g. when conditions align).
func (sm *StateMachine) ToReady() {
	sm.Set(StateReady)
}

// ToWaitPullback recycles to WAIT_PULLBACK (e.g. M5 timing reject soft); does not reset setup.
func (sm *StateMachine) ToWaitPullback() {
	sm.Set(StateWaitPullback)
}
