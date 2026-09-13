package execacct

import (
	"fmt"
	"strings"

	"aurumflow/config"
	"aurumflow/internal/market"
)

const (
	Explicit                 = "EXPLICIT"
	ResolvedSingleCandidate  = "RESOLVED_SINGLE_CANDIDATE"
	SelectionRequired        = "ACCOUNT_SELECTION_REQUIRED"
	Blocked                  = "BLOCKED"
	Accounts0Fallback        = "DISABLED"
)

type Identity struct {
	AccountID  string
	Type       string
	Currency   string
	Status     string
	Preferred  bool
	Resolution string
	Verified   bool
	MayTrade   bool
	Reason     string
}

func Mask(id string) string {
	return config.RedactAccountID(id)
}

func Resolve(accounts []market.AccountInfo, explicitID, wantCurrency string) Identity {
	explicitID = strings.TrimSpace(explicitID)
	wantCurrency = strings.ToUpper(strings.TrimSpace(wantCurrency))
	if len(accounts) == 0 {
		return Identity{Resolution: Blocked, Reason: "no DEMO accounts returned"}
	}
	if explicitID != "" {
		for _, a := range accounts {
			if a.AccountID == explicitID {
				if !compatible(a, wantCurrency) {
					return Identity{
						AccountID: a.AccountID, Type: a.AccountType, Currency: a.Currency, Status: a.Status,
						Preferred: a.Preferred, Resolution: Blocked, Reason: "explicit account fails operational constraints",
					}
				}
				return Identity{
					AccountID: a.AccountID, Type: emptyType(a), Currency: a.Currency, Status: enabledLabel(a),
					Preferred: a.Preferred, Resolution: Explicit, Verified: true, MayTrade: true, Reason: "explicit env/config verified",
				}
			}
		}
		return Identity{Resolution: Blocked, Reason: "explicit account not found"}
	}
	var cand []market.AccountInfo
	for _, a := range accounts {
		if compatible(a, wantCurrency) {
			cand = append(cand, a)
		}
	}
	if len(cand) == 1 {
		a := cand[0]
		return Identity{
			AccountID: a.AccountID, Type: emptyType(a), Currency: a.Currency, Status: enabledLabel(a),
			Preferred: a.Preferred, Resolution: ResolvedSingleCandidate, Verified: false, MayTrade: false,
			Reason: "single compatible candidate; set AURUMFLOW_DEMO_ACCOUNT_ID to trade",
		}
	}
	if len(cand) == 0 {
		return Identity{Resolution: Blocked, Reason: "no compatible DEMO/CFD account"}
	}
	return Identity{Resolution: SelectionRequired, Reason: fmt.Sprintf("%d compatible accounts; refusing accounts[0]/preferred/balance guess", len(cand))}
}

func compatible(a market.AccountInfo, wantCurrency string) bool {
	if !isEnabled(a) {
		return false
	}
	if !isCFD(a) {
		return false
	}
	if wantCurrency != "" && strings.ToUpper(strings.TrimSpace(a.Currency)) != wantCurrency {
		return false
	}
	return true
}

func isEnabled(a market.AccountInfo) bool {
	s := strings.ToUpper(strings.TrimSpace(a.Status))
	return s == "" || s == "ENABLED" || s == "ACTIVE" || s == "OPEN"
}

func isCFD(a market.AccountInfo) bool {
	t := strings.ToUpper(strings.TrimSpace(a.AccountType))
	return t == "" || t == "CFD" || strings.Contains(t, "CFD")
}

func emptyType(a market.AccountInfo) string {
	if strings.TrimSpace(a.AccountType) == "" {
		return "CFD"
	}
	return a.AccountType
}

func enabledLabel(a market.AccountInfo) string {
	if strings.TrimSpace(a.Status) == "" {
		return "ENABLED"
	}
	return a.Status
}

func Instruction(id Identity) string {
	if id.Resolution == ResolvedSingleCandidate {
		return fmt.Sprintf("Set AURUMFLOW_DEMO_ACCOUNT_ID to the unique ENABLED CFD account %s (do not commit it)", Mask(id.AccountID))
	}
	if id.Resolution == SelectionRequired {
		return "ACCOUNT_SELECTION_REQUIRED: set AURUMFLOW_DEMO_ACCOUNT_ID to the intended DEMO account"
	}
	return id.Reason
}
