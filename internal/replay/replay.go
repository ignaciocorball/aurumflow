package replay

import (
	"aurumflow/internal/clock"
	"aurumflow/internal/md"
)

type Mode string

const (
	ModeRealtime    Mode = "REALTIME"
	ModeAccelerated Mode = "ACCELERATED"
	ModeMaxSpeed    Mode = "MAX_SPEED"
	ModeStep        Mode = "STEP"
)

type Engine struct {
	Mode  Mode
	Clock *clock.ReplayClock
}

func New() *Engine {
	return &Engine{Mode: ModeMaxSpeed, Clock: &clock.ReplayClock{}}
}

func (e *Engine) Sort(events []md.Event) []md.Event {
	return md.OrderEvents(events)
}

func (e *Engine) Run(events []md.Event, fn func(md.Event, *clock.ReplayClock)) {
	ordered := e.Sort(events)
	for _, ev := range ordered {
		e.Clock.Set(ev.EventTime)
		fn(ev, e.Clock)
	}
}
