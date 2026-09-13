package capitalhist

import (
	"context"
	"fmt"
	"time"

	"aurumflow/internal/md"
)

// StreamingProvider is a read-only Capital market-data adapter.
// It never holds an ExecutionProvider.
type StreamingProvider struct {
	Host   string
	Status string
}

func (s *StreamingProvider) Name() string { return "capital_streaming" }

func (s *StreamingProvider) Capabilities() md.Caps {
	return md.CapQuotes | md.CapCandles | md.CapHealth
}

func (s *StreamingProvider) Run(ctx context.Context, out *md.Bus) error {
	if s.Status == "" {
		s.Status = "RUNTIME_LIMITED"
	}
	_ = out.Publish(ctx, md.Event{
		Kind: md.KindProviderHealth, EventTime: time.Now().UTC(), ReceiveTime: time.Now().UTC(),
		Provider: s.Name(), Venue: "capital.com",
		Payload: md.Health{OK: false, LastError: s.Status},
		Prov: md.Provenance{Provider: s.Name(), Venue: "capital.com", Quality: md.QualityDirect, FreshnessClass: md.FreshDirect},
	})
	<-ctx.Done()
	return ctx.Err()
}

func (s *StreamingProvider) NormalizeQuote(epic string, bid, ask float64, t time.Time) md.Event {
	return md.Event{
		Kind: md.KindQuote, EventTime: t, ReceiveTime: t,
		Provider: s.Name(), Venue: "capital.com", Instrument: epic,
		Payload: md.Quote{Bid: bid, Ask: ask},
		Prov: md.Provenance{
			Provider: s.Name(), Venue: "capital.com", Instrument: epic,
			Quality: md.QualityDirect, FreshnessClass: md.FreshDirect,
			EventTime: t, AvailableAt: t,
		},
	}
}

func StreamingNote() string {
	return fmt.Sprintf("Capital streaming implemented as read-only adapter; live WSS marked RUNTIME_LIMITED until DEMO CST session is attached")
}
