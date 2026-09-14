package demomut

import (
	"testing"

	"aurumflow/internal/execacct"
	"aurumflow/internal/portfoliorisk"
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
	if OriginCalibration != "CALIBRATION_CANARY" || OriginMirror != "DEMO_MIRROR" || OriginGold != "GOLD_STRATEGY" {
		t.Fatal("tags")
	}
}

func TestPortfolioMaxTwo(t *testing.T) {
	c := New()
	req := Request{Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginMirror, Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true, OpenStrategy: 2}
	if err := c.Reserve(req); err != ErrNotEligible {
		t.Fatal(err)
	}
}

func TestUSEquityMaxOne(t *testing.T) {
	c := New()
	req := Request{
		Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginMirror,
		Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true, Market: "US100",
		GroupOpen: map[string]int{portfoliorisk.GroupUSEquity: 1},
	}
	if err := c.Reserve(req); err != ErrNotEligible {
		t.Fatal(err)
	}
}

func TestGoldAndUS100Coexist(t *testing.T) {
	c := New()
	req := Request{
		Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginMirror,
		Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true, Market: "US100",
		OpenStrategy: 1, GroupOpen: map[string]int{portfoliorisk.GroupPrecious: 1},
	}
	if err := c.Reserve(req); err != nil {
		t.Fatal(err)
	}
	c.Release()
	gold := Request{
		Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginGold,
		Market: "GOLD", OpenStrategy: 1, GroupOpen: map[string]int{portfoliorisk.GroupUSEquity: 1},
	}
	if err := c.Reserve(gold); err != nil {
		t.Fatal(err)
	}
}

func TestUS100PlusUS500Blocked(t *testing.T) {
	c := New()
	req := Request{
		Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginMirror,
		Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true, Market: "US500",
		GroupOpen: map[string]int{portfoliorisk.GroupUSEquity: 1},
	}
	if err := c.Reserve(req); err != ErrNotEligible {
		t.Fatal(err)
	}
	req.Market = "US30"
	if err := c.Reserve(req); err != ErrNotEligible {
		t.Fatal(err)
	}
}

func TestHaltNewOrders(t *testing.T) {
	c := New()
	req := Request{Env: "DEMO", Account: demoID(), WantAccountID: "30demo6430", Origin: OriginMirror, Halt: true, Validated: true, MonetaryOK: true, DataFresh: true, Tradeable: true}
	if err := c.Reserve(req); err != ErrHalt {
		t.Fatal(err)
	}
}
