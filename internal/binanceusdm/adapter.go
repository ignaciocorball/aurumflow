package binanceusdm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"aurumflow/internal/book"
	"aurumflow/internal/flow"
	"aurumflow/internal/md"

	"github.com/gorilla/websocket"
)

const (
	RestBase = "https://fapi.binance.com"
	WSBase   = "wss://fstream.binance.com/market/stream"
	WSMarket = "wss://fstream.binance.com/market/stream"
	WSPublic = "wss://fstream.binance.com/public/stream"
	// WSPublicSingle is the official 2026 USD-M public single-stream base
	// (Diff Book Depth: /public/ws/<symbol>@depth@100ms).
	WSPublicSingle = "wss://fstream.binance.com/public/ws"
	Symbol         = "BTCUSDT"
)

// Adapter is a public, read-only Binance USD-M market-data provider.
// It never authenticates and never places orders.
type Adapter struct {
	HTTP       *http.Client
	WSURL      string
	RESTBase   string
	Symbol     string
	Book       *book.Book
	Flow       *flow.Engine
	Reconnects int
	LastError  string
	Started    time.Time
	lastAgg    int64

	mu       sync.Mutex
	buffered []depthEvent
}

func New() *Adapter {
	return &Adapter{
		HTTP:     &http.Client{Timeout: 15 * time.Second},
		WSURL:    WSBase,
		RESTBase: RestBase,
		Symbol:   Symbol,
		Book:     book.New(),
		Flow:     &flow.Engine{},
		Started:  time.Now().UTC(),
	}
}

func (a *Adapter) Name() string { return "binance_usdm_public" }

func (a *Adapter) Capabilities() md.Caps {
	return md.CapTrades | md.CapBookSnapshot | md.CapBookDelta | md.CapHealth
}

type depthSnapshot struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

type streamWrap struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

type depthEvent struct {
	Event     string     `json:"e"`
	EventTime int64      `json:"E"`
	FirstID   int64      `json:"U"`
	FinalID   int64      `json:"u"`
	PrevFinal int64      `json:"pu"`
	Bids      [][]string `json:"b"`
	Asks      [][]string `json:"a"`
}

type aggTrade struct {
	ID         int64  `json:"a"`
	EventTime  int64  `json:"E"`
	TradeTime  int64  `json:"T"`
	Price      string `json:"p"`
	Qty        string `json:"q"`
	BuyerMaker bool   `json:"m"`
}

func (a *Adapter) FetchSnapshot(ctx context.Context) (depthSnapshot, error) {
	url := fmt.Sprintf("%s/fapi/v1/depth?symbol=%s&limit=1000", strings.TrimRight(a.RESTBase, "/"), a.Symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return depthSnapshot{}, err
	}
	resp, err := a.HTTP.Do(req)
	if err != nil {
		return depthSnapshot{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return depthSnapshot{}, fmt.Errorf("depth snapshot HTTP %d", resp.StatusCode)
	}
	var snap depthSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		return snap, err
	}
	return snap, nil
}

func (a *Adapter) FetchAggTrades(ctx context.Context) ([]aggTrade, error) {
	url := fmt.Sprintf("%s/fapi/v1/aggTrades?symbol=%s&limit=200", strings.TrimRight(a.RESTBase, "/"), a.Symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("aggTrades HTTP %d", resp.StatusCode)
	}
	var out []aggTrade
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *Adapter) emitTrade(ctx context.Context, out *md.Bus, tr aggTrade) {
	a.mu.Lock()
	if tr.ID != 0 && tr.ID <= a.lastAgg {
		a.mu.Unlock()
		return
	}
	if tr.ID > a.lastAgg {
		a.lastAgg = tr.ID
	}
	a.mu.Unlock()
	px, _ := strconv.ParseFloat(tr.Price, 64)
	qty, _ := strconv.ParseFloat(tr.Qty, 64)
	a.Flow.OnTrade(flow.Trade{Price: px, Qty: qty, BuyerMaker: tr.BuyerMaker})
	et := time.UnixMilli(tr.TradeTime).UTC()
	if tr.TradeTime == 0 {
		et = time.UnixMilli(tr.EventTime).UTC()
	}
	if et.IsZero() {
		et = time.Now().UTC()
	}
	out.Publish(ctx, md.Event{
		Kind: md.KindTrade, EventTime: et, ReceiveTime: time.Now().UTC(),
		Provider: a.Name(), Venue: "binance_usdm", Instrument: a.Symbol, Seq: tr.ID,
		Payload: md.Trade{Price: px, Qty: qty, BuyerMaker: tr.BuyerMaker},
	})
}

func (a *Adapter) runREST(ctx context.Context, out *md.Bus) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if trades, err := a.FetchAggTrades(ctx); err == nil {
				a.mu.Lock()
				seeded := a.lastAgg == 0 && len(trades) > 0
				if seeded {
					a.lastAgg = trades[len(trades)-1].ID
				}
				a.mu.Unlock()
				if !seeded {
					for _, tr := range trades {
						a.emitTrade(ctx, out, tr)
					}
				}
			} else {
				a.LastError = err.Error()
			}
			// REST snapshots are resync-only. Treating 1Hz REST refresh as
			// incremental flow desynchronizes the official depth stream.
			if synced, _, _, _ := a.Book.Meta(); !synced {
				if snap, err := a.FetchSnapshot(ctx); err == nil {
					a.applySnapshot(snap)
					a.Book.NoteResync("rest_unsynced")
					out.Publish(ctx, md.Event{
						Kind: md.KindBookSnapshot, EventTime: time.Now().UTC(), ReceiveTime: time.Now().UTC(),
						Provider: a.Name(), Venue: "binance_usdm", Instrument: a.Symbol, Seq: snap.LastUpdateID,
						Payload: md.BookDelta{Bids: levelsMD(snap.Bids), Asks: levelsMD(snap.Asks), FinalID: snap.LastUpdateID},
					})
				}
			}
		}
	}
}

func levels(rows [][]string) []book.Level {
	out := make([]book.Level, 0, len(rows))
	for _, r := range rows {
		if len(r) < 2 {
			continue
		}
		p, _ := strconv.ParseFloat(r[0], 64)
		q, _ := strconv.ParseFloat(r[1], 64)
		out = append(out, book.Level{Price: p, Qty: q})
	}
	return out
}

func levelsMD(rows [][]string) []md.Level {
	out := make([]md.Level, 0, len(rows))
	for _, r := range rows {
		if len(r) < 2 {
			continue
		}
		p, _ := strconv.ParseFloat(r[0], 64)
		q, _ := strconv.ParseFloat(r[1], 64)
		out = append(out, md.Level{Price: p, Qty: q})
	}
	return out
}

func (a *Adapter) applySnapshot(snap depthSnapshot) {
	a.Book.ApplySnapshot(snap.LastUpdateID, levels(snap.Bids), levels(snap.Asks))
}

func (a *Adapter) applyBuffered(snapID int64) {
	a.mu.Lock()
	buf := a.buffered
	a.buffered = nil
	a.mu.Unlock()
	started := false
	for _, ev := range buf {
		if Obsolete(snapID, ev.FinalID) {
			continue
		}
		if !started {
			if !FirstApplicable(snapID, ev.FirstID, ev.FinalID) {
				a.Book.MarkUnsynced()
				a.Book.NoteResync("buffered_not_applicable")
				return
			}
			started = true
		}
		_ = a.Book.ApplyFuturesDelta(ev.FirstID, ev.FinalID, ev.PrevFinal, levels(ev.Bids), levels(ev.Asks))
		if synced, _, _, _ := a.Book.Meta(); !synced {
			return
		}
	}
}

func (a *Adapter) Run(ctx context.Context, out *md.Bus) error {
	go a.runREST(ctx, out)
	backoff := time.Second
	rotateEvery := 23 * time.Hour
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := a.runOnce(ctx, out, rotateEvery)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		a.Reconnects++
		a.LastError = errString(err)
		a.Book.MarkUnsynced()
		_, _, _, resyncs := a.Book.Meta()
		_ = out.Publish(ctx, md.Event{
			Kind: md.KindProviderHealth, EventTime: time.Now().UTC(), ReceiveTime: time.Now().UTC(),
			Provider: a.Name(), Venue: "binance_usdm", Instrument: a.Symbol,
			Payload: md.Health{OK: false, BookSynced: false, Reconnects: a.Reconnects, Resyncs: resyncs, LastError: a.LastError},
		})
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		if backoff < 60*time.Second {
			backoff *= 2
		}
	}
}

func errString(err error) string {
	if err == nil {
		return "disconnected"
	}
	return err.Error()
}

func (a *Adapter) runOnce(ctx context.Context, out *md.Bus, rotateEvery time.Duration) error {
	sym := strings.ToLower(a.Symbol)
	tradeURL := fmt.Sprintf("%s?streams=%s@aggTrade", WSMarket, sym)
	// Official Diff Book Depth (100ms) on the 2026 public single-stream path.
	depthURL := fmt.Sprintf("%s/%s@depth@100ms", WSPublicSingle, sym)
	dialer := websocket.Dialer{HandshakeTimeout: 15 * time.Second}
	connT, _, err := dialer.DialContext(ctx, tradeURL, nil)
	if err != nil {
		return err
	}
	defer connT.Close()
	connD, _, err := dialer.DialContext(ctx, depthURL, nil)
	if err != nil {
		return err
	}
	defer connD.Close()

	runCtx, cancel := context.WithTimeout(ctx, rotateEvery)
	defer cancel()

	a.mu.Lock()
	a.buffered = nil
	a.mu.Unlock()
	a.Book.MarkUnsynced()

	snapCh := make(chan depthSnapshot, 1)
	go func() {
		time.Sleep(200 * time.Millisecond)
		snap, err := a.FetchSnapshot(runCtx)
		if err != nil {
			a.LastError = err.Error()
			close(snapCh)
			return
		}
		snapCh <- snap
	}()

	arm := func(c *websocket.Conn) {
		_ = c.SetReadDeadline(time.Now().Add(90 * time.Second))
		c.SetPongHandler(func(string) error {
			return c.SetReadDeadline(time.Now().Add(90 * time.Second))
		})
	}
	arm(connT)
	arm(connD)

	var snapApplied bool
	pingTick := time.NewTicker(15 * time.Second)
	defer pingTick.Stop()

	type incoming struct {
		msg []byte
		err error
	}
	reads := make(chan incoming, 8)
	readLoop := func(c *websocket.Conn) {
		for {
			_, msg, err := c.ReadMessage()
			select {
			case reads <- incoming{msg, err}:
			case <-runCtx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}
	go readLoop(connT)
	go readLoop(connD)

	for {
		select {
		case <-runCtx.Done():
			return runCtx.Err()
		case <-pingTick.C:
			_ = connT.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
			_ = connD.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
		case snap, ok := <-snapCh:
			if ok {
				a.applySnapshot(snap)
				a.applyBuffered(snap.LastUpdateID)
				snapApplied = true
				_ = out.Publish(runCtx, md.Event{
					Kind: md.KindBookSnapshot, EventTime: time.Now().UTC(), ReceiveTime: time.Now().UTC(),
					Provider: a.Name(), Venue: "binance_usdm", Instrument: a.Symbol, Seq: snap.LastUpdateID,
					Payload: md.BookDelta{Bids: levelsMD(snap.Bids), Asks: levelsMD(snap.Asks), FinalID: snap.LastUpdateID},
				})
			}
			snapCh = nil
		case in := <-reads:
			if in.err != nil {
				return in.err
			}
			_ = connT.SetReadDeadline(time.Now().Add(90 * time.Second))
			_ = connD.SetReadDeadline(time.Now().Add(90 * time.Second))
			if err := a.handleMessage(runCtx, out, in.msg, snapApplied); err != nil {
				return err
			}
		}
	}
}

func (a *Adapter) handleMessage(ctx context.Context, out *md.Bus, raw []byte, snapApplied bool) error {
	recv := time.Now().UTC()
	var wrap streamWrap
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil
	}
	payload := wrap.Data
	if len(payload) == 0 {
		payload = raw
	}
	switch {
	case strings.Contains(wrap.Stream, "aggTrade") || bytesHas(payload, `"e":"aggTrade"`):
		var tr aggTrade
		if err := json.Unmarshal(payload, &tr); err != nil {
			return nil
		}
		a.emitTrade(ctx, out, tr)
	default:
		var ev depthEvent
		if err := json.Unmarshal(payload, &ev); err != nil {
			return nil
		}
		if !isDepthUpdate(wrap, ev) {
			return nil
		}
		synced, _, _, _ := a.Book.Meta()
		if !snapApplied || !synced {
			a.mu.Lock()
			a.buffered = append(a.buffered, ev)
			if len(a.buffered) > 5000 {
				a.buffered = a.buffered[len(a.buffered)-2500:]
			}
			a.mu.Unlock()
			return nil
		}
		if err := a.Book.ApplyFuturesDelta(ev.FirstID, ev.FinalID, ev.PrevFinal, levels(ev.Bids), levels(ev.Asks)); err != nil {
			// Detectable gap: resync from REST. Do not treat the stale book as truth.
			if snap, sErr := a.FetchSnapshot(ctx); sErr == nil {
				a.applySnapshot(snap)
				out.Publish(ctx, md.Event{
					Kind: md.KindBookSnapshot, EventTime: time.Now().UTC(), ReceiveTime: recv,
					Provider: a.Name(), Venue: "binance_usdm", Instrument: a.Symbol, Seq: snap.LastUpdateID,
					Payload: md.BookDelta{Bids: levelsMD(snap.Bids), Asks: levelsMD(snap.Asks), FinalID: snap.LastUpdateID},
				})
				return nil
			}
			return err
		}
		et := time.UnixMilli(ev.EventTime).UTC()
		if et.IsZero() {
			et = recv
		}
		out.Publish(ctx, md.Event{
			Kind: md.KindBookDelta, EventTime: et, ReceiveTime: recv,
			Provider: a.Name(), Venue: "binance_usdm", Instrument: a.Symbol, Seq: ev.FinalID,
			Payload: md.BookDelta{
				Bids: levelsMD(ev.Bids), Asks: levelsMD(ev.Asks),
				FirstID: ev.FirstID, FinalID: ev.FinalID, PrevFinal: ev.PrevFinal,
			},
		})
	}
	return nil
}

func isDepthUpdate(wrap streamWrap, ev depthEvent) bool {
	if ev.FinalID == 0 && ev.FirstID == 0 {
		return false
	}
	if ev.Event == "depthUpdate" {
		return true
	}
	return strings.Contains(wrap.Stream, "depth")
}

func bytesHas(b []byte, s string) bool {
	return strings.Contains(string(b), s)
}
