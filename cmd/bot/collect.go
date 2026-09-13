package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aurumflow/internal/binanceusdm"
	"aurumflow/internal/collector"
	"aurumflow/internal/logger"
	"aurumflow/internal/md"
	"aurumflow/internal/radar"
)

func runCollectMarketData(ctx context.Context, dur time.Duration) {
	if dur <= 0 {
		dur = time.Hour
	}
	runCtx, cancel := context.WithTimeout(ctx, dur)
	defer cancel()
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()
	day := time.Now().UTC().Format("2006/01/02")
	w := &collector.RotatingWriter{Dir: "data/live/binance/BTCUSDT/" + day, Name: "md"}
	defer w.Close()
	bus := md.NewBus(4096)
	ad := binanceusdm.New()
	eng := radar.NewEngine("BTCUSDT")
	eng.Book = ad.Book
	eng.Flow = ad.Flow
	go func() { _ = ad.Run(runCtx, bus) }()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-runCtx.Done():
			logger.Info("collector stop rows=%d bytes=%d reconnects=%d synced=%v", w.Rows, w.Bytes, ad.Reconnects, ad.Book.Synced)
			return
		case ev := <-bus.C():
			_ = w.Write(map[string]any{"kind": ev.Kind, "t": ev.EventTime, "seq": ev.Seq})
			if tr, ok := ev.Payload.(md.Trade); ok {
				eng.OnTrade(tr.Price, tr.Qty, tr.BuyerMaker)
			}
		case <-tick.C:
			_ = w.Write(map[string]any{"kind": "radar", "snap": eng.Snapshot(time.Now().UTC(), ad.Book.Synced, 1)})
		}
	}
}

func runCollectFreeMesh(ctx context.Context, dur time.Duration) {
	logger.Info("free-mesh collector: binance public + radar SHADOW (capital streaming optional/runtime-limited)")
	runCollectMarketData(ctx, dur)
}
