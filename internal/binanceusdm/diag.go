package binanceusdm

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type WSSDiag struct {
	DNS     string
	TLS     string
	Connect string
	Read    string
	Status  string
}

func DiagnoseWSS(ctx context.Context) WSSDiag {
	d := WSSDiag{Status: "ENVIRONMENT_BLOCKED_OR_UNVERIFIED"}
	host := "fstream.binance.com"
	ips, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		d.DNS = err.Error()
		return d
	}
	d.DNS = fmt.Sprintf("ok n=%d", len(ips))
	dialer := &tls.Dialer{Config: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	conn, err := dialer.DialContext(cctx, "tcp", net.JoinHostPort(host, "443"))
	cancel()
	if err != nil {
		d.TLS = err.Error()
		return d
	}
	_ = conn.Close()
	d.TLS = "ok"
	// Official 2026 split: /market for aggTrade, /public for depth. Legacy /ws retired 2026-04-23.
	urls := []string{
		"wss://fstream.binance.com/market/ws/btcusdt@aggTrade",
		"wss://fstream.binance.com/market/stream?streams=btcusdt@aggTrade",
		"wss://fstream.binance.com/public/ws/btcusdt@depth@100ms",
		"wss://fstream.binance.com/public/stream?streams=btcusdt@depth@100ms",
		"wss://fstream.binance.com/ws/btcusdt@aggTrade",
	}
	ws := websocket.Dialer{HandshakeTimeout: 8 * time.Second}
	for _, u := range urls {
		wctx, cancel2 := context.WithTimeout(ctx, 8*time.Second)
		wc, _, err := ws.DialContext(wctx, u, nil)
		cancel2()
		if err != nil {
			d.Connect = err.Error()
			continue
		}
		d.Connect = "ok " + u
		_ = wc.SetReadDeadline(time.Now().Add(6 * time.Second))
		_, _, err = wc.ReadMessage()
		_ = wc.Close()
		if err != nil {
			d.Read = err.Error()
			continue
		}
		d.Read = "ok " + u
		d.Status = "OPERATIONAL"
		return d
	}
	if d.Connect == "" {
		d.Connect = "all endpoints failed"
	}
	if d.Read == "" {
		d.Read = "FIRST_FRAME_TIMEOUT_OR_ERROR"
	}
	return d
}
