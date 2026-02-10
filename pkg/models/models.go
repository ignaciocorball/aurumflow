package models

import "time"

// Candle represents a single OHLC candle.
type Candle struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Swing represents a pivot high or low.
type Swing struct {
	Index int
	Price float64
	Type  string // "HIGH" or "LOW"
}

// MarketStructure holds trend and structure flags.
type MarketStructure struct {
	Trend string // "BULLISH", "BEARISH", "RANGE"
	HH    bool   // Higher High
	HL    bool   // Higher Low
	LH    bool   // Lower High
	LL    bool   // Lower Low
}

// StructureState holds phase and direction.
type StructureState struct {
	Phase     string // "IMPULSE", "PULLBACK", "EXPANSION"
	Direction string // "BULLISH", "BEARISH"
}

// Fibonacci holds retracement levels (auto-anchored to swings).
// Low/High are always the min/max price; ImpulseDirection is "UP" (last swing was high) or "DOWN" (last swing was low).
type Fibonacci struct {
	Level382         float64
	Level50          float64
	Level618         float64
	Level705         float64
	High             float64
	Low              float64
	ImpulseDirection string // "UP", "DOWN" — for direction-aware SL/TP
}

// LiquidityZone represents a zone where stops may cluster.
type LiquidityZone struct {
	Price    float64
	Type     string  // "BUY_LIQUIDITY", "SELL_LIQUIDITY"
	Strength float64
}

// LiquidityEvent represents a liquidity sweep or similar event.
type LiquidityEvent struct {
	Type     string  // "BUY_SIDE", "SELL_SIDE"
	Strength float64
}

// TradeSignal is the output of the strategy composer.
type TradeSignal struct {
	Direction   string  // "BUY", "SELL"
	Entry       float64
	StopLoss    float64
	TakeProfit  float64
	Confidence  float64
	Score       int
}
