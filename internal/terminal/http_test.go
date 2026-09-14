package terminal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAllowHTTP(t *testing.T) {
	if !AllowHTTP("GET") || !AllowHTTP("HEAD") || AllowHTTP("POST") || AllowHTTP("DELETE") {
		t.Fatal("http allow")
	}
	if BrokerMutationCapability() != "NONE" {
		t.Fatal("mutation")
	}
	if !ContainsMutationVerb("BUY") || ContainsMutationVerb("GOLD") {
		t.Fatal("verbs")
	}
}

func TestRejectMutations(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ui/snapshot", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	})
	s := httptest.NewServer(rejectMutations(mux))
	defer s.Close()
	resp, err := http.Post(s.URL+"/ui/snapshot", "application/json", strings.NewReader(`{"buy":1}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", resp.StatusCode)
	}
	get, err := http.Get(s.URL + "/ui/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer get.Body.Close()
	if get.StatusCode != 200 {
		t.Fatalf("get %d", get.StatusCode)
	}
}

func TestCanaryExclusion(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	trades := []ClosedTrade{
		{At: now, Market: "GOLD", Origin: "GOLD_STRATEGY", Scope: "BROKER_DEMO", PnL: 12.5, R: 0.4},
		{At: now, Market: "US100", Origin: "CALIBRATION_CANARY", Scope: "BROKER_DEMO", PnL: 99, R: 9, Canary: true, Reason: "CANARY_INFRASTRUCTURE_TEST"},
		{At: now, Market: "US100", Origin: "DEMO_MIRROR", Scope: "SHADOW", PnL: 5, R: 0.2},
	}
	p := AggregatePnL(trades, "BROKER_DEMO", "ALL", now, 1000)
	if p.AllTime != 12.5 || p.TradeCount != 1 {
		t.Fatalf("broker demo %+v", p)
	}
	sh := AggregatePnL(trades, "SHADOW", "ALL", now, 1000)
	if sh.AllTime != 5 {
		t.Fatalf("shadow %v", sh.AllTime)
	}
}

func TestNarrativeDerivability(t *testing.T) {
	s := Snapshot{
		World: WorldView{Dislocation: "NORMAL"},
		Markets: []Market{
			{ID: "J225", Label: "J225", Region: "JAPAN", Salience: 83, Attention: 45, Trend: "STRONG_DOWN", TrendDisplay: "Strong down"},
			{ID: "OIL_CRUDE", Label: "OIL", Salience: 73, Attention: 45, Trend: "UP"},
			{ID: "US100", Label: "US100", Salience: 71, Attention: 45, DemoStatus: "DEMO_ELIGIBLE", Pilot: true},
			{ID: "GOLD", Label: "GOLD", Attention: 60, Salience: 36, Setup: "NONE"},
		},
		Portfolio: Portfolio{OpenRisk: 0},
		Execution: ExecutionView{Positions: []Position{
			{Market: "GOLD", Waiting: "STRATEGY WAITING"},
			{Market: "US100", Armed: "ON", Eligible: "YES"},
		}},
		Health: []HealthItem{{ID: "capital", Label: "Capital", Status: "HEALTHY"}},
	}
	n := BuildNarratives(s)
	if !strings.Contains(n.World, "broadly stable") {
		t.Fatal(n.World)
	}
	if !strings.Contains(n.World, "J225") || !strings.Contains(n.World, "downside") {
		t.Fatal(n.World)
	}
	if !strings.Contains(n.Execution, "No positions are open") {
		t.Fatal(n.Execution)
	}
	if !strings.Contains(n.Health, "healthy") {
		t.Fatal(n.Health)
	}
	joined := strings.Join(n.Matters, " ")
	if !strings.Contains(joined, "GOLD") || !strings.Contains(joined, "US100") {
		t.Fatal(joined)
	}
}

func TestDisplayCopy(t *testing.T) {
	if Display("RUNTIME_VALIDATED") != "Runtime validated" {
		t.Fatal(Display("RUNTIME_VALIDATED"))
	}
	if Display("data_frozen:M15") != "M15 data temporarily stale" {
		t.Fatal(Display("data_frozen:M15"))
	}
}

func TestJournalScanExcludesCanary(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "journals"), 0o755)
	line := `{"ts":"2026-09-14T01:17:48Z","event":"position_closed","epic":"US100","reason":"CANARY_INFRASTRUCTURE_TEST","pnl":1.5}`
	os.WriteFile(filepath.Join(dir, "journals", "canary-lifecycle.jsonl"), []byte(line+"\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "journals", "demo-week.jsonl"), []byte(`{"ts":"2026-09-14T02:00:00Z","event":"signal_rejected"}`+"\n"), 0o644)
	tr := LoadTrades(dir)
	p := AggregatePnL(tr, "BROKER_DEMO", "ALL", time.Date(2026, 9, 14, 3, 0, 0, 0, time.UTC), 1000)
	if p.HasBrokerTrades || p.TradeCount != 0 {
		t.Fatalf("canary leaked %+v trades=%d", p, p.TradeCount)
	}
}

func TestSnapshotHydrationFromHub(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/status":
			json.NewEncoder(w).Encode(map[string]any{"demo_balance": 999.96, "capital_environment": "DEMO", "live_possible": "IMPOSSIBLE / FAIL-CLOSED", "positions_known": true, "strategy_ready": true, "gold_bid": 4344.0, "gold_ask": 4344.5, "market_status": "TRADEABLE"})
		case "/api/world":
			json.NewEncoder(w).Encode(map[string]any{
				"Valid": "CURRENT_WORLD_STATE_VALID", "Hash": "abc", "Risk": "MIXED", "Confidence": 0.8,
				"Markets": map[string]any{"GOLD": map[string]any{"Market": "GOLD", "Mid": 4344.25, "Region": "GLOBAL", "PriceTrend": "FLAT", "DataQuality": "HEALTHY"}},
				"Opportunity": []any{map[string]any{"Market": "GOLD", "Attention": 60.0, "Salience": 36.0, "Coverage": 62.5}},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{})
		}
	}))
	defer upstream.Close()
	h := NewHub(NewReadClient(time.Second), Endpoints{Gold: upstream.URL, Intel: upstream.URL, US100: upstream.URL, Root: t.TempDir()}, nil, filepath.Join(t.TempDir(), "missing.json"), false)
	h.refreshAll(time.Date(2026, 9, 14, 3, 0, 0, 0, time.UTC))
	snap := h.Snapshot()
	if snap.Portfolio.Equity != 999.96 {
		t.Fatalf("equity %v", snap.Portfolio.Equity)
	}
	found := false
	for _, m := range snap.Markets {
		if m.ID == "GOLD" && m.Attention == 60 && m.Salience == 36 {
			found = true
		}
	}
	if !found {
		t.Fatalf("gold merge %+v", snap.Markets)
	}
	if snap.Environment.Demo != true {
		t.Fatal("demo")
	}
}

func TestServerReadOnlyAndSnapshot(t *testing.T) {
	h := NewHub(NewReadClient(time.Second), Endpoints{Root: t.TempDir()}, nil, "", false)
	s := NewServer("127.0.0.1:0", h)
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && (s.Addr() == "" || s.Addr() == "127.0.0.1:0") {
		time.Sleep(10 * time.Millisecond)
	}
	resp, err := http.Post("http://"+s.Addr()+"/ui/snapshot", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("post %d", resp.StatusCode)
	}
	get, err := http.Get("http://" + s.Addr() + "/ui/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer get.Body.Close()
	if get.StatusCode != 200 {
		t.Fatalf("get %d", get.StatusCode)
	}
	hz, err := http.Get("http://" + s.Addr() + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer hz.Body.Close()
	var body map[string]any
	json.NewDecoder(hz.Body).Decode(&body)
	if body["mutation"] != "NONE" {
		t.Fatalf("%v", body)
	}
	_ = s.Shutdown()
}
