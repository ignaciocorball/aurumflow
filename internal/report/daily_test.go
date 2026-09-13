package report

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDailyFromJSONL(t *testing.T) {
	p := filepath.Join(t.TempDir(), "j.jsonl")
	body := `{"event":"signal_generated","session":"london"}
{"event":"position_closed","pnl":2.5}
{"event":"signal_rejected"}
{"event":"radar_state","state":"ABSORPTION","pressure":20}
`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := FromJSONL(p)
	if err != nil || d.Signals != 1 || d.Trades != 1 || d.Wins != 1 || d.Rejections != 1 || d.RadarStates["ABSORPTION"] != 1 {
		t.Fatalf("%+v %v", d, err)
	}
}
