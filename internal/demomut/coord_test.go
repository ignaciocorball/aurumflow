package demomut

import (
	"testing"

	"aurumflow/internal/execacct"
)

func demoID() execacct.Identity {
	return execacct.Identity{AccountID: "30demo6430", Verified: true, MayTrade: true, Resolution: execacct.Explicit}
}

func TestRejectsLIVEAndWrongAccount(t *testing.T) {
	c := New()
	if err := c.Reserve(Request{Live: true, Account: demoID(), Origin: OriginMirror}); err != ErrLive {
		t.Fatal(err)
	}
	if err := c.Reserve(Request{Env: "DEMO", Account: demoID(), WantAccountID: "LIVEID", Origin: OriginCalibration}); err != ErrNoDemo {
		t.Fatal(err)
	}
}

func TestSerializeAndMirrorGates(t *testing.T) {
	c := New()
	req := Request{Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginMirror, Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true}
	if err := c.Reserve(req); err != nil {
		t.Fatal(err)
	}
	if err := c.Reserve(req); err != ErrBusy {
		t.Fatal(err)
	}
	c.Release()
	bad := req
	bad.MonetaryOK = false
	if err := c.Reserve(bad); err != ErrNotEligible {
		t.Fatal(err)
	}
	c.Release()
	stale := req
	stale.Validated = false
	if err := c.Reserve(stale); err != ErrNotEligible {
		t.Fatal("unvalidated")
	}
}

func TestOriginTags(t *testing.T) {
	if OriginCalibration != "CALIBRATION_CANARY" || OriginMirror != "DEMO_MIRROR" {
		t.Fatal("tags")
	}
}

func TestConcurrencyCap(t *testing.T) {
	c := New()
	req := Request{Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginMirror, Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true, OpenStrategy: 2}
	if err := c.Reserve(req); err != ErrNotEligible {
		t.Fatal(err)
	}
}
