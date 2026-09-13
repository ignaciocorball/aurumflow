package optionsctx

import "testing"

func TestNotConnected(t *testing.T) {
	p := New()
	if p.Realtime != "NOT_CONNECTED" || p.Historical != "NOT_CONNECTED" {
		t.Fatalf("%+v", p)
	}
}
