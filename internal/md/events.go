package md

import "time"

type Kind string

const (
	KindTrade         Kind = "trade"
	KindQuote         Kind = "quote"
	KindBookSnapshot  Kind = "book_snapshot"
	KindBookDelta     Kind = "book_delta"
	KindCandle        Kind = "candle"
	KindProviderHealth Kind = "provider_health"
)

type Event struct {
	Kind        Kind
	EventTime   time.Time
	ReceiveTime time.Time
	Provider    string
	Venue       string
	Instrument  string
	Seq         int64
	Payload     any
}

type Trade struct {
	Price, Qty float64
	BuyerMaker bool
}

type Quote struct {
	Bid, Ask, BidQty, AskQty float64
}

type Level struct {
	Price, Qty float64
}

type BookDelta struct {
	Bids, Asks []Level
	FirstID    int64
	FinalID    int64
	PrevFinal  int64
}

func OrderEvents(in []Event) []Event {
	out := append([]Event(nil), in...)
	for i := 1; i < len(out); i++ {
		j := i
		for j > 0 && eventLess(out[j], out[j-1]) {
			out[j], out[j-1] = out[j-1], out[j]
			j--
		}
	}
	return out
}

func eventLess(a, b Event) bool {
	if a.Seq != 0 && b.Seq != 0 && a.Seq != b.Seq {
		return a.Seq < b.Seq
	}
	if !a.EventTime.Equal(b.EventTime) {
		return a.EventTime.Before(b.EventTime)
	}
	return a.ReceiveTime.Before(b.ReceiveTime)
}

type Health struct {
	OK          bool
	BookSynced  bool
	Reconnects  int
	Resyncs     int
	LastError   string
	EventsIn    int64
	EventsDrop  int64
}
