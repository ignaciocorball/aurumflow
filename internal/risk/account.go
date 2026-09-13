package risk

import (
	"fmt"
	"strings"

	"aurumflow/internal/market"
)

// SelectAccount returns the exact configured account. No silent fallback to accounts[0].
// SelectDemoAccount chooses a DEMO CFD account from broker metadata.
// It never silently returns accounts[0]. currentID is the session currentAccountId.
func SelectDemoAccount(accounts []market.AccountInfo, currentID string) (market.AccountInfo, error) {
	if len(accounts) == 0 {
		return market.AccountInfo{}, fmt.Errorf("account selection: no DEMO accounts returned")
	}
	if len(accounts) == 1 {
		return accounts[0], nil
	}

	var cfd []market.AccountInfo
	for _, a := range accounts {
		if isCFDAccount(a) {
			cfd = append(cfd, a)
		}
	}
	if len(cfd) == 1 {
		return cfd[0], nil
	}

	var preferred []market.AccountInfo
	pool := accounts
	if len(cfd) > 0 {
		pool = cfd
	}
	for _, a := range pool {
		if a.Preferred {
			preferred = append(preferred, a)
		}
	}
	if len(preferred) == 1 {
		return preferred[0], nil
	}

	if currentID != "" {
		for _, a := range pool {
			if a.AccountID == currentID {
				return a, nil
			}
		}
	}
	return market.AccountInfo{}, fmt.Errorf("account selection: %d DEMO accounts; cannot uniquely choose a CFD/preferred account without an explicit id", len(accounts))
}

func isCFDAccount(a market.AccountInfo) bool {
	t := strings.ToUpper(strings.TrimSpace(a.AccountType))
	return t == "" || t == "CFD" || strings.Contains(t, "CFD")
}

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
