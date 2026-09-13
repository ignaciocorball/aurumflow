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
	mu         sync.RWMutex
	status     Status
	http       *http.Server
	addr       string
	series     []SeriesPoint
	timeline   []TimelineEvent
	lastSample time.Time
	prev       Status
	world      []byte
	worldAt    time.Time
	sources    []byte
	absShown   string
	absPend    string
	absPendAt  time.Time
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
	mux.HandleFunc("/api/timeseries", s.apiTimeseries)
	mux.HandleFunc("/api/events", s.apiEvents)
	mux.HandleFunc("/api/world", s.apiWorld)
	mux.HandleFunc("/api/regions", s.apiWorldSlice("Regions"))
	mux.HandleFunc("/api/assets", s.apiWorldSlice("AssetClasses"))
	mux.HandleFunc("/api/opportunities", s.apiWorldSlice("Opportunity"))
	mux.HandleFunc("/api/context/sources", s.apiSources)
	mux.HandleFunc("/api/institutional", s.apiInstitutional)
	s.http = &http.Server{Addr: addr, Handler: rejectMutations(mux), ReadHeaderTimeout: 5 * time.Second}
	return s
}

func (s *Server) Set(st Status) {
	s.mu.Lock()
	st.GoldQuotesOK = GoldQuotesOK(st.MarketStatus, st.GoldBid, st.GoldAsk)
	st.PositionOpen = st.PositionsKnown && st.OpenPositions > 0
	st.DecisionWhy = DecisionWhy(st)
	s.appendSeriesLocked(st, time.Now().UTC())
	if n := len(s.timeline); n > 0 {
		st.LastEvent = s.timeline[n-1].Text
	}
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
	out := map[string]any{
		"market_status": GoldMarketLabel(st.MarketStatus),
		"quotes_ok":     st.GoldQuotesOK,
		"validation":    emptyNA(st.GoldValidation),
		"session":       emptyNA(st.GoldSession),
		"legacy":        emptyNA(st.LastStrategy),
		"kill_switch":   st.KillSwitch,
		"execution_started": st.ExecutionMode == "DEMO",
		"strategy_session": st.StrategySession,
		"strategy_ready":   st.StrategyReady,
		"strategy_waiting": st.StrategyWaiting,
		"next_session":     st.NextSession,
		"last_scan":        st.LastScan,
		"last_signal":      st.LastSignal,
		"trade_count":      st.TradeCount,
		"max_trades":       st.MaxTrades,
		"monetary_status":  st.MonetaryStatus,
		"operational_trust": st.OperationalTrust,
	}
	if st.GoldQuotesOK {
		out["bid"], out["ask"], out["spread"] = st.GoldBid, st.GoldAsk, st.GoldSpread
	}
	if st.PositionOpen {
		out["open_positions"] = st.OpenPositions
		out["entry"], out["sl"], out["tp"], out["upnl"] = st.GoldEntry, st.GoldSL, st.GoldTP, st.GoldUPnL
	}
	if st.DemoBalanceOK {
		out["demo_balance"] = st.DemoBalance
	}
	writeJSON(w, out)
}

func emptyNA(v string) string {
	if v == "" {
		return "WAITING"
	}
	return v
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
		"spread": st.L2Spread, "mid": st.L2Mid, "quotes_ok": st.L2QuotesOK,
		"basis": st.Basis, "basis_z": st.BasisZ,
		"top_bids": st.TopBids, "top_asks": st.TopAsks,
	})
}

func (s *Server) apiTimeseries(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{"points": s.Series(), "bound": MaxSeriesPoints, "horizon_s": int(SeriesHorizon.Seconds())})
}

func (s *Server) apiEvents(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{"events": s.Timeline(), "bound": MaxTimeline})
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
