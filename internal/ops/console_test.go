package ops

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func waitServer(t *testing.T, s *Server) string {
	t.Helper()
	errCh := make(chan error, 1)
	go func() { errCh <- s.ListenAndServe() }()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if s.Addr() != "" && s.Addr() != "127.0.0.1:0" {
			if resp, err := http.Get("http://" + s.Addr() + "/healthz"); err == nil {
				resp.Body.Close()
				return s.Addr()
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("server")
	return ""
}

func TestConsoleSafetyAndReadOnly(t *testing.T) {
	st := NewStatus()
	st.ExecutionMode = "DEMO"
	st.KillSwitch = true
	st.OpenPositions = 0
	st.LastV1Class = "FLOW_NEUTRAL"
	s := NewServer("127.0.0.1:0", st)
	addr := waitServer(t, s)
	resp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	html := string(body)
	if !strings.Contains(html, "LIVE IMPOSSIBLE") || !strings.Contains(html, "SHADOW") {
		t.Fatal("safety bar")
	}
	if !strings.Contains(html, "Global capital surface") {
		t.Fatal("global tab")
	}
	if !strings.Contains(html, "Official sources") || !strings.Contains(html, "CONTEXT CALENDAR") {
		t.Fatal("official panels")
	}
	if !strings.Contains(html, "WORLD MARKET TAPE") || !strings.Contains(html, "PRICE LEADERSHIP") || !strings.Contains(html, "CAPITAL FLOW EVIDENCE") {
		t.Fatal("live tape panels")
	}
	if !strings.Contains(html, "SESSION STRIP") || !strings.Contains(html, "TOKYO") || !strings.Contains(html, "CORRELATION ≠ CAUSATION") {
		t.Fatal("p82 ui")
	}
	if !strings.Contains(html, "$300 DEMO") || !strings.Contains(html, "account_currency_risk") {
		t.Fatal("risk units")
	}
	for _, bad := range []string{"BUY", "SELL", "CLOSE", "FLATTEN", "CST", "X-SECURITY-TOKEN", "password", "api_key"} {
		if strings.Contains(html, bad) && bad != "CLOSE" {
			t.Fatalf("must not expose %s", bad)
		}
	}
	if strings.Contains(html, "BUY") || strings.Contains(html, "SELL") || strings.Contains(html, "FLATTEN") {
		t.Fatal("mutation controls")
	}
	post, err := http.Post("http://"+addr+"/api/status", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	post.Body.Close()
	if post.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("mutation %d", post.StatusCode)
	}
	gold, err := http.Get("http://" + addr + "/api/gold")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(gold.Body)
	gold.Body.Close()
	if strings.Contains(string(raw), "CST") || strings.Contains(string(raw), "X-SECURITY") {
		t.Fatal("secrets")
	}
}

func TestObservatoryBindingsAndStates(t *testing.T) {
	st := NewStatus()
	st.MarketStatus = "CLOSED"
	st.GoldBid, st.GoldAsk, st.GoldSpread = 0, 0, 12
	st.L2Spread, st.L2Mid, st.L2QuotesOK = 1.25, 100, true
	st.BookSynced = false
	st.L2ProxyQuality = "DEGRADED"
	st.LatencyP95 = 612
	st.DemoBalance = 0
	s := NewServer("127.0.0.1:0", st)
	s.Set(st)
	addr := waitServer(t, s)

	htmlResp, err := http.Get("http://" + addr + "/")
	if err != nil {
		t.Fatal(err)
	}
	htmlb, _ := io.ReadAll(htmlResp.Body)
	htmlResp.Body.Close()
	html := string(htmlb)
	if !strings.Contains(html, "l2_spread") {
		t.Fatal("L2 spread must bind l2_spread")
	}
	if !strings.Contains(html, "num(l2ok(s), s.l2_spread") {
		t.Fatal("L2 metrics must read l2_spread, not gold_spread")
	}
	if !strings.Contains(html, "—") || !strings.Contains(html, "WAITING") {
		t.Fatal("missing unavailable glyphs")
	}
	if !strings.Contains(html, "Awaiting broker TRADEABLE") || !strings.Contains(html, "CLOSED") {
		t.Fatal("gold closed state")
	}
	if !strings.Contains(html, "L2 quality") || !strings.Contains(html, "DEGRADED") {
		t.Fatal("data quality degraded binding")
	}
	if !strings.Contains(html, "BOOK UNSYNCED") && !strings.Contains(html, "book_synced") {
		t.Fatal("unsynced binding")
	}
	if strings.Contains(html, "BUY") || strings.Contains(html, "SELL") || strings.Contains(html, "FLATTEN") {
		t.Fatal("mutation controls")
	}
	if strings.Contains(html, "password") || strings.Contains(html, "api_key") || strings.Contains(html, "CST") {
		t.Fatal("secrets")
	}

	book, err := http.Get("http://" + addr + "/api/book")
	if err != nil {
		t.Fatal(err)
	}
	var bm map[string]any
	if err := json.NewDecoder(book.Body).Decode(&bm); err != nil {
		t.Fatal(err)
	}
	book.Body.Close()
	if bm["spread"] != 1.25 {
		t.Fatalf("book spread binding %+v", bm["spread"])
	}

	gold, err := http.Get("http://" + addr + "/api/gold")
	if err != nil {
		t.Fatal(err)
	}
	var gm map[string]any
	if err := json.NewDecoder(gold.Body).Decode(&gm); err != nil {
		t.Fatal(err)
	}
	gold.Body.Close()
	if gm["market_status"] != "CLOSED" {
		t.Fatalf("closed %v", gm["market_status"])
	}
	if _, ok := gm["bid"]; ok {
		t.Fatal("closed gold must omit bid")
	}
	if _, ok := gm["demo_balance"]; ok {
		t.Fatal("unavailable balance must omit numeric zero")
	}

	st.MarketStatus = "TRADEABLE"
	st.GoldBid, st.GoldAsk = 2300, 2301
	s.Set(st)
	gold2, err := http.Get("http://" + addr + "/api/gold")
	if err != nil {
		t.Fatal(err)
	}
	var gm2 map[string]any
	if err := json.NewDecoder(gold2.Body).Decode(&gm2); err != nil {
		t.Fatal(err)
	}
	gold2.Body.Close()
	if gm2["market_status"] != "TRADEABLE" || gm2["quotes_ok"] != true {
		t.Fatalf("tradeable %+v", gm2)
	}

	req, _ := http.NewRequest(http.MethodGet, "http://"+addr+"/api/stream", nil)
	sser, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer sser.Body.Close()
	rd := bufio.NewReader(sser.Body)
	line, err := rd.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	raw := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	var snap map[string]any
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		t.Fatal(err, raw)
	}
	for _, k := range []string{"l2_spread", "gold_quotes_ok", "l2_quotes_ok", "decision_why", "book_synced", "l2_proxy_quality"} {
		if _, ok := snap[k]; !ok {
			t.Fatalf("sse schema missing %s", k)
		}
	}
}

func TestAPISerializationAndSSE(t *testing.T) {
	s := NewServer("127.0.0.1:0", NewStatus())
	addr := waitServer(t, s)
	postW, err := http.Post("http://"+addr+"/api/world", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	postW.Body.Close()
	if postW.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("world mutation %d", postW.StatusCode)
	}
	for _, p := range []string{"/api/status", "/api/intelligence", "/api/book", "/api/prospective", "/api/research", "/api/timeseries", "/api/events", "/api/world", "/api/regions", "/api/opportunities", "/api/context/sources", "/api/institutional"} {
		resp, err := http.Get("http://" + addr + p)
		if err != nil {
			t.Fatal(p, err)
		}
		var m map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
			t.Fatal(p, err)
		}
		resp.Body.Close()
	}
	req, _ := http.NewRequest(http.MethodGet, "http://"+addr+"/api/stream", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatal(resp.Header.Get("Content-Type"))
	}
	rd := bufio.NewReader(resp.Body)
	line, err := rd.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "data:") {
		t.Fatalf("sse %q %v", line, err)
	}
}
