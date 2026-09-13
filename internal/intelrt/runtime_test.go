package intelrt

import (
	"testing"

	"aurumflow/internal/shadow"
)

func TestUnifiedNoExecutionProvider(t *testing.T) {
	r := New()
	if r.CanMutateBroker() || r.HasExecutionProvider() || shadow.HasExecutionProvider(r) {
		t.Fatal("execution")
	}
	if BTCMicroHealthy(true, true, true, true, true, false) {
		t.Fatal("absorption missing still healthy")
	}
	if !BTCMicroHealthy(true, true, true, true, true, true) {
		t.Fatal("all healthy")
	}
}

func TestProviderIsolation(t *testing.T) {
	if BTCMicroHealthy(false, false, false, false, false, false) {
		t.Fatal("binance down must not advertise BTC micro")
	}
	if New().CanMutateBroker() {
		t.Fatal("capital fail must not grant mutation")
	}
}
