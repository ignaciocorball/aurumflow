package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"aurumflow/internal/terminal"
	"aurumflow/internal/terminal/news"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8770", "loopback bind address")
	gold := flag.String("gold", "http://127.0.0.1:8765", "GOLD status base")
	intel := flag.String("intel", "http://127.0.0.1:8766", "intelligence status base")
	us100 := flag.String("us100", "http://127.0.0.1:8767", "US100 status base")
	root := flag.String("root", ".", "repository root for journals and config")
	fixtures := flag.Bool("dev-fixtures", false, "dev-only fixtures; rejected unless AURUMFLOW_DEV_FIXTURE_MODE=1")
	flag.Parse()

	if *fixtures && os.Getenv("AURUMFLOW_DEV_FIXTURE_MODE") != "1" {
		fmt.Fprintln(os.Stderr, "fixtures rejected: production terminal does not load UI fixtures")
		os.Exit(2)
	}

	fmt.Printf("AurumFlow terminal addr=%s mutation=%s gold=%s intel=%s us100=%s\n",
		*addr, terminal.BrokerMutationCapability(), *gold, *intel, *us100)

	client := terminal.NewReadClient(4 * time.Second)
	store := news.NewStore(filepath.Join(*root, "data", "news", "store.json"))
	newsSvc := news.NewService(store, client)
	hub := terminal.NewHub(client, terminal.Endpoints{
		Gold: *gold, Intel: *intel, US100: *us100, Root: *root,
	}, newsSvc, filepath.Join(*root, "config", "econ", "schedule.json"), *fixtures)

	stop := make(chan struct{})
	go hub.Run(stop)

	srv := terminal.NewServer(*addr, hub)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	close(stop)
	_ = srv.Shutdown()
}
