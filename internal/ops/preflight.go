package ops

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	PreflightReady         = "READY"
	PreflightReadyWarnings = "READY_WITH_WARNINGS"
	PreflightBlocked       = "BLOCKED"
)

type Check struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Detail  string `json:"detail"`
}

type Preflight struct {
	Result string  `json:"result"`
	Checks []Check `json:"checks"`
}

func EvaluatePreflight(in PreflightInput) Preflight {
	var p Preflight
	add := func(name, st, detail string) {
		p.Checks = append(p.Checks, Check{Name: name, Status: st, Detail: detail})
	}
	if in.TestsOK {
		add("tests", "PASS", in.Version)
	} else {
		add("tests", "WARN", "tests not re-run in this process")
	}
	host := strings.ToLower(in.DemoHost)
	if strings.Contains(host, "demo") || host == "" {
		add("demo_host", "PASS", in.DemoHost)
	} else {
		add("demo_host", "FAIL", "LIVE host is fail-closed")
	}
	if in.LiveRequested {
		add("live", "FAIL", "LIVE requests are impossible")
	} else {
		add("live", "PASS", "LIVE IMPOSSIBLE / FAIL-CLOSED")
	}
	if in.KillSwitch {
		add("kill_switch", "WARN", "HALT_NEW_ORDERS")
	} else {
		add("kill_switch", "PASS", "off")
	}
	if in.OpenPositions < 0 {
		add("positions", "FAIL", "unreadable")
	} else {
		add("positions", "PASS", itoa(in.OpenPositions))
	}
	if in.AccountOK {
		add("account", "PASS", "demo session")
	} else {
		add("account", "WARN", "demo session not confirmed")
	}
	if in.GoldStatus == "" {
		add("gold", "WARN", "GOLD status not attached")
	} else {
		add("gold", "PASS", in.GoldStatus)
	}
	if in.GoldMonetary == "" {
		add("gold_monetary", "WARN", "no cached InstrumentSpec")
	} else if in.GoldMonetary == "RUNTIME_VALIDATED" {
		add("gold_monetary", "PASS", in.GoldMonetary)
	} else {
		add("gold_monetary", "WARN", in.GoldMonetary)
	}
	if in.BTCHealth {
		add("btc_provider", "PASS", in.V1Provider)
	} else {
		add("btc_provider", "WARN", "BTC provider not yet healthy")
	}
	if in.L2Synced && in.L2Deltas > 0 {
		add("l2", "PASS", in.L2Provider)
	} else if in.L2Deltas == 0 {
		add("l2", "WARN", "incremental deltas not yet observed")
	} else {
		add("l2", "WARN", "book unsynced")
	}
	if in.DiskFreeMB >= 256 {
		add("disk", "PASS", "ok")
	} else {
		add("disk", "WARN", "low disk")
	}
	if in.JournalWritable {
		add("journal", "PASS", in.JournalDir)
	} else {
		add("journal", "FAIL", "journal not writable")
	}
	if in.ProspectiveWritable {
		add("prospective", "PASS", "signal_inputs.jsonl")
	} else {
		add("prospective", "FAIL", "prospective recorder not writable")
	}
	fail, warn := 0, 0
	for _, c := range p.Checks {
		if c.Status == "FAIL" {
			fail++
		}
		if c.Status == "WARN" {
			warn++
		}
	}
	switch {
	case fail > 0:
		p.Result = PreflightBlocked
	case warn > 0:
		p.Result = PreflightReadyWarnings
	default:
		p.Result = PreflightReady
	}
	return p
}

type PreflightInput struct {
	TestsOK              bool
	Version              string
	DemoHost             string
	LiveRequested        bool
	KillSwitch           bool
	OpenPositions        int
	GoldStatus           string
	GoldMonetary         string
	AccountOK            bool
	BTCHealth            bool
	V1Provider           string
	L2Provider           string
	L2Synced             bool
	L2Deltas             int
	DiskFreeMB           int
	JournalWritable      bool
	JournalDir           string
	ProspectiveWritable  bool
}

func Writable(dir string) bool {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	p := filepath.Join(dir, ".preflight-write")
	if err := os.WriteFile(p, []byte(runtime.Version()), 0o644); err != nil {
		return false
	}
	_ = os.Remove(p)
	return true
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
