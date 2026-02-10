package core

import (
	"testing"
	"time"
)

func TestStateMachine_TTL_ResetsToIdle(t *testing.T) {
	sm := NewStateMachine()
	baseTime := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)

	// Enter SEEK_LIQUIDITY
	sm.Transition(true, false, false, false, TransitionParams{
		LastCandleTime: baseTime,
		StructureTrend: "BULLISH",
		SetupDirection: "",
		StateTTLCandles: 3,
		InvalidateOnOppositeStructure: true,
	})
	if sm.State() != StateSeekLiquidity {
		t.Fatalf("expected SEEK_LIQUIDITY, got %s", sm.State())
	}

	// WAIT_SWEEP
	sm.Transition(true, true, false, false, TransitionParams{
		LastCandleTime: baseTime,
		StructureTrend: "BULLISH",
		SetupDirection: "BUY",
		StateTTLCandles: 3,
		InvalidateOnOppositeStructure: true,
	})
	if sm.State() != StateWaitSweep {
		t.Fatalf("expected WAIT_SWEEP, got %s", sm.State())
	}

	// WAIT_PULLBACK
	sm.Transition(true, true, false, false, TransitionParams{
		LastCandleTime: baseTime,
		StructureTrend: "BULLISH",
		SetupDirection: "BUY",
		StateTTLCandles: 3,
		InvalidateOnOppositeStructure: true,
	})
	if sm.State() != StateWaitPullback {
		t.Fatalf("expected WAIT_PULLBACK, got %s", sm.State())
	}

	// Simulate 3 new candles (change lastCandleTime each time)
	for i := 1; i <= 3; i++ {
		newCandleTime := baseTime.Add(time.Duration(i*15) * time.Minute)
		sm.Transition(true, true, false, false, TransitionParams{
			LastCandleTime: newCandleTime,
			StructureTrend: "BULLISH",
			SetupDirection: "BUY",
			StateTTLCandles: 3,
			InvalidateOnOppositeStructure: true,
		})
	}
	if sm.State() != StateIDLE {
		t.Errorf("expected IDLE after TTL (3 candles), got %s", sm.State())
	}
}

func TestStateMachine_InvalidateOnOppositeStructure(t *testing.T) {
	sm := NewStateMachine()
	baseTime := time.Date(2026, 2, 6, 12, 0, 0, 0, time.UTC)

	// Reach WAIT_PULLBACK with BUY setup
	sm.Transition(true, false, false, false, TransitionParams{
		LastCandleTime: baseTime,
		StructureTrend: "BULLISH",
		SetupDirection: "",
		StateTTLCandles: 12,
		InvalidateOnOppositeStructure: true,
	})
	sm.Transition(true, true, false, false, TransitionParams{
		LastCandleTime: baseTime,
		StructureTrend: "BULLISH",
		SetupDirection: "BUY",
		StateTTLCandles: 12,
		InvalidateOnOppositeStructure: true,
	})
	sm.Transition(true, true, false, false, TransitionParams{
		LastCandleTime: baseTime,
		StructureTrend: "BULLISH",
		SetupDirection: "BUY",
		StateTTLCandles: 12,
		InvalidateOnOppositeStructure: true,
	})
	if sm.State() != StateWaitPullback {
		t.Fatalf("expected WAIT_PULLBACK, got %s", sm.State())
	}

	// Next tick: structure turns BEARISH -> should reset to IDLE
	sm.Transition(true, true, false, false, TransitionParams{
		LastCandleTime: baseTime.Add(15 * time.Minute),
		StructureTrend: "BEARISH",
		SetupDirection: "BUY",
		StateTTLCandles: 12,
		InvalidateOnOppositeStructure: true,
	})
	if sm.State() != StateIDLE {
		t.Errorf("expected IDLE when setupDirection BUY and structure BEARISH, got %s", sm.State())
	}
}
