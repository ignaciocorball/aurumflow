package strategy

import (
	"strings"
	"time"
)

const (
	PolicyAll              = "ALL"
	ReasonEmptyMeansAll    = "EMPTY_SESSION_LIST_MEANS_ALL"
	ReasonPolicyAll        = "POLICY_ALL"
	ReasonInAllowed        = "IN_ALLOWED_SESSION"
	ReasonOutsideAllowed   = "OUTSIDE_ALLOWED_SESSIONS"
	NextEligibleNow        = "NOW"
	NextEligibleNA         = "N/A"
)

type SessionEvaluation struct {
	ClockSession string
	ConfigPolicy string
	Eligible     bool
	Reason       string
}

func SessionPolicy(allowed []string) string {
	if len(allowed) == 0 {
		return PolicyAll
	}
	for _, s := range allowed {
		if strings.ToUpper(strings.TrimSpace(s)) == SessionAll {
			return PolicyAll
		}
	}
	var parts []string
	for _, s := range allowed {
		n := strings.ToUpper(strings.TrimSpace(s))
		if n != "" {
			parts = append(parts, n)
		}
	}
	if len(parts) == 0 {
		return PolicyAll
	}
	return strings.Join(parts, ",")
}

func EvaluateSession(now time.Time, allowed []string) SessionEvaluation {
	ev := SessionEvaluation{
		ClockSession: GetSessionInfo(now),
		ConfigPolicy: SessionPolicy(allowed),
		Eligible:     CanTrade(now, allowed),
	}
	if len(allowed) == 0 {
		ev.Reason = ReasonEmptyMeansAll
	} else if ev.ConfigPolicy == PolicyAll {
		ev.Reason = ReasonPolicyAll
	} else if ev.Eligible {
		ev.Reason = ReasonInAllowed
	} else {
		ev.Reason = ReasonOutsideAllowed
	}
	return ev
}

func NextEligibleLabel(ev SessionEvaluation, now time.Time, allowed []string) (string, time.Time) {
	if ev.Eligible || ev.ConfigPolicy == PolicyAll {
		return NextEligibleNA, now
	}
	name, at, ok := NextSessionStart(now, allowed)
	if !ok {
		return NextEligibleNA, time.Time{}
	}
	return name, at
}
