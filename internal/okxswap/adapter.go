package okxswap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"aurumflow/internal/book"
	"aurumflow/internal/md"

	"github.com/gorilla/websocket"
)

const (
	Name     = "okx_swap_public"
	RESTBase = "https://www.okx.com"
	WSURL    = "wss://ws.okx.com:8443/ws/v5/public"
	InstID   = "BTC-USDT-SWAP"
	Relation = "CORRELATED_PROXY"
)

// Adapter is a public, unauthenticated OKX SWAP market-data sensor.
// It never logs in and never places orders. Not IDENTICAL to Binance BTCUSDT.
type Adapter struct {
	HTTP       *http.Client
	WSURL      string
	RESTBase   string
	InstID     string
	Book       *book.Book
	Reconnects int
	LastError  string
	Started    time.Time
}

func New() *Adapter {
	return &Adapter{
		HTTP: &http.Client{Timeout: 15 * time.Second},
		WSURL: WSURL, RESTBase: RESTBase, InstID: InstID,
		Book: book.New(), Started: time.Now().UTC(),
	}
}

func (a *Adapter) Name() string { return Name }

func (a *Adapter) Capabilities() md.Caps {
	return md.CapTrades | md.CapBookSnapshot | md.CapBookDelta | md.CapHealth
}

type restBooks struct {
	Code string `json:"code"`
	Data []struct {
		Ts   string     `json:"ts"`
		Bids [][]string `json:"bids"`
		Asks [][]string `json:"asks"`
		Seq  string     `json:"seqId"`
	} `json:"data"`
}

func (a *Adapter) FetchSnapshot(ctx context.Context) (int64, []md.Level, []md.Level, error) {
	url := fmt.Sprintf("%s/api/v5/market/books?instId=%s&sz=50", strings.TrimRight(a.RESTBase, "/"), a.InstID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := a.HTTP.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	var out restBooks
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, nil, nil, err
	}
	if len(out.Data) == 0 {
		return 0, nil, nil, fmt.Errorf("empty books")
	}
	seq, _ := strconv.ParseInt(out.Data[0].Seq, 10, 64)
	return seq, levels(out.Data[0].Bids), levels(out.Data[0].Asks), nil
}

func levels(rows [][]string) []md.Level {
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

func bookLevels(xs []md.Level) []book.Level {
	out := make([]book.Level, len(xs))
	for i, l := range xs {
		out[i] = book.Level{Price: l.Price, Qty: l.Qty}
	}
	return out
}

func (a *Adapter) Run(ctx context.Context, out *md.Bus) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := a.runOnce(ctx, out); err != nil {
			a.LastError = err.Error()
			a.Reconnects++
			if out != nil {
				out.Publish(ctx, md.Event{
					Kind: md.KindProviderHealth, EventTime: time.Now().UTC(), ReceiveTime: time.Now().UTC(),
					Provider: Name, Venue: "okx", Instrument: InstID,
					Payload: md.Health{OK: false, BookSynced: a.Book.Synced, Reconnects: a.Reconnects, Resyncs: a.Book.Resyncs, LastError: a.LastError},
				})
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func (a *Adapter) runOnce(ctx context.Context, out *md.Bus) error {
	seq, bids, asks, err := a.FetchSnapshot(ctx)
	if err != nil {
		return err
	}
	a.Book.ApplySnapshot(seq, bookLevels(bids), bookLevels(asks))
	if out != nil {
		out.Publish(ctx, md.Event{
			Kind: md.KindBookSnapshot, EventTime: time.Now().UTC(), ReceiveTime: time.Now().UTC(),
			Provider: Name, Venue: "okx", Instrument: InstID, Seq: seq,
			Payload: md.BookDelta{Bids: bids, Asks: asks, FinalID: seq},
		})
	}
	dial := websocket.Dialer{HandshakeTimeout: 10 * time.Second}
	wc, _, err := dial.DialContext(ctx, a.WSURL, nil)
	if err != nil {
		return err
	}
	defer wc.Close()
	sub := map[string]any{
		"op": "subscribe",
		"args": []map[string]string{
			{"channel": "books", "instId": a.InstID},
			{"channel": "trades", "instId": a.InstID},
		},
	}
	if err := wc.WriteJSON(sub); err != nil {
		return err
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		_ = wc.SetReadDeadline(time.Now().Add(25 * time.Second))
		_, msg, err := wc.ReadMessage()
		if err != nil {
			return err
		}
		if string(msg) == "ping" {
			_ = wc.WriteMessage(websocket.TextMessage, []byte("pong"))
			continue
		}
		a.handle(ctx, msg, out)
	}
}

type wsMsg struct {
	Arg  struct{ Channel string `json:"channel"` } `json:"arg"`
	Action string `json:"action"`
	Data []json.RawMessage `json:"data"`
}

func (a *Adapter) handle(ctx context.Context, msg []byte, out *md.Bus) {
	var m wsMsg
	if json.Unmarshal(msg, &m) != nil {
		return
	}
	now := time.Now().UTC()
	switch m.Arg.Channel {
	case "trades":
		for _, raw := range m.Data {
			var t2 struct {
				Px   string `json:"px"`
				Sz   string `json:"sz"`
				Side string `json:"side"`
				Ts   string `json:"ts"`
			}
			if json.Unmarshal(raw, &t2) != nil {
				continue
			}
			px, _ := strconv.ParseFloat(t2.Px, 64)
			sz, _ := strconv.ParseFloat(t2.Sz, 64)
			ms, _ := strconv.ParseInt(t2.Ts, 10, 64)
			et := time.UnixMilli(ms).UTC()
			if et.IsZero() {
				et = now
			}
			maker := strings.EqualFold(t2.Side, "sell")
			if out != nil {
				out.Publish(ctx, md.Event{
					Kind: md.KindTrade, EventTime: et, ReceiveTime: now,
					Provider: Name, Venue: "okx", Instrument: InstID,
					Payload: md.Trade{Price: px, Qty: sz, BuyerMaker: maker},
				})
			}
		}
	case "books":
		for _, raw := range m.Data {
			var d struct {
				Bids     [][]string `json:"bids"`
				Asks     [][]string `json:"asks"`
				SeqId    int64      `json:"seqId"`
				PrevSeq  int64      `json:"prevSeqId"`
				Ts       string     `json:"ts"`
			}
			if json.Unmarshal(raw, &d) != nil {
				continue
			}
			bl, al := levels(d.Bids), levels(d.Asks)
			if m.Action == "snapshot" {
				a.Book.ApplySnapshot(d.SeqId, bookLevels(bl), bookLevels(al))
				et := eventTS(d.Ts, now)
				if out != nil {
				out.Publish(ctx, md.Event{
					Kind: md.KindBookSnapshot, EventTime: et, ReceiveTime: now,
						Provider: Name, Venue: "okx", Instrument: InstID, Seq: d.SeqId,
						Payload: md.BookDelta{Bids: bl, Asks: al, FinalID: d.SeqId},
					})
				}
				continue
			}
			if err := a.Book.ApplySeqDelta(d.SeqId, d.PrevSeq, bookLevels(bl), bookLevels(al)); err != nil {
				if seq, bids, asks, sErr := a.FetchSnapshot(context.Background()); sErr == nil {
					a.Book.ApplySnapshot(seq, bookLevels(bids), bookLevels(asks))
				}
				continue
			}
			et := eventTS(d.Ts, now)
			if out != nil {
				out.Publish(ctx, md.Event{
					Kind: md.KindBookDelta, EventTime: et, ReceiveTime: now,
					Provider: Name, Venue: "okx", Instrument: InstID, Seq: d.SeqId,
					Payload: md.BookDelta{Bids: bl, Asks: al, FirstID: d.PrevSeq, FinalID: d.SeqId, PrevFinal: d.PrevSeq},
				})
			}
		}
	}
}

func eventTS(ts string, fallback time.Time) time.Time {
	ms, err := strconv.ParseInt(ts, 10, 64)
	if err != nil || ms <= 0 {
		return fallback
	}
	return time.UnixMilli(ms).UTC()
}

type instrumentResp struct {
	Code string `json:"code"`
	Data []struct {
		InstID   string `json:"instId"`
		InstType string `json:"instType"`
		State    string `json:"state"`
	} `json:"data"`
}

func (a *Adapter) FetchInstrument(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s/api/v5/public/instruments?instType=SWAP&instId=%s", strings.TrimRight(a.RESTBase, "/"), a.InstID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := a.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out instrumentResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Data) == 0 {
		return "", fmt.Errorf("instrument not listed")
	}
	return out.Data[0].InstID, nil
}
