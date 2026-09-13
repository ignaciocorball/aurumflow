package main

import (
	"encoding/json"
	"os"
	"time"

	"aurumflow/internal/journal"
	"aurumflow/internal/logger"
	"aurumflow/internal/research"
)

func runProspectiveStatus() {
	st := research.ReadProspectiveStatus(research.ProspectiveDir)
	b, _ := json.MarshalIndent(st, "", "  ")
	logger.Info("prospective status %s", string(b))
}

func runLabelProspective() {
	dir := research.ProspectiveDir
	before, _ := os.ReadFile(research.InputPath(dir))
	n, err := research.LabelDue(dir, nil, time.Now().UTC())
	if err != nil {
		logger.Error("label-prospective: %v", err)
		os.Exit(1)
	}
	after, _ := os.ReadFile(research.InputPath(dir))
	if string(before) != string(after) {
		logger.Error("IMMUTABLE VIOLATION: signal_inputs.jsonl changed")
		os.Exit(1)
	}
	st := research.ReadProspectiveStatus(dir)
	logger.Info("label-prospective labeled=%d event=%s inputs_unchanged=true inputs=%d exhaustion=%d mature15m=%d mature1h=%d pending=%d",
		n, journal.EventProspectiveOutcome, st.Signals, st.Exhaustion, st.Mature15m, st.Mature1h, st.Pending)
}
