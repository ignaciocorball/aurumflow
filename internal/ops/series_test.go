package ops

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestSeriesBounded(t *testing.T) {
	s := &Server{}
	now := time.Unix(1_700_000_000, 0).UTC()
	for i := 0; i < 4000; i++ {
		s.lastSample = time.Time{}
		s.appendSeriesLocked(Status{BTCPrice: float64(i)}, now.Add(time.Duration(i)*time.Second))
	}
	if len(s.series) > MaxSeriesPoints {
		t.Fatalf("unbounded %d", len(s.series))
	}
	if len(s.timeline) > MaxTimeline {
		t.Fatalf("timeline %d", len(s.timeline))
	}
}

func TestTimeseriesAPIBoundAndHorizon(t *testing.T) {
	s := NewServer("127.0.0.1:0", NewStatus())
	now := time.Now().UTC()
	s.mu.Lock()
	for i := 0; i < 10; i++ {
		s.lastSample = time.Time{}
		s.appendSeriesLocked(Status{BTCPrice: float64(i + 1)}, now.Add(time.Duration(i)*time.Second))
	}
	s.mu.Unlock()
	addr := waitServer(t, s)
	resp, err := http.Get("http://" + addr + "/api/timeseries")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		Points []SeriesPoint `json:"points"`
		Bound  int           `json:"bound"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Bound != MaxSeriesPoints {
		t.Fatalf("bound %d", out.Bound)
	}
	if len(out.Points) == 0 || len(out.Points) > MaxSeriesPoints {
		t.Fatalf("points %d", len(out.Points))
	}
}

func TestSignalTimeline(t *testing.T) {
	s := &Server{}
	t0 := time.Unix(1_700_000_000, 0).UTC()
	s.appendSeriesLocked(Status{LastV1Class: "FLOW_NEUTRAL"}, t0)
	s.lastSample = time.Time{}
	s.appendSeriesLocked(Status{LastV1Class: "FLOW_EXHAUSTION_CONFIRM", LastLegacyDir: 1, BookSynced: true}, t0.Add(time.Second))
	s.lastSample = time.Time{}
	s.appendSeriesLocked(Status{LastV1Class: "FLOW_EXHAUSTION_CONFIRM", LastLegacyDir: 1, BookSynced: false, BookGaps: 1}, t0.Add(2*time.Second))
	s.lastSample = time.Time{}
	s.appendSeriesLocked(Status{LastV1Class: "FLOW_EXHAUSTION_CONFIRM", MarketStatus: "TRADEABLE", GoldBid: 1, GoldAsk: 2}, t0.Add(3*time.Second))
	kinds := map[string]bool{}
	for _, e := range s.timeline {
		kinds[e.Kind] = true
		if e.Text == "GOLD MARKET TRADEABLE" {
			kinds["gold_tradeable"] = true
		}
	}
	for _, k := range []string{"v1", "legacy", "book", "gold_tradeable"} {
		if !kinds[k] {
			t.Fatalf("missing %s %+v", k, s.timeline)
		}
	}
}

func TestAbsorptionTimelineDebouncesFlicker(t *testing.T) {
	s := &Server{}
	t0 := time.Unix(1_700_000_000, 0).UTC()
	for i, st := range []string{"MIXED", "NOT_SUPPORTIVE", "MIXED", "NOT_SUPPORTIVE", "MIXED"} {
		s.lastSample = time.Time{}
		s.appendSeriesLocked(Status{AbsorptionStatus: st}, t0.Add(time.Duration(i)*time.Second))
	}
	for _, e := range s.timeline {
		if e.Kind == "absorption" {
			t.Fatalf("flicker leaked %+v", s.timeline)
		}
	}
	s.lastSample = time.Time{}
	s.appendSeriesLocked(Status{AbsorptionStatus: "NOT_SUPPORTIVE"}, t0.Add(5*time.Second))
	s.lastSample = time.Time{}
	s.appendSeriesLocked(Status{AbsorptionStatus: "NOT_SUPPORTIVE"}, t0.Add(10*time.Second))
	got := 0
	for _, e := range s.timeline {
		if e.Kind == "absorption" && e.Text == "NOT_SUPPORTIVE" {
			got++
		}
	}
	if got != 1 {
		t.Fatalf("dwell emit %d %+v", got, s.timeline)
	}
}

func TestAbsorptionTimelineMaterialImmediate(t *testing.T) {
	s := &Server{}
	t0 := time.Unix(1_700_000_000, 0).UTC()
	s.appendSeriesLocked(Status{AbsorptionStatus: "MIXED"}, t0)
	s.lastSample = time.Time{}
	s.appendSeriesLocked(Status{AbsorptionStatus: "SUPPORTIVE"}, t0.Add(time.Second))
	s.lastSample = time.Time{}
	s.appendSeriesLocked(Status{AbsorptionStatus: "UNAVAILABLE"}, t0.Add(2*time.Second))
	kinds := []string{}
	for _, e := range s.timeline {
		if e.Kind == "absorption" {
			kinds = append(kinds, e.Text)
		}
	}
	if len(kinds) != 2 || kinds[0] != "SUPPORTIVE" || kinds[1] != "UNAVAILABLE" {
		t.Fatalf("%v", kinds)
	}
}

func TestAbsorptionRawStatusUnaffected(t *testing.T) {
	s := NewServer("127.0.0.1:0", NewStatus())
	st := s.Get()
	st.AbsorptionStatus = "MIXED"
	s.Set(st)
	st = s.Get()
	st.AbsorptionStatus = "NOT_SUPPORTIVE"
	s.Set(st)
	if s.Get().AbsorptionStatus != "NOT_SUPPORTIVE" {
		t.Fatal(s.Get().AbsorptionStatus)
	}
	for _, e := range s.Timeline() {
		if e.Kind == "absorption" {
			t.Fatalf("raw flip published %+v", s.Timeline())
		}
	}
}
