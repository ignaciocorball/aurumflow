package okxswap

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Diag struct {
	REST    string
	DNS     string
	TLS     string
	Connect string
	Read    string
	Status  string
}

func Diagnose(ctx context.Context) Diag {
	d := Diag{Status: "ENVIRONMENT_BLOCKED_OR_UNVERIFIED"}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, RESTBase+"/api/v5/market/books?instId="+InstID+"&sz=5", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		d.REST = err.Error()
	} else {
		_ = resp.Body.Close()
		d.REST = fmt.Sprintf("http_%d", resp.StatusCode)
	}
	host := "ws.okx.com"
	ips, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		d.DNS = err.Error()
		return d
	}
	d.DNS = fmt.Sprintf("ok n=%d", len(ips))
	dialer := &tls.Dialer{Config: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	conn, err := dialer.DialContext(cctx, "tcp", net.JoinHostPort(host, "8443"))
	cancel()
	if err != nil {
		d.TLS = err.Error()
		return d
	}
	_ = conn.Close()
	d.TLS = "ok"
	ws := websocket.Dialer{HandshakeTimeout: 8 * time.Second}
	wctx, cancel2 := context.WithTimeout(ctx, 10*time.Second)
	wc, _, err := ws.DialContext(wctx, WSURL, nil)
	cancel2()
	if err != nil {
		d.Connect = err.Error()
		return d
	}
	d.Connect = "ok"
	_ = wc.WriteJSON(map[string]any{
		"op": "subscribe",
		"args": []map[string]string{{"channel": "trades", "instId": InstID}},
	})
	_ = wc.SetReadDeadline(time.Now().Add(8 * time.Second))
	_, msg, err := wc.ReadMessage()
	_ = wc.Close()
	if err != nil {
		d.Read = err.Error()
		return d
	}
	d.Read = fmt.Sprintf("ok n=%d", len(msg))
	d.Status = "OPERATIONAL"
	return d
}
