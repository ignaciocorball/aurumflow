package report

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

type Daily struct {
	RuntimeEvents   int
	Signals         int
	Trades          int
	Wins            int
	Losses          int
	GrossPnL        float64
	Rejections      int
	Errors          int
	KillSwitch      int
	Recoveries      int
	RadarStates     map[string]int
	PressureSum     float64
	PressureN       int
	Sessions        map[string]int
	LastError       string
}

func FromJSONL(path string) (Daily, error) {
	d := Daily{RadarStates: map[string]int{}, Sessions: map[string]int{}}
	f, err := os.Open(path)
	if err != nil {
		return d, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		d.RuntimeEvents++
		ev := strings.ToLower(asString(raw["event"]))
		switch {
		case strings.Contains(ev, "signal_generated"):
			d.Signals++
			if s := asString(raw["session"]); s != "" {
				d.Sessions[s]++
			}
		case strings.Contains(ev, "signal_rejected"):
			d.Rejections++
		case strings.Contains(ev, "position_closed") || strings.Contains(ev, "positionclosed"):
			d.Trades++
			pnl := asFloat(raw["pnl"])
			if pnl == 0 {
				pnl = asFloat(raw["realized_pnl"])
			}
			d.GrossPnL += pnl
			if pnl > 0 {
				d.Wins++
			} else if pnl < 0 {
				d.Losses++
			}
		case strings.Contains(ev, "error"):
			d.Errors++
			d.LastError = asString(raw["error"])
		case strings.Contains(ev, "kill"):
			d.KillSwitch++
		case strings.Contains(ev, "recover"):
			d.Recoveries++
		case strings.Contains(ev, "radar"):
			d.RadarStates[asString(raw["state"])]++
			if p := asFloat(raw["pressure"]); p != 0 {
				d.PressureSum += p
				d.PressureN++
			}
		}
	}
	return d, sc.Err()
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return 0
	}
}
