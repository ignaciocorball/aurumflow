package terminal

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"aurumflow/internal/disloc"
	"aurumflow/internal/terminal/econ"
	"aurumflow/internal/terminal/news"
)

type Endpoints struct {
	Gold  string
	Intel string
	US100 string
	Root  string
}

type Hub struct {
	mu          sync.RWMutex
	client      *ReadClient
	ep          Endpoints
	gold        map[string]any
	intel       map[string]any
	us100       map[string]any
	world       map[string]any
	book        map[string]any
	series      []MicroPt
	timeline    []ActivityItem
	research    map[string]any
	goldAt      time.Time
	intelAt     time.Time
	us100At     time.Time
	worldAt     time.Time
	bookAt      time.Time
	newsAt      time.Time
	econAt      time.Time
	goldErr     string
	intelErr    string
	us100Err    string
	worldErr    string
	sparks      map[string][]float64
	trades      []ClosedTrade
	newsSvc     *news.Service
	econPath    string
	econEvents  []econ.Event
	fixture     bool
	snap        Snapshot
	subs        map[chan StreamEvent]struct{}
	lastMarket  time.Time
}

func NewHub(client *ReadClient, ep Endpoints, newsSvc *news.Service, econPath string, fixture bool) *Hub {
	return &Hub{
		client: client, ep: ep, newsSvc: newsSvc, econPath: econPath, fixture: fixture,
		sparks: map[string][]float64{}, subs: map[chan StreamEvent]struct{}{},
	}
}

func (h *Hub) Snapshot() Snapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.snap
}

func (h *Hub) Subscribe() chan StreamEvent {
	ch := make(chan StreamEvent, 32)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan StreamEvent) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *Hub) emit(ev StreamEvent) {
	h.mu.RLock()
	subs := make([]chan StreamEvent, 0, len(h.subs))
	for ch := range h.subs {
		subs = append(subs, ch)
	}
	h.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

func (h *Hub) Run(stop <-chan struct{}) {
	h.refreshAll(time.Now().UTC())
	statusT := time.NewTicker(time.Second)
	worldT := time.NewTicker(15 * time.Second)
	bookT := time.NewTicker(2 * time.Second)
	newsT := time.NewTicker(30 * time.Second)
	econT := time.NewTicker(time.Minute)
	tradeT := time.NewTicker(30 * time.Second)
	defer statusT.Stop()
	defer worldT.Stop()
	defer bookT.Stop()
	defer newsT.Stop()
	defer econT.Stop()
	defer tradeT.Stop()
	for {
		select {
		case <-stop:
			return
		case <-statusT.C:
			h.pollStatus()
			h.rebuild(time.Now().UTC())
		case <-worldT.C:
			h.pollWorld()
			h.rebuild(time.Now().UTC())
		case <-bookT.C:
			h.pollBook()
			h.rebuild(time.Now().UTC())
		case <-newsT.C:
			if h.newsSvc != nil && h.newsSvc.Tick(time.Now().UTC()) {
				h.emit(StreamEvent{Type: "NEWS_CLUSTER_UPDATED", At: nowUTC()})
			}
			h.rebuild(time.Now().UTC())
		case <-econT.C:
			h.pollEcon(time.Now().UTC())
			h.rebuild(time.Now().UTC())
		case <-tradeT.C:
			tr := LoadTrades(h.ep.Root)
			h.mu.Lock()
			h.trades = tr
			h.mu.Unlock()
			h.rebuild(time.Now().UTC())
		}
	}
}

func (h *Hub) refreshAll(now time.Time) {
	h.pollStatus()
	h.pollWorld()
	h.pollBook()
	h.pollResearch()
	h.pollEcon(now)
	tr := LoadTrades(h.ep.Root)
	h.mu.Lock()
	h.trades = tr
	h.mu.Unlock()
	h.rebuild(now)
	if h.newsSvc != nil {
		go func() {
			h.newsSvc.Tick(time.Now().UTC())
			h.rebuild(time.Now().UTC())
		}()
	}
}

func (h *Hub) pollStatus() {
	h.getMap(h.ep.Gold, "/status", &h.gold, &h.goldAt, &h.goldErr)
	h.getMap(h.ep.Intel, "/status", &h.intel, &h.intelAt, &h.intelErr)
	h.getMap(h.ep.US100, "/status", &h.us100, &h.us100At, &h.us100Err)
}

func (h *Hub) pollWorld() {
	h.getMap(h.ep.Intel, "/api/world", &h.world, &h.worldAt, &h.worldErr)
}

func (h *Hub) pollBook() {
	var dummy time.Time
	var err string
	h.getMap(h.ep.Intel, "/api/book", &h.book, &dummy, &err)
	if err == "" {
		h.mu.Lock()
		h.bookAt = time.Now().UTC()
		h.mu.Unlock()
	}
	var ts map[string]any
	if e := h.client.GetJSON(joinURL(h.ep.Intel, "/api/timeseries"), &ts); e == nil {
		h.mu.Lock()
		h.series = parseSeries(ts)
		h.mu.Unlock()
	}
	var ev map[string]any
	if e := h.client.GetJSON(joinURL(h.ep.Intel, "/api/events"), &ev); e == nil {
		h.mu.Lock()
		h.timeline = parseTimeline(ev, "intel")
		h.mu.Unlock()
	}
}

func (h *Hub) pollResearch() {
	_ = h.client.GetJSON(joinURL(h.ep.Intel, "/api/research"), &h.research)
}

func (h *Hub) pollEcon(now time.Time) {
	ev := econ.LoadSchedule(h.econPath, now)
	h.mu.Lock()
	h.econEvents = ev
	h.econAt = now
	h.mu.Unlock()
}

func (h *Hub) getMap(base, path string, dest *map[string]any, at *time.Time, errStr *string) {
	var m map[string]any
	err := h.client.GetJSON(joinURL(base, path), &m)
	h.mu.Lock()
	defer h.mu.Unlock()
	if err != nil {
		*errStr = err.Error()
		return
	}
	*dest = m
	*at = time.Now().UTC()
	*errStr = ""
}

func (h *Hub) rebuild(now time.Time) {
	h.mu.Lock()
	prev := h.snap
	snap := h.buildLocked(now)
	h.snap = snap
	h.mu.Unlock()
	h.diffEmit(prev, snap, now)
}

func (h *Hub) buildLocked(now time.Time) Snapshot {
	markets := h.mergeMarkets()
	h.updateSparks(markets)
	for i := range markets {
		markets[i].Spark = append([]float64(nil), h.sparks[markets[i].ID]...)
	}
	dis := classifyDislocation(markets)
	world := worldView(h.world, dis)
	gold := h.gold
	us := h.us100
	intel := h.intel
	equity := asFloat(gold["demo_balance"])
	if equity == 0 {
		equity = asFloat(us["demo_balance"])
	}
	port := AggregatePnL(h.trades, "BROKER_DEMO", "ALL", now, equity)
	port.Unrealized = asFloat(gold["gold_upnl"]) + asFloat(us["us100_upl"])
	port.Net = port.Realized + port.Unrealized
	port.OpenRisk = asFloat(gold["expected_risk"]) + asFloat(us["us100_risk_used"])
	port.DrawdownPct = firstFloat(gold, "daily_dd_pct")
	port.PilotCap = firstFloat(gold, "pilot_per_position_cap")
	if port.PilotCap == 0 {
		port.PilotCap = 5
	}
	port.AggregateCap = firstFloat(gold, "pilot_aggregate_cap")
	if port.AggregateCap == 0 {
		port.AggregateCap = 10
	}
	port.OpenRiskCap = port.AggregateCap
	port.PreciousRisk = asFloat(gold["expected_risk"])
	port.PreciousCap = 5
	port.USEquityRisk = asFloat(us["us100_risk_used"])
	port.USEquityCap = 5
	port.TodayPnL += port.Unrealized
	armed := 0
	if asBool(gold["strategy_ready"]) {
		armed++
	}
	if strings.EqualFold(asString(us["us100_mirror_armed"]), "ON") || strings.Contains(strings.ToUpper(asString(us["us100_demo_eligible"])), "YES") || strings.Contains(strings.ToUpper(asString(us["us100_status"])), "ELIGIBLE") {
		armed++
	}
	port.PilotsArmed = armed

	exec := buildExecution(gold, us)
	health := h.buildHealth(now)
	overall := overallHealth(health)
	env := Environment{
		Name: "DEMO", Demo: true,
		LivePossible: asString(gold["live_possible"]),
		LivePossibleDisplay: Display(asString(gold["live_possible"])),
		AccountMasked: firstString(gold, "account_masked"),
		BrokerStatus: brokerStatus(gold, us),
		WorldValid: world.Valid, WorldValidDisplay: world.ValidDisplay,
		DataHealth: overall, UTC: now.Format("15:04:05") + " UTC",
	}
	s := Snapshot{
		GeneratedAt: now.Format(time.RFC3339), FixtureMode: h.fixture,
		Environment: env, World: world, Markets: markets, Portfolio: port,
		Execution: exec, Research: h.buildResearch(markets, gold, us),
		Health: health, Micro: h.buildMicro(),
		Regions: buildRegions(markets, h.econEvents, now),
		Flows: buildFlows(h.world),
		Activity: h.timeline,
	}
	if h.newsSvc != nil {
		for _, c := range h.newsSvc.Clusters(now) {
			if len(s.News) >= 12 {
				break
			}
			s.News = append(s.News, NewsCluster{
				ID: c.ID, Title: c.Title, Summary: c.Summary, Source: c.Source, Sources: c.Sources,
				URL: c.URL, Published: c.Published.UTC().Format(time.RFC3339),
				Age: ageLabel(now.Sub(c.Published)), Markets: c.Markets, Themes: c.Themes,
				Relevance: c.Relevance, Importance: c.Importance, Count: c.Count, Region: c.Region,
			})
		}
	}
	win := econ.FilterWindow(h.econEvents, now, 36*time.Hour, 21*24*time.Hour)
	for _, e := range win {
		if len(s.Events) >= 12 {
			break
		}
		st, window, cd := econ.StatusOf(e, now)
		s.Events = append(s.Events, EconEvent{
			ID: e.ID, Name: e.Name, Region: e.Region, At: e.At.UTC().Format(time.RFC3339),
			Importance: e.Importance, Source: e.Source, SourceURL: e.SourceURL,
			AssetClasses: e.AssetClasses, Markets: e.Markets, Status: st, Window: window,
			Countdown: cd, Actual: e.Actual, Forecast: e.Forecast, Previous: e.Previous,
		})
	}
	s.Narratives = BuildNarratives(s)
	_ = intel
	return s
}

func (h *Hub) mergeMarkets() []Market {
	wm := map[string]map[string]any{}
	if raw, ok := h.world["Markets"].(map[string]any); ok {
		for k, v := range raw {
			if mm, ok := v.(map[string]any); ok {
				wm[k] = mm
			}
		}
	}
	opp := []map[string]any{}
	if arr, ok := h.world["Opportunity"].([]any); ok {
		for _, v := range arr {
			if mm, ok := v.(map[string]any); ok {
				opp = append(opp, mm)
			}
		}
	}
	ids := []string{"US100", "US500", "US30", "DE40", "UK100", "J225", "CN50", "GOLD", "SILVER", "OIL_CRUDE", "BTC"}
	out := make([]Market, 0, len(ids))
	for _, id := range ids {
		src := wm[id]
		if src == nil {
			src = map[string]any{}
		}
		for _, o := range opp {
			if strings.EqualFold(asString(o["Market"]), id) {
				src = mergeMaps(src, o)
			}
		}
		mid := asFloat(src["Mid"])
		bid := asFloat(src["Bid"])
		ask := asFloat(src["Ask"])
		if id == "GOLD" && mid == 0 {
			bid = asFloat(h.gold["gold_bid"])
			ask = asFloat(h.gold["gold_ask"])
			if bid > 0 && ask > 0 {
				mid = (bid + ask) / 2
			}
		}
		if id == "BTC" && mid == 0 {
			mid = asFloat(h.intel["btc_price"])
		}
		if id == "US100" && mid == 0 {
			mid = asFloat(h.us100["us100_current"])
		}
		chg := changePct(h.sparks[id], mid)
		m := Market{
			ID: id, Label: MarketLabel(id),
			Region: asString(src["Region"]), AssetClass: asString(src["AssetClass"]),
			Bid: bid, Ask: ask, Mid: mid, ChangePct: chg,
			Attention: asFloat(src["Attention"]), Salience: asFloat(src["Salience"]), Coverage: asFloat(src["Coverage"]),
			Trend: asString(src["PriceTrend"]), TrendDisplay: Display(asString(src["PriceTrend"])),
			Setup: asString(src["Setup"]), SetupDisplay: Display(asString(src["Setup"])),
			Session: asString(src["SessionLocal"]), SessionDisplay: Display(asString(src["SessionLocal"])),
			Quality: asString(src["DataQuality"]), QualityDisplay: Display(asString(src["DataQuality"])),
			MarketStatus: asString(src["MarketStatus"]), MarketStatusDisplay: Display(asString(src["MarketStatus"])),
			Eligibility: asString(src["Eligibility"]),
			Positioning: asString(src["Positioning"]), FlowContext: asString(src["CapitalFlowContext"]),
			SalienceReason: asString(src["SalienceReason"]),
			Research: asString(src["LegacyCompat"]), ResearchDisplay: Display(asString(src["LegacyCompat"])),
			DemoStatus: asString(src["Eligibility"]), DemoDisplay: Display(asString(src["Eligibility"])),
			RegionLabel: regionLabel(asString(src["Region"])),
		}
		if ev, ok := src["Evidence"].([]any); ok {
			for _, e := range ev {
				if s := asString(e); s != "" {
					m.Evidence = append(m.Evidence, s)
				}
			}
		}
		if id == "GOLD" {
			m.Pilot = true
			m.Research = firstString(h.gold, "gold_validation", "monetary_status")
			m.ResearchDisplay = Display(m.Research)
			m.DemoStatus = "DEMO"
			m.DemoDisplay = "DEMO"
			if m.MarketStatus == "" {
				m.MarketStatus = asString(h.gold["market_status"])
				m.MarketStatusDisplay = Display(m.MarketStatus)
			}
			m.Session = firstString(h.gold, "strategy_session")
			m.SessionDisplay = Display(m.Session)
			if wait := asString(h.gold["strategy_waiting"]); wait != "" {
				m.SetupDisplay = Display(wait)
			}
		}
		if id == "US100" {
			m.Pilot = true
			m.DemoStatus = firstString(h.us100, "us100_status", "us100_demo_eligible")
			m.DemoDisplay = Display(m.DemoStatus)
			m.Research = firstString(h.us100, "us100_origin")
			if m.Research == "" {
				m.Research = "DEMO_MIRROR"
			}
			m.ResearchDisplay = Display(m.Research)
			m.Session = firstString(h.us100, "strategy_session")
			m.SessionDisplay = Display(m.Session)
		}
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Salience > out[j].Salience })
	return out
}

func (h *Hub) updateSparks(ms []Market) {
	for _, m := range ms {
		if m.Mid == 0 {
			continue
		}
		cur := h.sparks[m.ID]
		if n := len(cur); n > 0 && cur[n-1] == m.Mid {
			continue
		}
		cur = append(cur, m.Mid)
		if len(cur) > 48 {
			cur = cur[len(cur)-48:]
		}
		h.sparks[m.ID] = cur
	}
}

func (h *Hub) buildHealth(now time.Time) []HealthItem {
	item := func(id, label string, at time.Time, err, detail string, staleAfter time.Duration) HealthItem {
		st := "HEALTHY"
		if err != "" && at.IsZero() {
			st = "OFFLINE"
			detail = err
		} else if err != "" {
			st = "DEGRADED"
			detail = err
		} else if !at.IsZero() && now.Sub(at) > staleAfter {
			st = "DEGRADED"
			detail = "Stale"
		}
		last := ""
		if !at.IsZero() {
			last = at.UTC().Format(time.RFC3339)
		}
		return HealthItem{ID: id, Label: label, Status: st, Display: Display(st), LastUpdate: last, Detail: detail}
	}
	out := []HealthItem{
		item("capital", "Capital", h.goldAt, h.goldErr, "DEMO broker", 8*time.Second),
		item("binance", "Binance", h.bookAt, "", bookDetail(h.intel), 8*time.Second),
		item("world", "WorldState", h.worldAt, h.worldErr, asString(h.world["Valid"]), 45*time.Second),
	}
	if h.newsSvc != nil {
		st, detail, last := h.newsSvc.Status()
		out = append(out, HealthItem{ID: "news", Label: "News", Status: st, Display: Display(st), LastUpdate: last.Format(time.RFC3339), Detail: detail})
	} else {
		out = append(out, HealthItem{ID: "news", Label: "News", Status: "OFFLINE", Display: Display("OFFLINE")})
	}
	econSt := "HEALTHY"
	if len(h.econEvents) == 0 {
		econSt = "DEGRADED"
	}
	out = append(out,
		HealthItem{ID: "econ", Label: "Economic events", Status: econSt, Display: Display(econSt), LastUpdate: h.econAt.Format(time.RFC3339), Detail: "Official schedules"},
		HealthItem{ID: "research", Label: "Research recorder", Status: healthFrom(h.intelAt, h.intelErr, now, 8*time.Second), Display: Display(healthFrom(h.intelAt, h.intelErr, now, 8*time.Second)), LastUpdate: h.intelAt.Format(time.RFC3339)},
		item("broker", "Broker execution", h.goldAt, h.goldErr, firstString(h.gold, "account_verified"), 8*time.Second),
	)
	if !asBool(h.intel["book_synced"]) && h.intelErr == "" && !h.intelAt.IsZero() {
		for i := range out {
			if out[i].ID == "binance" {
				out[i].Status = "DEGRADED"
				out[i].Display = "Book resync"
				out[i].Detail = firstString(h.intel, "resync_reason")
			}
		}
	}
	return out
}

func (h *Hub) buildMicro() MicroView {
	st := h.intel
	bids, asks := bookSides(h.book)
	return MicroView{
		Price: asFloat(st["btc_price"]), Microprice: asFloat(st["microprice"]),
		Spread: asFloat(st["l2_spread"]), CVD: asFloat(st["cvd"]), Pressure: asFloat(st["pressure"]),
		Imbalance: asFloat(st["imbalance"]), Imb1: asFloat(st["imbalance_1"]), Imb5: asFloat(st["imbalance_5"]), Imb10: asFloat(st["imbalance_10"]),
		BookSynced: asBool(st["book_synced"]), BookAgeMs: asInt64(st["book_age_ms"]), Gaps: asInt(st["book_gaps"]),
		Resyncs: asInt(st["resyncs"]), Drops: asInt64(st["dropped_events"]),
		LatencyP50: asFloat(st["latency_p50_ms"]), LatencyP95: asFloat(st["latency_p95_ms"]), LatencyP99: asFloat(st["latency_p99_ms"]),
		Provider: firstString(st, "l2_provider", "v1_flow_provider"), Capability: asString(st["book_capability"]),
		Absorption: firstString(st, "absorption_status"), Exhaustion: firstString(st, "last_v1_classification"),
		Integrity: firstString(st, "integrity_status"), EventRate: asFloat(st["event_rate"]),
		Series: capSeries(h.series, 240), Bids: bids, Asks: asks, ResyncReason: asString(st["resync_reason"]),
	}
}

func (h *Hub) buildResearch(ms []Market, gold, us map[string]any) ResearchView {
	row := func(market, status string, extra ResearchRow) ResearchRow {
		extra.Market = market
		extra.Status = status
		extra.StatusDisplay = Display(status)
		return extra
	}
	hold := ResearchRow{}
	if h.research != nil {
		hold.Spec = asString(h.research["spec"])
		hold.HoldoutN = asInt(h.research["holdout_n"])
		hold.DiscoveryN = asInt(h.research["discovery_n"])
		hold.Expectancy = asFloat(h.research["holdout_mean_bp"])
		hold.ProfitFactor = asFloat(h.research["mfe_mae"])
		hold.Hit = asFloat(h.research["holdout_hit"])
		hold.Note = asString(h.research["note"])
		hold.Dataset = "FLOW_EXHAUSTION_V1 holdout"
		hold.Status = asString(h.research["status"])
	}
	rows := []ResearchRow{
		row("GOLD", researchStatus(gold, "GOLD"), ResearchRow{Spec: firstString(gold, "strategy_name"), SpecHash: asString(gold["spec_hash"]), Note: "Legacy DEMO pilot"}),
		row("US100", researchStatus(us, "US100"), ResearchRow{Spec: firstString(us, "strategy_name"), SpecHash: asString(us["spec_hash"]), Note: "DEMO_MIRROR of LEGACY_NORMALIZED_V0"}),
		row("BTC", firstNonEmpty(hold.Status, "SHADOW_VALIDATED"), hold),
	}
	for _, m := range ms {
		if m.ID == "GOLD" || m.ID == "US100" || m.ID == "BTC" {
			continue
		}
		st := "DATA_READY"
		if m.Quality == "UNKNOWN" || m.Quality == "" {
			st = "BLOCKED"
		} else if m.Eligibility == "ANALYSIS_ONLY" {
			st = "RESEARCH_READY"
		}
		rows = append(rows, row(m.ID, st, ResearchRow{Note: "Not a DEMO pilot"}))
	}
	return ResearchView{Rows: rows}
}

func (h *Hub) diffEmit(prev, cur Snapshot, now time.Time) {
	at := now.UTC().Format(time.RFC3339)
	if prev.World.Hash != cur.World.Hash && cur.World.Hash != "" {
		h.emit(StreamEvent{Type: "WORLD_UPDATED", At: at})
	}
	if prev.World.Dislocation != cur.World.Dislocation && cur.World.Dislocation != "" {
		h.emit(StreamEvent{Type: "DISLOCATION_CHANGED", At: at, Data: cur.World.Dislocation})
	}
	if overallHealth(prev.Health) != overallHealth(cur.Health) {
		h.emit(StreamEvent{Type: "DATA_HEALTH_CHANGED", At: at, Data: overallHealth(cur.Health)})
	}
	if executionChanged(prev.Execution, cur.Execution) {
		h.emit(StreamEvent{Type: "POSITION_UPDATED", At: at})
	}
	if prev.Portfolio.TradeCount != cur.Portfolio.TradeCount {
		h.emit(StreamEvent{Type: "TRADE_COMPLETED", At: at})
		h.emit(StreamEvent{Type: "PNL_UPDATED", At: at})
	}
	if now.Sub(h.lastMarket) >= 500*time.Millisecond {
		h.lastMarket = now
		h.emit(StreamEvent{Type: "MARKET_UPDATED", At: at})
	}
}

func executionChanged(a, b ExecutionView) bool {
	if len(a.Positions) != len(b.Positions) {
		return true
	}
	for i := range a.Positions {
		if a.Positions[i].Open != b.Positions[i].Open || a.Positions[i].UPL != b.Positions[i].UPL || a.Positions[i].LastExecution != b.Positions[i].LastExecution {
			return true
		}
	}
	return false
}

func buildExecution(gold, us map[string]any) ExecutionView {
	gOpen := asBool(gold["position_open"]) || asInt(gold["open_positions"]) > 0
	uOpen := asInt(us["open_positions"]) > 0 || asFloat(us["us100_entry"]) > 0 && asString(us["us100_direction"]) != ""
	gRisk := asFloat(gold["expected_risk"])
	uRisk := asFloat(us["us100_risk_used"])
	if uRisk == 0 {
		uRisk = asFloat(us["us100_risk"])
	}
	g := Position{
		Market: "GOLD", Origin: "GOLD_STRATEGY", OriginDisplay: Display("GOLD_STRATEGY"), Kind: "DEMO",
		Strategy: firstString(gold, "strategy_name"), Direction: dirFrom(gold),
		Entry: asFloat(gold["gold_entry"]), Current: midOf(asFloat(gold["gold_bid"]), asFloat(gold["gold_ask"])),
		SL: asFloat(gold["gold_sl"]), TP: asFloat(gold["gold_tp"]), UPL: asFloat(gold["gold_upnl"]),
		PlannedRisk: gRisk, ActualRisk: gRisk, Account: firstString(gold, "account_masked"),
		Open: gOpen, Session: firstString(gold, "strategy_session"), Waiting: firstString(gold, "strategy_waiting"),
		LastExecution: firstString(gold, "last_execution"), Size: 0,
		Eligible: firstString(gold, "operational_trust"), Armed: boolArmed(asBool(gold["strategy_ready"])),
		BrokerReconcile: reconcile(asBool(gold["positions_known"]), asInt(gold["open_positions"]), asInt(gold["open_positions"])),
	}
	if g.Current == 0 {
		g.Current = g.Entry
	}
	g.ReturnR = rMultiple(g.UPL, g.PlannedRisk)
	g.Lifecycle = lifecycleGold(gold, gOpen)
	g.LifecycleLabel = activeLife(g.Lifecycle)

	u := Position{
		Market: "US100", Origin: firstString(us, "us100_origin"), OriginDisplay: Display(firstString(us, "us100_origin")),
		Kind: "DEMO", Strategy: firstString(us, "strategy_name"), Direction: asString(us["us100_direction"]),
		Entry: asFloat(us["us100_entry"]), Current: asFloat(us["us100_current"]),
		SL: asFloat(us["us100_sl"]), TP: asFloat(us["us100_tp"]), UPL: asFloat(us["us100_upl"]),
		PlannedRisk: uRisk, ActualRisk: uRisk, Account: firstString(us, "account_masked"),
		Open: uOpen, Session: firstString(us, "strategy_session"), Waiting: firstString(us, "strategy_waiting"),
		LastExecution: firstString(us, "last_execution"), DealRef: firstString(us, "us100_deal_reference"),
		Size: asFloat(us["us100_size"]), Eligible: firstString(us, "us100_demo_eligible"),
		Armed: firstString(us, "us100_mirror_armed"),
		BrokerReconcile: reconcile(asBool(us["positions_known"]), asInt(us["open_positions"]), asInt(us["open_positions"])),
	}
	if u.Origin == "" {
		u.Origin = "DEMO_MIRROR"
		u.OriginDisplay = Display(u.Origin)
	}
	if strings.Contains(strings.ToUpper(firstString(us, "us100_shadow")), "SHADOW") && !uOpen {
		u.Kind = "SHADOW"
	}
	u.ReturnR = rMultiple(u.UPL, u.PlannedRisk)
	u.Lifecycle = lifecycleUS100(us, uOpen)
	u.LifecycleLabel = activeLife(u.Lifecycle)

	return ExecutionView{
		Banner: "CAPITAL DEMO — broker-backed test execution. Live trading is impossible.",
		Positions: []Position{g, u},
		Halt: asBool(gold["halt_new_orders"]) || asBool(gold["kill_switch"]),
		RiskGroups: []RiskBar{
			{ID: "open", Label: "Open risk", Used: gRisk + uRisk, Cap: firstFloat(gold, "pilot_aggregate_cap")},
			{ID: "precious", Label: "Precious", Used: gRisk, Cap: 5},
			{ID: "us_equity", Label: "US equity", Used: uRisk, Cap: 5},
		},
		Trust: []TrustItem{
			{Market: "GOLD", Code: firstString(gold, "operational_trust"), Display: Display(firstString(gold, "operational_trust")), Hint: TrustHint(firstString(gold, "operational_trust"))},
			{Market: "US100", Code: firstString(us, "us100_status"), Display: Display(firstString(us, "us100_status")), Hint: TrustHint(firstString(us, "us100_status"))},
		},
	}
}

func lifecycleGold(st map[string]any, open bool) []LifeStage {
	sig := "pending"
	if firstString(st, "last_signal") != "" || strings.Contains(strings.ToUpper(firstString(st, "last_execution")), "REJ") {
		sig = "done"
	}
	risk := stageBool(asBool(st["risk_ready"]))
	order := "pending"
	if firstString(st, "last_execution") != "" {
		order = "done"
	}
	broker := stageBool(asBool(st["broker_ready"]) || asBool(st["positions_known"]))
	prot, mon, exit, rec := "pending", "pending", "pending", "pending"
	if open {
		prot = "done"
		if asFloat(st["gold_sl"]) > 0 && asFloat(st["gold_tp"]) > 0 {
			prot = "done"
		}
		mon = "active"
	} else {
		if asInt(st["trade_count"]) > 0 {
			exit = "done"
			rec = "done"
		}
	}
	return stages(sig, risk, order, broker, prot, mon, exit, rec)
}

func lifecycleUS100(st map[string]any, open bool) []LifeStage {
	sig := "pending"
	if firstString(st, "last_signal") != "" {
		sig = "done"
	}
	risk := stageBool(asBool(st["risk_ready"]))
	order := "pending"
	if firstString(st, "last_execution") != "" {
		order = "done"
	}
	broker := stageBool(asBool(st["broker_ready"]) || asBool(st["positions_known"]))
	prot, mon, exit, rec := "pending", "pending", "pending", "pending"
	if open {
		if asFloat(st["us100_sl"]) > 0 {
			prot = "done"
		}
		mon = "active"
	}
	return stages(sig, risk, order, broker, prot, mon, exit, rec)
}

func stages(sig, risk, order, broker, prot, mon, exit, rec string) []LifeStage {
	return []LifeStage{
		{ID: "signal", Label: "Signal", State: sig},
		{ID: "risk", Label: "Risk", State: risk},
		{ID: "order", Label: "Order", State: order},
		{ID: "broker", Label: "Broker", State: broker},
		{ID: "protection", Label: "Protection", State: prot},
		{ID: "monitor", Label: "Monitor", State: mon},
		{ID: "exit", Label: "Exit", State: exit},
		{ID: "reconcile", Label: "Reconcile", State: rec},
	}
}

func stageBool(ok bool) string {
	if ok {
		return "done"
	}
	return "pending"
}

func activeLife(ss []LifeStage) string {
	for i := len(ss) - 1; i >= 0; i-- {
		if ss[i].State == "active" {
			return ss[i].Label
		}
	}
	for i := len(ss) - 1; i >= 0; i-- {
		if ss[i].State == "done" {
			return ss[i].Label
		}
	}
	return "Idle"
}

func worldView(w map[string]any, dis string) WorldView {
	valid := asString(w["Valid"])
	risk := asString(w["Risk"])
	liq := ""
	if lm, ok := w["Liquidity"].(map[string]any); ok {
		liq = asString(lm["Class"])
	}
	usd := ""
	if um, ok := w["USD"].(map[string]any); ok {
		usd = asString(um["USD"])
	}
	asof := asString(w["AsOf"])
	return WorldView{
		AsOf: asof, Valid: valid, ValidDisplay: Display(valid), Hash: asString(w["Hash"]),
		Session: asString(w["Session"]), Regime: risk, RegimeDisplay: Display(risk),
		Dislocation: dis, DislocationDisplay: Display(dis), Confidence: asFloat(w["Confidence"]),
		Liquidity: Display(liq), USD: Display(usd),
	}
}

func classifyDislocation(ms []Market) string {
	high := 0
	var sals []float64
	for _, m := range ms {
		sals = append(sals, m.Salience)
		if m.Salience >= 60 {
			high++
		}
	}
	sort.Float64s(sals)
	med := 0.0
	if n := len(sals); n > 0 {
		med = sals[n/2]
	}
	return disloc.Classify(disloc.Input{HighSalience: high, MedianSal: med})
}

func buildRegions(ms []Market, events []econ.Event, now time.Time) []RegionView {
	type spec struct{ id, label string; markets []string }
	specs := []spec{
		{"NORTH_AMERICA", "North America", []string{"US100", "US500", "US30"}},
		{"UNITED_KINGDOM", "United Kingdom", []string{"UK100"}},
		{"EUROPE", "Europe", []string{"DE40", "UK100"}},
		{"JAPAN", "Japan", []string{"J225"}},
		{"CHINA_HONG_KONG", "China / Hong Kong", []string{"CN50"}},
		{"ASIA_PACIFIC", "Asia-Pacific", []string{"J225", "CN50"}},
		{"COMMODITIES", "Global commodities", []string{"GOLD", "SILVER", "OIL_CRUDE"}},
		{"CRYPTO", "Crypto", []string{"BTC"}},
	}
	idx := map[string]Market{}
	for _, m := range ms {
		idx[m.ID] = m
	}
	var out []RegionView
	for _, sp := range specs {
		r := RegionView{ID: sp.id, Label: sp.label, Markets: sp.markets, Quality: "UNKNOWN"}
		n := 0.0
		for _, id := range sp.markets {
			m, ok := idx[id]
			if !ok {
				continue
			}
			r.Salience += m.Salience
			r.Attention += m.Attention
			r.Coverage += m.Coverage
			n++
			if m.Quality == "HEALTHY" {
				r.Quality = "HEALTHY"
			} else if r.Quality != "HEALTHY" && m.Quality != "" {
				r.Quality = m.Quality
			}
		}
		if n > 0 {
			r.Salience /= n
			r.Attention /= n
			r.Coverage /= n
		}
		r.Narrative = regionNarrative(sp.label, idx, sp.markets)
		r.NextEvent = nextEventFor(sp.markets, events, now)
		out = append(out, r)
	}
	return out
}

func regionNarrative(label string, idx map[string]Market, ids []string) string {
	var best Market
	for _, id := range ids {
		m := idx[id]
		if m.Salience >= best.Salience {
			best = m
		}
	}
	if best.ID == "" {
		return label + " has no live market frame."
	}
	return best.Label + " " + strings.ToLower(best.TrendDisplay) + ", salience " + trim1(best.Salience) + "."
}

func nextEventFor(markets []string, events []econ.Event, now time.Time) string {
	want := map[string]bool{}
	for _, m := range markets {
		want[MarketLabel(m)] = true
		want[m] = true
	}
	var best *econ.Event
	for i := range events {
		e := events[i]
		if e.At.Before(now) {
			continue
		}
		ok := false
		for _, m := range e.Markets {
			if want[m] || want[MarketLabel(m)] {
				ok = true
			}
		}
		if ok && (best == nil || e.At.Before(best.At)) {
			best = &events[i]
		}
	}
	if best == nil {
		return ""
	}
	return best.Name
}

func buildFlows(world map[string]any) []FlowView {
	arr, _ := world["Flows"].([]any)
	var out []FlowView
	for _, v := range arr {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		dir := strings.ToUpper(asString(m["Direction"]))
		if dir == "" || dir == "UNKNOWN" || dir == "MIXED" {
			continue
		}
		ev := ""
		if xs, ok := m["Evidence"].([]any); ok && len(xs) > 0 {
			ev = asString(xs[0])
		}
		if ev == "" {
			continue
		}
		class := "INFERRED"
		health := strings.ToUpper(asString(m["Health"]))
		if health == "HEALTHY" {
			class = "OBSERVED"
		}
		from := asString(m["Region"])
		to := asString(m["AssetClass"])
		out = append(out, FlowView{
			From: from, To: to, Class: class, ClassDisplay: Display(class),
			Evidence: ev, Confidence: "medium", Strength: asFloat(m["Strength"]),
		})
	}
	return out
}

func mergeMaps(a, b map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if s := asString(v); s == "" && asFloat(v) == 0 && !asBool(v) {
			if _, exists := out[k]; exists {
				continue
			}
		}
		out[k] = v
	}
	return out
}

func changePct(spark []float64, mid float64) float64 {
	if mid == 0 || len(spark) < 2 {
		return 0
	}
	base := spark[0]
	if base == 0 {
		return 0
	}
	return (mid - base) / base * 100
}

func overallHealth(items []HealthItem) string {
	st := "HEALTHY"
	for _, h := range items {
		if h.ID == "news" || h.ID == "econ" {
			continue
		}
		switch strings.ToUpper(h.Status) {
		case "OFFLINE":
			return "OFFLINE"
		case "DEGRADED", "STALE":
			st = "DEGRADED"
		}
	}
	return st
}

func brokerStatus(gold, us map[string]any) string {
	if asBool(gold["broker_ready"]) || asBool(gold["positions_known"]) {
		return "DEMO READY"
	}
	if asString(gold["account_verified"]) != "" {
		return Display(asString(gold["account_verified"]))
	}
	_ = us
	return "UNKNOWN"
}

func bookDetail(st map[string]any) string {
	if asBool(st["book_synced"]) {
		return "Book synced"
	}
	if r := asString(st["resync_reason"]); r != "" {
		return r
	}
	return "Book"
}

func healthFrom(at time.Time, err string, now time.Time, stale time.Duration) string {
	if err != "" && at.IsZero() {
		return "OFFLINE"
	}
	if err != "" || (!at.IsZero() && now.Sub(at) > stale) {
		return "DEGRADED"
	}
	if at.IsZero() {
		return "OFFLINE"
	}
	return "HEALTHY"
}

func researchStatus(st map[string]any, market string) string {
	if market == "GOLD" {
		if strings.Contains(strings.ToUpper(asString(st["operational_trust"])), "TRUST") {
			return "DEMO_OPERATIONAL"
		}
		if asBool(st["strategy_ready"]) {
			return "DEMO_ELIGIBLE"
		}
		return "SHADOW_VALIDATED"
	}
	u := strings.ToUpper(asString(st["us100_status"]))
	if strings.Contains(u, "ELIGIBLE") {
		return "DEMO_ELIGIBLE"
	}
	return "SHADOW_VALIDATED"
}

func parseSeries(ts map[string]any) []MicroPt {
	arr, _ := ts["points"].([]any)
	var out []MicroPt
	for _, v := range arr {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, MicroPt{
			T: asInt64(m["t"]), Price: asFloat(m["price"]), Micro: asFloat(m["microprice"]),
			Pressure: asFloat(m["pressure"]), CVD: asFloat(m["cvd"]),
		})
	}
	return out
}

func parseTimeline(ev map[string]any, process string) []ActivityItem {
	arr, _ := ev["events"].([]any)
	var out []ActivityItem
	for _, v := range arr {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		t := asInt64(m["t"])
		at := ""
		if t > 0 {
			at = time.UnixMilli(t).UTC().Format(time.RFC3339)
		}
		out = append(out, ActivityItem{At: at, Kind: asString(m["kind"]), Text: asString(m["text"]), Process: process})
	}
	if len(out) > 40 {
		out = out[len(out)-40:]
	}
	return out
}

func bookSides(book map[string]any) (bids, asks []BookLvl) {
	parse := func(key string) []BookLvl {
		arr, _ := book[key].([]any)
		var out []BookLvl
		for _, v := range arr {
			m, ok := v.(map[string]any)
			if !ok {
				continue
			}
			out = append(out, BookLvl{Price: asFloat(m["price"]), Qty: asFloat(m["qty"])})
			if len(out) >= 8 {
				break
			}
		}
		return out
	}
	return parse("top_bids"), parse("top_asks")
}

func capSeries(in []MicroPt, n int) []MicroPt {
	if len(in) <= n {
		return in
	}
	return in[len(in)-n:]
}

func dirFrom(gold map[string]any) string {
	d := asInt(gold["last_legacy_direction"])
	if d > 0 {
		return "BUY"
	}
	if d < 0 {
		return "SELL"
	}
	return ""
}

func midOf(bid, ask float64) float64 {
	if bid > 0 && ask > 0 {
		return (bid + ask) / 2
	}
	return bid
}

func rMultiple(upl, risk float64) float64 {
	if risk == 0 {
		return 0
	}
	return upl / risk
}

func reconcile(known bool, local, broker int) string {
	if !known {
		return "Unknown"
	}
	if local == broker {
		return "Matched"
	}
	return "Mismatch"
}

func boolArmed(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}

func regionLabel(r string) string {
	switch r {
	case "UNITED_STATES":
		return "North America"
	case "EUROPE":
		return "Europe"
	case "JAPAN":
		return "Japan"
	case "CHINA_HONG_KONG":
		return "China / Hong Kong"
	case "GLOBAL":
		return "Global"
	default:
		return Display(r)
	}
}

func ageLabel(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return trim0(d.Minutes()) + " min ago"
	}
	if d < 48*time.Hour {
		return trim0(d.Hours()) + " h ago"
	}
	return trim0(d.Hours()/24) + " d ago"
}

func trim0(v float64) string {
	return strings.TrimSuffix(strings.TrimSuffix(jsonNum(v, 0), ".0"), ".")
}

func trim1(v float64) string { return jsonNum(v, 0) }

func jsonNum(v float64, prec int) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			return x
		}
	}
	return ""
}

func JournalsRootHint(root string) string {
	return filepath.Join(root, "journals")
}
