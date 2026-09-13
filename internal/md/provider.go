package md

import "context"

// Capability bits. Missing capabilities are not stubbed as empty streams.
type Caps uint32

const (
	CapQuotes Caps = 1 << iota
	CapTrades
	CapCandles
	CapBookSnapshot
	CapBookDelta
	CapHealth
)

type MarketDataProvider interface {
	Name() string
	Capabilities() Caps
	Run(ctx context.Context, out *Bus) error
}

func (c Caps) Has(x Caps) bool { return c&x != 0 }
