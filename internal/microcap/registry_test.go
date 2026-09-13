package microcap

import "testing"

func TestAvailableIsNotASignal(t *testing.T) {
	c := BTCFromRuntime(true, true, true, true, true, "HEALTHY")
	if !c.Available() || !c.Healthy() {
		t.Fatal(c)
	}
	if c.Available() && (c.Market != "BTC") {
		t.Fatal("capability leaked a direction")
	}
}

func TestRegistryRoundTrip(t *testing.T) {
	r := New()
	r.Set(BTCFromRuntime(true, true, true, true, true, "HEALTHY"))
	got := r.Get("btc")
	if !got.Trades || !got.IncrementalL2 || !got.Pressure || !got.Exhaustion || !got.Absorption {
		t.Fatal(got)
	}
	if r.Get("GOLD").Available() {
		t.Fatal("gold invented micro")
	}
}
