package core

import (
	"aurumflow/internal/journal"
)

func (l *Loop) journalLife(e journal.Lifecycle) {
	if l == nil || l.Journal == nil {
		return
	}
	if e.Epic == "" {
		e.Epic = l.Epic
	}
	if e.State == "" && l.State != nil {
		e.State = string(l.State.State())
	}
	if e.ExecutionMode == "" && l.Config != nil {
		e.ExecutionMode = string(l.Config.ExecMode())
	}
	_ = l.Journal.WriteLifecycle(e)
}
