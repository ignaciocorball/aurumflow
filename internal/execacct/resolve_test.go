package execacct

import (
	"strings"
	"testing"

	"aurumflow/internal/market"
)

func TestExplicitValid(t *testing.T) {
	accs := []market.AccountInfo{
		{AccountID: "AAA111", AccountType: "CFD", Currency: "USD", Status: "ENABLED"},
		{AccountID: "BBB222", AccountType: "CFD", Currency: "EUR", Status: "ENABLED"},
	}
	id := Resolve(accs, "AAA111", "USD")
	if !id.MayTrade || !id.Verified || id.Resolution != Explicit {
		t.Fatalf("%+v", id)
	}
	if strings.Contains(Mask(id.AccountID), "AAA111") {
		t.Fatal("mask leaked")
	}
}

func TestExplicitInvalid(t *testing.T) {
	accs := []market.AccountInfo{{AccountID: "AAA111", AccountType: "CFD", Currency: "USD", Status: "ENABLED"}}
	if id := Resolve(accs, "ZZZ", "USD"); id.MayTrade || id.Resolution != Blocked {
		t.Fatalf("%+v", id)
	}
	if id := Resolve(accs, "AAA111", "EUR"); id.MayTrade {
		t.Fatal("currency mismatch must block")
	}
}

func TestSingleCandidateNotSilentTrade(t *testing.T) {
	accs := []market.AccountInfo{{AccountID: "ONLY1", AccountType: "CFD", Currency: "USD", Status: "ENABLED"}}
	id := Resolve(accs, "", "USD")
	if id.Resolution != ResolvedSingleCandidate || id.MayTrade || id.Verified {
		t.Fatalf("%+v", id)
	}
	if !strings.Contains(Instruction(id), "AURUMFLOW_DEMO_ACCOUNT_ID") {
		t.Fatal(Instruction(id))
	}
}

func TestMultipleRequiresSelection(t *testing.T) {
	accs := []market.AccountInfo{
		{AccountID: "A1", AccountType: "CFD", Currency: "USD", Preferred: true, Balance: market.Balance{Balance: 99}},
		{AccountID: "B1", AccountType: "CFD", Currency: "USD", Preferred: false, Balance: market.Balance{Balance: 1}},
	}
	id := Resolve(accs, "", "USD")
	if id.Resolution != SelectionRequired || id.MayTrade || id.AccountID != "" {
		t.Fatalf("must not pick preferred/largest: %+v", id)
	}
}

func TestAccounts0FallbackDisabled(t *testing.T) {
	if Accounts0Fallback != "DISABLED" {
		t.Fatal(Accounts0Fallback)
	}
}
