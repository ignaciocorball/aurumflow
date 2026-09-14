package demomut

import (
	"errors"
	"strings"
	"sync"

	"aurumflow/internal/execacct"
	"aurumflow/internal/portfoliorisk"
)

const (
	OriginCalibration = "CALIBRATION_CANARY"
	OriginGold        = "GOLD_STRATEGY"
	OriginMirror      = "DEMO_MIRROR"
)

var (
	ErrLive        = errors.New("LIVE fail-closed")
	ErrNoDemo      = errors.New("explicit DEMO account required")
	ErrBusy        = errors.New("mutation in progress")
	ErrNotEligible = errors.New("DEMO_MIRROR gates failed")
	ErrHalt        = errors.New("HALT_NEW_ORDERS")
)

type Request struct {
	Origin        string
	Market        string
	Live          bool
	Env           string
	Account       execacct.Identity
	WantAccountID string
	Validated     bool
	MonetaryOK    bool
	DataFresh     bool
	Tradeable     bool
	Halt          bool
	OpenStrategy  int
	GroupOpen     map[string]int
}

type Coordinator struct {
	mu   sync.Mutex
	busy bool
	log  []string
}

func New() *Coordinator { return &Coordinator{} }

func (c *Coordinator) Reserve(req Request) error {
	if req.Live || strings.EqualFold(req.Env, "live") {
		return ErrLive
	}
	if !req.Account.MayTrade || !req.Account.Verified || req.Account.AccountID == "" {
		return ErrNoDemo
	}
	if req.WantAccountID != "" && req.Account.AccountID != req.WantAccountID {
		return ErrNoDemo
	}
	if req.Halt {
		return ErrHalt
	}
	if req.Origin == OriginMirror || req.Origin == OriginGold {
		if req.OpenStrategy >= 2 {
			return ErrNotEligible
		}
		g := portfoliorisk.GroupOf(req.Market)
		if g != "" && req.GroupOpen[g] >= 1 {
			return ErrNotEligible
		}
	}
	if req.Origin == OriginMirror {
		if !req.Validated || !req.MonetaryOK || !req.DataFresh || !req.Tradeable {
			return ErrNotEligible
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.busy {
		return ErrBusy
	}
	c.busy = true
	c.log = append(c.log, req.Origin+":"+req.Market)
	return nil
}

func (c *Coordinator) Release() {
	c.mu.Lock()
	c.busy = false
	c.mu.Unlock()
}

func (c *Coordinator) Log() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string{}, c.log...)
}
