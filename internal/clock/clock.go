package clock

import "time"

type Clock interface {
	Now() time.Time
}

type LiveClock struct{}

func (LiveClock) Now() time.Time { return time.Now().UTC() }

type ReplayClock struct {
	T time.Time
}

func (c *ReplayClock) Now() time.Time { return c.T.UTC() }

func (c *ReplayClock) Set(t time.Time) { c.T = t.UTC() }
