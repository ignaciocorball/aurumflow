package strathist

import (
	"context"
	"testing"
	"time"

	"aurumflow/config"
	"aurumflow/pkg/models"
)

func candle(t time.Time, px float64) models.Candle {
	return models.Candle{Time: t, Open: px, High: px + 1, Low: px - 1, Close: px}
}

func TestThirtyM15MakesReady(t *testing.T) {
	asOf := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	var m15 []models.Candle
	for i := 0; i < 30; i++ {
		m15 = append(m15, candle(asOf.Add(-time.Duration(30-i)*15*time.Minute), 4300))
	}
	req := RequirementsFromConfig(&config.Config{})
	closed := ClosedOnly(m15, asOf, req.M15Period)
	if len(closed) < 15 {
		t.Fatalf("got %d", len(closed))
	}
	if Status(0, len(closed), 0, 0, req, false) != HistoryReady {
		t.Fatal(Status(0, len(closed), 0, 0, req, false))
	}
}

func TestSeedThenLiveNextCanonicalM15(t *testing.T) {
	asOf := time.Date(2026, 9, 13, 16, 0, 0, 0, time.UTC)
	req := RequirementsFromConfig(&config.Config{})
	var seed []models.Candle
	for i := 0; i < 20; i++ {
		seed = append(seed, candle(asOf.Add(-time.Duration(20-i)*15*time.Minute), 4300+float64(i)))
	}
	b := Book{M15: seed}
	b.Refresh(req)
	next := candle(asOf.Add(15*time.Minute), 4320)
	later := asOf.Add(16 * time.Minute)
	b.MergeLive(nil, []models.Candle{next}, nil, nil, req, later)
	if b.M15Count < 21 {
		t.Fatalf("live close not merged: %d", b.M15Count)
	}
	if !b.M15Latest.Equal(next.Time) {
		t.Fatalf("latest=%s want=%s", b.M15Latest, next.Time)
	}
}

func TestIncompleteBarNotDuplicated(t *testing.T) {
	asOf := time.Date(2026, 9, 11, 16, 7, 0, 0, time.UTC)
	start := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	seed := []models.Candle{candle(start, 4300)}
	live := []models.Candle{candle(start, 4301)}
	merged := mergeTF(seed, live, asOf, 15*time.Minute)
	if len(merged) != 1 {
		t.Fatalf("dup %d", len(merged))
	}
	if merged[0].Close != 4301 {
		t.Fatal("live must win same timestamp")
	}
}

func TestHistoricalLiveDedup(t *testing.T) {
	asOf := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	ts := asOf.Add(-30 * time.Minute)
	got := Merge([]models.Candle{candle(ts, 1)}, []models.Candle{candle(ts, 2)})
	if len(got) != 1 || got[0].Close != 2 {
		t.Fatalf("%+v", got)
	}
}

func TestWeekendGapNotSynthetic(t *testing.T) {
	fri := time.Date(2026, 9, 11, 20, 0, 0, 0, time.UTC)
	sun := time.Date(2026, 9, 13, 22, 0, 0, 0, time.UTC)
	if !ExpectedWeekendGap(fri, sun) {
		t.Fatal("friday to sunday")
	}
	cs := []models.Candle{candle(fri, 4300), candle(sun, 4310)}
	we, un := ClassifyGaps(cs, 15*time.Minute)
	if we != 1 || un != 0 {
		t.Fatalf("we=%d un=%d", we, un)
	}
	if len(cs) != 2 {
		t.Fatal("must not insert weekend bars")
	}
}

func TestHistoryFailedUnavailable(t *testing.T) {
	req := RequirementsFromConfig(nil)
	if Status(0, 0, 0, 0, req, true) != HistoryFailed {
		t.Fatal("unavailable")
	}
	if Status(0, 3, 0, 0, req, false) != HistoryPartial {
		t.Fatal("partial")
	}
}

func TestNoLookahead(t *testing.T) {
	asOf := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	future := candle(asOf.Add(15*time.Minute), 4400)
	if !RejectLookahead(future, asOf, 15*time.Minute) {
		t.Fatal("future bar")
	}
	forming := candle(asOf.Add(-5*time.Minute), 4300)
	if !Incomplete(forming, asOf, 15*time.Minute) {
		t.Fatal("forming")
	}
	out := ClosedOnly([]models.Candle{forming, future}, asOf, 15*time.Minute)
	if len(out) != 0 {
		t.Fatalf("%d", len(out))
	}
}

type fakeFetch struct {
	err error
	n   int
}

func (f fakeFetch) GetPrices(context.Context, string, string, int, time.Time, time.Time) ([]models.Candle, error) {
	if f.err != nil {
		return nil, f.err
	}
	asOf := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	var out []models.Candle
	for i := 0; i < f.n; i++ {
		out = append(out, candle(asOf.Add(-time.Duration(f.n-i)*15*time.Minute), 4300))
	}
	return out, nil
}

func TestSeedReadyFromFetcher(t *testing.T) {
	req := RequirementsFromConfig(&config.Config{})
	asOf := time.Date(2026, 9, 11, 16, 0, 0, 0, time.UTC)
	b := Seed(context.Background(), fakeFetch{n: 30}, "GOLD", req, asOf)
	if b.Status != HistoryReady || b.M15Count < 15 {
		t.Fatalf("%+v", b)
	}
}

func TestSeedFailed(t *testing.T) {
	req := RequirementsFromConfig(nil)
	b := Seed(context.Background(), fakeFetch{err: errUnavailable{}}, "GOLD", req, time.Now().UTC())
	if b.Status != HistoryFailed {
		t.Fatal(b.Status)
	}
}

type errUnavailable struct{}

func (errUnavailable) Error() string { return "unavailable" }
