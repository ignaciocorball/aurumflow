package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"aurumflow/internal/killswitch"
	"aurumflow/internal/logger"
	"aurumflow/internal/ops"
)

func runStatusInspect(addr string) {
	if addr == "" {
		addr = "http://127.0.0.1:8765"
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(addr + "/status")
	if err != nil {
		logger.Error("status: %v (is --demo-week or --shadow-soak running?)", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func runHaltNewOrders() {
	ks := killswitch.New(false, killswitch.DefaultFile)
	if err := ks.HaltPersist(); err != nil {
		logger.Error("halt: %v", err)
		os.Exit(1)
	}
	logger.Info("HALT_NEW_ORDERS persisted to %s", killswitch.DefaultFile)
	_ = json.NewEncoder(os.Stdout).Encode(ops.Status{KillSwitch: true, LastExecution: "HALT_NEW_ORDERS"})
}
