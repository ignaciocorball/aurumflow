package capitalhist

import (
	"testing"
	"time"
)

func TestNormalizeQuote(t *testing.T) {
	s := &StreamingProvider{}
	ev := s.NormalizeQuote("GOLD", 1900, 1901, time.Unix(1, 0).UTC())
	if ev.Instrument != "GOLD" || ev.Prov.Provider == "" {
		t.Fatalf("%+v", ev)
	}
}
