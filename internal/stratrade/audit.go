package stratrade

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Audit struct {
	ID         string              `json:"strategy_trade_id"`
	Events     []Event             `json:"events"`
	PreSignal  *PreSignalSnapshot  `json:"pre_signal,omitempty"`
	Final      *FinalRecord        `json:"final,omitempty"`
	Complete   bool                `json:"lifecycle_complete"`
	Ordered    bool                `json:"lifecycle_ordered"`
}

func Reconstruct(root, id string) (Audit, error) {
	a := Audit{ID: id}
	dir := filepath.Join(root, id)
	if b, err := os.ReadFile(filepath.Join(dir, "pre_signal.json")); err == nil {
		var s PreSignalSnapshot
		if json.Unmarshal(b, &s) == nil {
			a.PreSignal = &s
		}
	}
	if evs, err := readJSONL[Event](filepath.Join(dir, "lifecycle.jsonl")); err == nil {
		a.Events = evs
		names := make([]string, len(evs))
		for i, e := range evs {
			names[i] = e.Event
		}
		a.Complete = LifecycleComplete(names)
		a.Ordered = LifecycleOrdered(names)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "final.json")); err == nil {
		var f FinalRecord
		if json.Unmarshal(b, &f) == nil {
			a.Final = &f
		}
	}
	if a.PreSignal == nil && len(a.Events) == 0 && a.Final == nil {
		return a, fmt.Errorf("no forensic evidence for %s", id)
	}
	return a, nil
}

func FormatAudit(a Audit) string {
	var b strings.Builder
	fmt.Fprintf(&b, "STRATEGY TRADE %s\n", a.ID)
	if a.PreSignal != nil {
		fmt.Fprintf(&b, "signal=%s origin=%s dir=%s entry=%.2f sl=%.2f tp=%.2f\n",
			a.PreSignal.SignalID, a.PreSignal.Origin, a.PreSignal.Direction,
			a.PreSignal.EntryCandidate, a.PreSignal.StopLoss, a.PreSignal.TakeProfit)
	}
	for _, e := range a.Events {
		fmt.Fprintf(&b, "%s %s dealRef=%s dealId=%s\n", e.Timestamp.UTC().Format("15:04:05"), e.Event, e.DealReference, e.DealID)
	}
	fmt.Fprintf(&b, "ordered=%v complete=%v\n", a.Ordered, a.Complete)
	if a.Final != nil {
		fmt.Fprintf(&b, "exit=%s trade=%s ops=%s\n", a.Final.ExitReason, a.Final.TradeOutcome, a.Final.OperationalOutcome)
	}
	return b.String()
}

func readJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []T
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var v T
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			continue
		}
		out = append(out, v)
	}
	return out, sc.Err()
}
