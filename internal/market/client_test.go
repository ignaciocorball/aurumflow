package market

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aurumflow/config"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient(srv.URL)
	c.APIKey = "test-key"
	c.HTTPClient = srv.Client()
	return srv, c
}

func TestNewDemoClient_PinnedToDemoHost(t *testing.T) {
	c := NewDemoClient()
	if !config.IsDemoCapitalHost(c.BaseURL) {
		t.Fatalf("demo client host=%s", c.BaseURL)
	}
	if config.IsLiveCapitalHost(c.BaseURL) {
		t.Fatal("demo client must not be live")
	}
}

func TestDo_RefusesLiveHost(t *testing.T) {
	c := NewClient(config.LiveAPIHost)
	_, err := c.Do(context.Background(), "GET", "/api/v1/ping", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "LIVE") {
		t.Fatalf("live host must refuse, got %v", err)
	}
}

func TestDo_NoLiveURLInOutboundDemo(t *testing.T) {
	var sawLive bool
	srv, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Host, "api-capital.backend-capital.com") && !strings.Contains(r.Host, "demo") {
			sawLive = true
		}
		if strings.Contains(r.URL.String(), "api-capital.backend-capital.com") && !strings.Contains(r.URL.String(), "demo") {
			sawLive = true
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	_ = srv
	_, err := c.Do(context.Background(), "GET", "/api/v1/ping", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if sawLive {
		t.Fatal("production LIVE URL appeared in outbound request")
	}
	if config.IsLiveCapitalHost(c.BaseURL) {
		t.Fatal("test client must not be live")
	}
}

func TestCreateSession_CapturesHeaders(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/session" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-CAP-API-KEY") != "test-key" {
			t.Error("missing API key")
		}
		w.Header().Set("CST", "cst-1")
		w.Header().Set("X-SECURITY-TOKEN", "tok-1")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"currentAccountId": "ACC1",
			"accounts": []map[string]any{
				{"accountId": "ACC1", "balance": map[string]any{"balance": 10000.0, "available": 9000.0}},
			},
		})
	})
	sess, err := c.CreateSession(context.Background(), "id", "pw")
	if err != nil {
		t.Fatal(err)
	}
	if c.CST != "cst-1" || c.SecurityToken != "tok-1" {
		t.Fatalf("tokens not captured: %s %s", c.CST, c.SecurityToken)
	}
	if sess.CurrentAccountID != "ACC1" {
		t.Fatalf("account=%s", sess.CurrentAccountID)
	}
}

func TestCreateSession_MissingHeadersStillParses(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"currentAccountId": "ACC1", "accounts": []any{}})
	})
	_, err := c.CreateSession(context.Background(), "id", "pw")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPing(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/ping" {
			t.Errorf("path=%s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestGetAccounts(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accounts": []map[string]any{
				{"accountId": "A", "balance": map[string]any{"balance": 1.0}},
			},
		})
	})
	ar, err := c.GetAccounts(context.Background())
	if err != nil || len(ar.Accounts) != 1 || ar.Accounts[0].AccountID != "A" {
		t.Fatalf("%v %#v", err, ar)
	}
}

func TestSwitchAccount(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method=%s", r.Method)
		}
		w.Header().Set("CST", "cst-2")
		w.Header().Set("X-SECURITY-TOKEN", "tok-2")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	})
	if err := c.SwitchAccount(context.Background(), "ACC2"); err != nil {
		t.Fatal(err)
	}
	if c.CST != "cst-2" {
		t.Fatalf("cst=%s", c.CST)
	}
}

func TestGetMarketDetailsAndPrices(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v1/markets/GOLD"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"instrument": map[string]any{"epic": "GOLD", "name": "Gold", "type": "COMMODITIES", "lotSize": 1},
				"dealingRules": map[string]any{
					"minDealSize": map[string]any{"value": 0.1},
					"maxDealSize": map[string]any{"value": 100},
					"minSizeIncrement": map[string]any{"value": 0.1},
				},
				"snapshot": map[string]any{"bid": 2000.0, "offer": 2000.5, "marketStatus": "TRADEABLE"},
			})
		case strings.HasPrefix(r.URL.Path, "/api/v1/prices/GOLD"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"prices": []map[string]any{{
					"snapshotTimeUTC": "2026-01-01T00:00:00",
					"openPrice":       map[string]any{"bid": 1.0, "ask": 1.1},
					"highPrice":       map[string]any{"bid": 2.0, "ask": 2.1},
					"lowPrice":        map[string]any{"bid": 0.5, "ask": 0.6},
					"closePrice":      map[string]any{"bid": 1.5, "ask": 1.6},
					"lastTradedVolume": 10,
				}},
			})
		default:
			t.Errorf("path=%s", r.URL.Path)
			w.WriteHeader(404)
		}
	})
	md, err := c.GetMarketDetails(context.Background(), "GOLD")
	if err != nil || md.Instrument.Epic != "GOLD" {
		t.Fatalf("%v %#v", err, md)
	}
	spec, err := SpecFromDetails(md, 1.0)
	if err != nil || !spec.SizingComplete() {
		t.Fatalf("spec %v complete=%v err=%v", spec, spec.SizingComplete(), err)
	}
	candles, err := c.GetPrices(context.Background(), "GOLD", ResolutionMinute15, 10, time.Time{}, time.Time{})
	if err != nil || len(candles) != 1 || candles[0].Close != 1.5 {
		t.Fatalf("%v %#v", err, candles)
	}
}

func TestGetPositionsAndGetPosition(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/positions/D1" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"position": map[string]any{"dealId": "D1", "size": 0.1, "direction": "BUY", "level": 2000, "epic": "GOLD"},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"positions": []map[string]any{
				{"position": map[string]any{"dealId": "D1", "size": 0.1, "direction": "BUY", "level": 2000},
					"market": map[string]any{"epic": "GOLD"}},
			},
		})
	})
	pr, err := c.GetPositions(context.Background())
	if err != nil || len(pr.Positions) != 1 || pr.Positions[0].GetEpic() != "GOLD" {
		t.Fatalf("%v %#v", err, pr)
	}
	one, err := c.GetPosition(context.Background(), "D1")
	if err != nil || one.Position.DealID != "D1" {
		t.Fatalf("%v %#v", err, one)
	}
}

func TestCloseAndUpdatePosition(t *testing.T) {
	var methods []string
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		_ = json.NewEncoder(w).Encode(map[string]any{"dealReference": "REF1"})
	})
	ack, err := c.ClosePosition(context.Background(), "D1")
	if err != nil || ack.DealReference != "REF1" {
		t.Fatalf("close %v %#v", err, ack)
	}
	sl := 1990.0
	ack, err = c.UpdatePosition(context.Background(), "D1", UpdatePositionRequest{StopLevel: &sl})
	if err != nil || ack.DealReference != "REF1" {
		t.Fatalf("update %v %#v", err, ack)
	}
	if methods[0] != http.MethodDelete || methods[1] != http.MethodPut {
		t.Fatalf("methods=%v", methods)
	}
}

func TestDo_401(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`unauthorized`))
	})
	_, err := c.Do(context.Background(), "GET", "/api/v1/ping", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("got %v", err)
	}
}

func TestDo_4xx(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`bad`))
	})
	_, err := c.Do(context.Background(), "GET", "/api/v1/ping", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("got %v", err)
	}
}

func TestDo_5xx(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`err`))
	})
	_, err := c.Do(context.Background(), "GET", "/api/v1/ping", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("got %v", err)
	}
}

func TestDo_429RetriesThenFails(t *testing.T) {
	n := 0
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		n++
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`slow`))
	})
	_, err := c.Do(context.Background(), "GET", "/api/v1/ping", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("got %v", err)
	}
	if n < 2 {
		t.Fatalf("expected retries, n=%d", n)
	}
}

func TestDo_MalformedJSONPrices(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{not-json`)
	})
	_, err := c.GetPrices(context.Background(), "GOLD", ResolutionHour, 1, time.Time{}, time.Time{})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestDo_ContextCancel(t *testing.T) {
	started := make(chan struct{})
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-started
		cancel()
	}()
	_, err := c.Do(ctx, "GET", "/api/v1/ping", nil, nil)
	if err == nil {
		t.Fatal("expected cancel")
	}
}

func TestDo_Timeout(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	})
	c.HTTPClient.Timeout = 20 * time.Millisecond
	_, err := c.Do(context.Background(), "GET", "/api/v1/ping", nil, nil)
	if err == nil {
		t.Fatal("expected timeout")
	}
}

func TestSpecFromDetails_MissingRulesIncomplete(t *testing.T) {
	spec, err := SpecFromDetails(&MarketDetailsResponse{Instrument: MarketInstrument{Epic: "GOLD"}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if spec.SizingComplete() {
		t.Fatal("missing dealing rules and value_per_point must be incomplete")
	}
}

func TestConfirmDeal(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/confirms/R1" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"dealReference": "R1", "dealId": "D1", "status": "OPEN", "dealStatus": "ACCEPTED", "level": 2000.0, "size": 0.1,
		})
	})
	cr, err := c.ConfirmDeal(context.Background(), "R1")
	if err != nil || cr.DealID != "D1" {
		t.Fatalf("%v %#v", err, cr)
	}
}

func TestClosePosition_EmptyDealID(t *testing.T) {
	c := NewClient("http://127.0.0.1")
	if _, err := c.ClosePosition(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}
