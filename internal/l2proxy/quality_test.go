package l2proxy

import "testing"

func TestClassifyOperational(t *testing.T) {
	if Classify(Input{BookSynced: true, ProviderOK: true, BookAgeMs: 200, ReceiveLatencyP95: 80}).Quality != Good {
		t.Fatal("good")
	}
	if Classify(Input{BookSynced: true, ProviderOK: true, BookAgeMs: 2000}).Quality != Degraded {
		t.Fatal("degraded")
	}
	if Classify(Input{BookSynced: false, ProviderOK: true}).Quality != Unusable {
		t.Fatal("unsynced")
	}
	if Classify(Input{BookSynced: true, ProviderOK: true, AbsBasisZ: 7}).Quality != Unusable {
		t.Fatal("basis")
	}
}

func TestBasisPastOnly(t *testing.T) {
	b := NewBasis(32)
	var lastZ float64
	for i := 0; i < 20; i++ {
		_, lastZ = b.Observe(100, 100.1)
	}
	if lastZ != 0 && lastZ > 20 {
		t.Fatalf("z=%v", lastZ)
	}
	_, z := b.Observe(100, 110)
	if z <= 0 {
		t.Fatalf("dislocation z=%v", z)
	}
}
