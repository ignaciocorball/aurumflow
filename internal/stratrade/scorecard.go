package stratrade

type Scorecard struct {
	Identity      string `json:"identity"`
	MarketData    string `json:"market_data"`
	History       string `json:"history"`
	Strategy      string `json:"strategy"`
	Monetary      string `json:"monetary"`
	Risk          string `json:"risk"`
	Open          string `json:"open"`
	Protection    string `json:"protection"`
	Monitoring    string `json:"monitoring"`
	Close         string `json:"close"`
	Reconciliation string `json:"reconciliation"`
	Recovery      string `json:"recovery"`
	Overall       string `json:"overall"`
}

func EmptyScorecard() Scorecard {
	return Scorecard{
		Identity: TrustPASS, MarketData: TrustPASS, History: TrustPartial,
		Strategy: TrustPartial, Monetary: TrustPASS, Risk: TrustPartial,
		Open: TrustUntested, Protection: TrustUntested, Monitoring: TrustUntested,
		Close: TrustUntested, Reconciliation: TrustUntested, Recovery: TrustUntested,
		Overall: TrustPartial,
	}
}

func EvaluateTrust(sc Scorecard, firstLifecyclePASS, mismatch, duplicates, protectiveOK, riskOK, flat, recoveryPASS bool) Scorecard {
	if firstLifecyclePASS {
		sc.Open = TrustPASS
		sc.Close = TrustPASS
		sc.Monitoring = TrustPASS
		sc.Strategy = TrustPASS
	}
	if protectiveOK && firstLifecyclePASS {
		sc.Protection = TrustPASS
	}
	if riskOK {
		sc.Risk = TrustPASS
	}
	if mismatch {
		sc.Reconciliation = TrustFAIL
	} else if firstLifecyclePASS && flat {
		sc.Reconciliation = TrustPASS
	}
	if duplicates {
		sc.Open = TrustFAIL
	}
	if recoveryPASS {
		sc.Recovery = TrustPASS
	}
	sc.Overall = Overall(sc)
	return sc
}

func Overall(sc Scorecard) string {
	vals := []string{
		sc.Identity, sc.MarketData, sc.History, sc.Strategy, sc.Monetary, sc.Risk,
		sc.Open, sc.Protection, sc.Monitoring, sc.Close, sc.Reconciliation, sc.Recovery,
	}
	anyFail := false
	anyUntested := false
	anyPartial := false
	for _, v := range vals {
		switch v {
		case TrustFAIL:
			anyFail = true
		case TrustUntested:
			anyUntested = true
		case TrustPartial:
			anyPartial = true
		}
	}
	if anyFail {
		return TrustFAIL
	}
	if anyUntested || anyPartial {
		return TrustPartial
	}
	return TrustPASS
}

func MayPromoteDemoTrusted(monetaryPASS, lifecyclePASS, mismatch, duplicates, protectiveOK, riskOK, flat, recoveryPASS bool) bool {
	if !monetaryPASS || !lifecyclePASS || mismatch || duplicates || !protectiveOK || !riskOK || !flat || !recoveryPASS {
		return false
	}
	return true
}

func SeverityFor(kind string) string {
	switch kind {
	case "mismatch", "unknown_position", "protective_mismatch", "journal_failure":
		return "CRITICAL"
	case "quote_stale", "capital_reconnect", "worldstate_degraded":
		return "WARNING"
	default:
		return "INFO"
	}
}
