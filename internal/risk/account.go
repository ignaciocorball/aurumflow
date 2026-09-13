package risk

import (
	"fmt"

	"aurumflow/internal/market"
)

// SelectAccount returns the exact configured account. No silent fallback to accounts[0].
func SelectAccount(accounts []market.AccountInfo, wantID, currentID string) (market.AccountInfo, error) {
	if len(accounts) == 0 {
		return market.AccountInfo{}, fmt.Errorf("account selection: no accounts returned")
	}
	id := wantID
	if id == "" {
		id = currentID
	}
	if id == "" {
		return market.AccountInfo{}, fmt.Errorf("account selection: api.account_id is required; refusing accounts[0] fallback")
	}
	for _, a := range accounts {
		if a.AccountID == id {
			return a, nil
		}
	}
	return market.AccountInfo{}, fmt.Errorf("account selection: account %s not found", redact(id))
}

func redact(id string) string {
	if len(id) <= 4 {
		return "****"
	}
	return id[:2] + "****" + id[len(id)-2:]
}
