package intelrt

import (
	"aurumflow/internal/shadow"
)

const Mode = "UNIFIED_INTELLIGENCE_SHADOW"
const Port = "127.0.0.1:8766"

type Runtime struct {
	Mode string
}

func New() *Runtime { return &Runtime{Mode: Mode} }

func (r *Runtime) CanMutateBroker() bool { return false }

func (r *Runtime) HasExecutionProvider() bool {
	return shadow.HasExecutionProvider(r)
}

func BTCMicroHealthy(trades, l2, bookSynced, radar, exhaustion, absorption bool) bool {
	return trades && l2 && bookSynced && radar && exhaustion && absorption
}
