package book

// LiquidityHeuristics require persistence; a single large order is not a signal.
type LiquidityHeuristics struct {
	LastBidQty, LastAskQty float64
	BidPersist, AskPersist int
	LastBid, LastAsk       float64
}

type LiquidityFlags struct {
	DepletionBid     bool
	DepletionAsk     bool
	ReplenishBid     bool
	ReplenishAsk     bool
	PersistentBid    bool
	PersistentAsk    bool
	SweepBid         bool
	SweepAsk         bool
}

func (h *LiquidityHeuristics) Observe(bid, ask, bidQty, askQty, typicalQty float64) LiquidityFlags {
	if typicalQty <= 0 {
		typicalQty = 1
	}
	var f LiquidityFlags
	if h.LastBidQty > 0 && bidQty < h.LastBidQty*0.5 && h.LastBidQty >= typicalQty {
		f.DepletionBid = true
	}
	if h.LastAskQty > 0 && askQty < h.LastAskQty*0.5 && h.LastAskQty >= typicalQty {
		f.DepletionAsk = true
	}
	if h.LastBidQty > 0 && bidQty > h.LastBidQty*1.5 && bidQty >= typicalQty {
		f.ReplenishBid = true
	}
	if h.LastAskQty > 0 && askQty > h.LastAskQty*1.5 && askQty >= typicalQty {
		f.ReplenishAsk = true
	}
	if bid == h.LastBid && bidQty >= typicalQty {
		h.BidPersist++
	} else {
		h.BidPersist = 1
	}
	if ask == h.LastAsk && askQty >= typicalQty {
		h.AskPersist++
	} else {
		h.AskPersist = 1
	}
	f.PersistentBid = h.BidPersist >= 3
	f.PersistentAsk = h.AskPersist >= 3
	// Sweep: best level disappears and price steps through.
	if h.LastBid > 0 && bid < h.LastBid && f.DepletionBid {
		f.SweepBid = true
	}
	if h.LastAsk > 0 && ask > h.LastAsk && f.DepletionAsk {
		f.SweepAsk = true
	}
	h.LastBid, h.LastAsk, h.LastBidQty, h.LastAskQty = bid, ask, bidQty, askQty
	return f
}

func (b *Book) TopQty() (bidQty, askQty float64, ok bool) {
	bid, ask, ok := b.BestBidAsk()
	if !ok {
		return 0, 0, false
	}
	return b.Bids[bid], b.Asks[ask], true
}

func (b *Book) Depth(n int) (bidDepth, askDepth float64) {
	for _, l := range topN(b.Bids, n, true) {
		bidDepth += l.Qty
	}
	for _, l := range topN(b.Asks, n, false) {
		askDepth += l.Qty
	}
	return bidDepth, askDepth
}
