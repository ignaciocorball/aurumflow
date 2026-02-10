package strategy

import (
	"strings"
	"time"
)

const (
	SessionLondon   = "LONDON"
	SessionNY       = "NY"
	SessionAsia     = "ASIA"
	SessionAll      = "ALL"
	SessionLondonNY = "LONDON+NY" // overlap; hard-blocked in loop/backtest
)

// TradingSession represents a trading session with UTC hours.
type TradingSession struct {
	Name      string
	StartHour int // UTC hour (0-23)
	EndHour   int // UTC hour (0-23)
}

var (
	// Standard trading sessions (UTC)
	LondonSession = TradingSession{Name: SessionLondon, StartHour: 8, EndHour: 17}   // 08:00-17:00 UTC
	NYSession     = TradingSession{Name: SessionNY, StartHour: 13, EndHour: 22}      // 13:00-22:00 UTC
	AsiaSession   = TradingSession{Name: SessionAsia, StartHour: 0, EndHour: 8}      // 00:00-08:00 UTC (Tokyo/Sydney)
)

// IsInSession checks if the given UTC time is within the specified session.
func IsInSession(t time.Time, session TradingSession) bool {
	hour := t.UTC().Hour()
	if session.StartHour <= session.EndHour {
		return hour >= session.StartHour && hour < session.EndHour
	}
	// Handle overnight sessions (e.g., Asia 00:00-08:00)
	return hour >= session.StartHour || hour < session.EndHour
}

// GetCurrentSession returns the active trading session(s) for the given UTC time.
func GetCurrentSession(t time.Time) []string {
	var active []string
	if IsInSession(t, LondonSession) {
		active = append(active, SessionLondon)
	}
	if IsInSession(t, NYSession) {
		active = append(active, SessionNY)
	}
	if IsInSession(t, AsiaSession) {
		active = append(active, SessionAsia)
	}
	return active
}

// CanTrade checks if trading is allowed based on configured sessions and current time.
// If tradingSessions contains "ALL", always returns true.
// Otherwise, returns true only if current session is in the allowed list.
func CanTrade(t time.Time, tradingSessions []string) bool {
	if len(tradingSessions) == 0 {
		return true
	}
	// Check if "ALL" is in the list
	for _, s := range tradingSessions {
		if strings.ToUpper(strings.TrimSpace(s)) == SessionAll {
			return true
		}
	}
	// Get current active sessions
	currentSessions := GetCurrentSession(t)
	if len(currentSessions) == 0 {
		return false
	}
	// Check if any current session is in allowed list
	for _, current := range currentSessions {
		for _, allowed := range tradingSessions {
			if strings.ToUpper(strings.TrimSpace(allowed)) == current {
				return true
			}
		}
	}
	return false
}

// GetSessionInfo returns a human-readable description of current session(s).
func GetSessionInfo(t time.Time) string {
	sessions := GetCurrentSession(t)
	if len(sessions) == 0 {
		return "OFF_HOURS"
	}
	return strings.Join(sessions, "+")
}

// sessionStartOrder defines sessions in chronological order by start hour (UTC) for "next start" calculation.
var sessionStartOrder = []TradingSession{AsiaSession, LondonSession, NYSession}

// NextSessionStart returns the next session start time from now, considering only allowed sessions.
// If allowed contains "ALL", all sessions (ASIA, LONDON, NY) are considered.
// Returns the session name, the UTC time when it starts, and ok=true if a next start was found.
// If allowed is empty or no next start can be computed, ok=false.
func NextSessionStart(now time.Time, allowed []string) (session string, at time.Time, ok bool) {
	if len(allowed) == 0 {
		return "", time.Time{}, false
	}
	now = now.UTC()
	allowedSet := make(map[string]bool)
	for _, s := range allowed {
		norm := strings.ToUpper(strings.TrimSpace(s))
		if norm == SessionAll {
			allowedSet[SessionAsia] = true
			allowedSet[SessionLondon] = true
			allowedSet[SessionNY] = true
			break
		}
		allowedSet[norm] = true
	}
	var bestSession string
	var bestAt time.Time
	first := true
	for _, sess := range sessionStartOrder {
		if !allowedSet[sess.Name] {
			continue
		}
		// Next start for this session: today at StartHour:00:00 or tomorrow if now is past that
		y, m, d := now.Date()
		startToday := time.Date(y, m, d, sess.StartHour, 0, 0, 0, time.UTC)
		var nextStart time.Time
		if now.Before(startToday) {
			nextStart = startToday
		} else {
			nextStart = startToday.Add(24 * time.Hour)
		}
		if first || nextStart.Before(bestAt) {
			bestAt = nextStart
			bestSession = sess.Name
			first = false
		}
	}
	if first {
		return "", time.Time{}, false
	}
	return bestSession, bestAt, true
}
