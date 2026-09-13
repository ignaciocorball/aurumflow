package ops

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	mu     sync.RWMutex
	status Status
	http   *http.Server
	addr   string
}

func NewServer(addr string, initial Status) *Server {
	if addr == "" {
		addr = "127.0.0.1:8765"
	}
	s := &Server{status: initial, addr: addr}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.console)
	mux.HandleFunc("/healthz", s.healthz)
	mux.HandleFunc("/status", s.statusHandler)
	mux.HandleFunc("/api/status", s.statusHandler)
	mux.HandleFunc("/api/gold", s.apiGold)
	mux.HandleFunc("/api/intelligence", s.apiIntel)
	mux.HandleFunc("/api/book", s.apiBook)
	mux.HandleFunc("/api/prospective", s.apiProspective)
	mux.HandleFunc("/api/research", s.apiResearch)
	mux.HandleFunc("/api/stream", s.apiStream)
	s.http = &http.Server{Addr: addr, Handler: rejectMutations(mux), ReadHeaderTimeout: 5 * time.Second}
	return s
}

func (s *Server) Set(st Status) {
	s.mu.Lock()
	s.status = st
	s.mu.Unlock()
}

func (s *Server) Get() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) statusHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.Get())
}

func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.addr = ln.Addr().String()
	s.http.Addr = s.addr
	return s.http.Serve(ln)
}

func (s *Server) Addr() string { return s.addr }

func rejectMutations(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "read-only console", http.StatusMethodNotAllowed)
		}
	})
}

func (s *Server) console(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(consoleHTML))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) apiGold(w http.ResponseWriter, _ *http.Request) {
	st := s.Get()
	writeJSON(w, map[string]any{
		"market_status": st.MarketStatus, "bid": st.GoldBid, "ask": st.GoldAsk, "spread": st.GoldSpread,
		"validation": st.GoldValidation, "session": st.GoldSession, "open_positions": st.OpenPositions,
		"entry": st.GoldEntry, "sl": st.GoldSL, "tp": st.GoldTP, "upnl": st.GoldUPnL,
		"daily_pnl": st.DailyPnL, "daily_dd_pct": st.DailyDDPct, "trades_today": st.TradesToday,
		"legacy": st.LastStrategy, "kill_switch": st.KillSwitch,
	})
}

func (s *Server) apiIntel(w http.ResponseWriter, _ *http.Request) {
	st := s.Get()
	writeJSON(w, map[string]any{
		"price": st.BTCPrice, "cvd": st.CVD, "pressure": st.Pressure, "directional_pressure": st.DirectionalP,
		"v1": st.LastV1Class, "flow_efficiency": st.FlowEfficiency, "impact_failure": st.ImpactFailure,
		"vel_1s": st.Vel1s, "vel_5s": st.Vel5s, "vel_30s": st.Vel30s,
		"agg_buy": st.AggBuy, "agg_sell": st.AggSell, "provider": st.V1FlowProvider,
	})
}

func (s *Server) apiBook(w http.ResponseWriter, _ *http.Request) {
	st := s.Get()
	writeJSON(w, map[string]any{
		"provider": st.L2Provider, "instrument": st.L2Instrument, "relation": st.L2Relation,
		"proxy_quality": st.L2ProxyQuality, "synced": st.BookSynced, "age_ms": st.BookAgeMs,
		"microprice": st.Microprice, "imb1": st.Imb1, "imb5": st.Imb5, "imb10": st.Imb10, "imb20": st.Imb20,
		"bid_replenishment": st.BidRepl, "ask_replenishment": st.AskRepl,
		"bid_depletion": st.BidDepl, "ask_depletion": st.AskDepl,
		"bid_persistence": st.BidPersist, "ask_persistence": st.AskPersist,
		"basis": st.Basis, "basis_z": st.BasisZ,
	})
}

func (s *Server) apiProspective(w http.ResponseWriter, _ *http.Request) {
	st := s.Get()
	writeJSON(w, map[string]any{
		"legacy_total": st.ProspectiveTotal, "exhaustion": st.ProspectiveExh,
		"continuation": st.ProspectiveCont, "neutral": st.ProspectiveNeu,
		"mature_15m": st.Mature15m, "mature_1h": st.Mature1h,
		"exh_l2_unavailable": st.ExhL2Unavailable, "exh_supportive": st.ExhSupportive,
	})
}

func (s *Server) apiResearch(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"spec": "FLOW_EXHAUSTION_V1",
		"status": "VALIDATED_EXTERNAL_HOLDOUT",
		"discovery_n": 56, "holdout_n": 50,
		"holdout_mean_bp": 9.8, "holdout_hit": 0.66, "mfe_mae": 1.99,
		"live_execution": "DISABLED",
		"note": "historical event study, not account return",
	})
}

func (s *Server) apiStream(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "no stream", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-tick.C:
			raw, _ := json.Marshal(s.Get())
			_, _ = w.Write([]byte("data: "))
			_, _ = w.Write(raw)
			_, _ = w.Write([]byte("\n\n"))
			fl.Flush()
		}
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
