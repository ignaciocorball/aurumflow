package binanceusdm

import (
	"context"
	"time"

	"aurumflow/internal/md"
)

// DepthProbe is a bounded live check. A WebSocket connect alone is not success.
type DepthProbe struct {
	Endpoint  string `json:"endpoint"`
	Snapshot  bool   `json:"snapshot"`
	Deltas    int    `json:"deltas"`
	Synced    bool   `json:"synced"`
	Gaps      int    `json:"gaps"`
	Resyncs   int    `json:"resyncs"`
	Status    string `json:"status"`
	LastError string `json:"last_error,omitempty"`
}

// ProbeIncrementalDepth connects to the official public depth stream, obtains
// a REST snapshot, and counts incremental events. Success requires deltas > 0
// and a synced book.
func ProbeIncrementalDepth(ctx context.Context, wait time.Duration) DepthProbe {
	if wait <= 0 {
		wait = 12 * time.Second
	}
	p := DepthProbe{Endpoint: WSPublicSingle + "/btcusdt@depth@100ms", Status: "NO_DELTAS"}
	runCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	a := New()
	bus := md.NewBus(1024)
	go func() { _ = a.Run(runCtx, bus) }()
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		select {
		case <-runCtx.Done():
			goto done
		case ev := <-bus.C():
			switch ev.Kind {
			case md.KindBookSnapshot:
				p.Snapshot = true
			case md.KindBookDelta:
				p.Deltas++
			}
		case <-time.After(200 * time.Millisecond):
		}
		if a.Book != nil {
			p.Synced = a.Book.Synced
			p.Gaps = a.Book.Gaps
			p.Resyncs = a.Book.Resyncs
		}
		if p.Snapshot && p.Deltas > 0 && p.Synced {
			p.Status = "OPERATIONAL"
			cancel()
			goto done
		}
	}
done:
	p.LastError = a.LastError
	if p.Status != "OPERATIONAL" {
		if !p.Snapshot {
			p.Status = "NO_SNAPSHOT"
		} else if p.Deltas == 0 {
			p.Status = "NO_DELTAS"
		} else if !p.Synced {
			p.Status = "UNSYNCED"
		}
	}
	return p
}
