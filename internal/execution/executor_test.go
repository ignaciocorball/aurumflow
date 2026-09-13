package execution

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aurumflow/config"
	"aurumflow/internal/market"
	"aurumflow/pkg/models"
)

func testClient(t *testing.T, h http.HandlerFunc) *market.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := market.NewClient(srv.URL)
	c.HTTPClient = srv.Client()
	return c
}

func TestExecutor_OpenConfirmCloseUpdate(t *testing.T) {
	var methods []string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/positions":
			_ = json.NewEncoder(w).Encode(map[string]any{"dealReference": "R1"})
		case strings.HasPrefix(r.URL.Path, "/api/v1/confirms/"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"dealReference": "R1", "dealId": "D1", "status": "OPEN", "dealStatus": "ACCEPTED", "level": 2000.0, "size": 0.1, "direction": "BUY", "epic": "GOLD",
			})
		case r.Method == http.MethodDelete:
			_ = json.NewEncoder(w).Encode(map[string]any{"dealReference": "R2"})
		case r.Method == http.MethodPut:
			_ = json.NewEncoder(w).Encode(map[string]any{"dealReference": "R3"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/positions":
			_ = json.NewEncoder(w).Encode(map[string]any{"positions": []any{}})
		default:
			w.WriteHeader(404)
		}
	})
	ex := NewExecutor(c, "GOLD", 0.1, 0.1)
	sig := &models.TradeSignal{Direction: "BUY", Entry: 2000, StopLoss: 1990, TakeProfit: 2020, Score: 7}
	open, err := ex.OpenPosition(context.Background(), OpenRequest{Epic: "GOLD", Direction: "BUY", Size: 0.1, StopLoss: 1990, TakeProfit: 2020, Signal: sig})
	if err != nil || open.DealReference != "R1" {
		t.Fatalf("open %v %#v", err, open)
	}
	conf, err := ex.Confirm(context.Background(), "R1")
	if err != nil || conf.DealID != "D1" {
		t.Fatalf("confirm %v %#v", err, conf)
	}
	sl := 1995.0
	if _, err := ex.UpdatePosition(context.Background(), UpdateRequest{DealID: "D1", StopLevel: &sl}); err != nil {
		t.Fatal(err)
	}
	cl, err := ex.ClosePosition(context.Background(), "D1")
	if err != nil || cl.DealReference == "" {
		t.Fatalf("close %v %#v", err, cl)
	}
	for _, m := range methods {
		if strings.Contains(m, "api-capital.backend-capital.com") && !strings.Contains(m, "demo") {
			t.Fatalf("live URL in request %s", m)
		}
	}
}

func TestDryRun_NeverMutates(t *testing.T) {
	hit := false
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete {
			hit = true
		}
		_, _ = io.WriteString(w, `{"positions":[]}`)
	})
	inner := NewExecutor(c, "GOLD", 0.1, 0.1)
	p := &DryRunProvider{Inner: inner}
	if p.Mode() != config.ExecutionDryRun {
		t.Fatal(p.Mode())
	}
	_, err := p.OpenPosition(context.Background(), OpenRequest{Epic: "GOLD", Direction: "BUY", Size: 0.1, StopLoss: 1, TakeProfit: 2})
	if err == nil || !strings.Contains(err.Error(), "would_have_sent") {
		t.Fatalf("dry run open: %v", err)
	}
	if _, err := p.ClosePosition(context.Background(), "D1"); err == nil {
		t.Fatal("close must fail")
	}
	if _, err := p.UpdatePosition(context.Background(), UpdateRequest{DealID: "D1"}); err == nil {
		t.Fatal("update must fail")
	}
	if hit {
		t.Fatal("DRY_RUN issued a mutating HTTP method")
	}
}

func TestDisabledHelper(t *testing.T) {
	if config.ExecutionDisabled != "DISABLED" {
		t.Fatal("const")
	}
}
