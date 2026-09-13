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
	ws := websocket.Dialer{HandshakeTimeout: 8 * time.Second}
	wctx, cancel2 := context.WithTimeout(ctx, 10*time.Second)
	wc, _, err := ws.DialContext(wctx, "wss://fstream.binance.com/ws/btcusdt@aggTrade", nil)
	cancel2()
	if err != nil {
		d.Connect = err.Error()
		return d
	}
	d.Connect = "ok"
	_ = wc.SetReadDeadline(time.Now().Add(8 * time.Second))
	_, _, err = wc.ReadMessage()
	_ = wc.Close()
	if err != nil {
		d.Read = err.Error()
		return d
	}
	d.Read = "ok"
	d.Status = "OPERATIONAL"
	return d
}
