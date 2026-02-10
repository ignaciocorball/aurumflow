package notifications

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"aurumflow/config"
	"aurumflow/internal/logger"
)

const (
	defaultQueueCap = 100
	droppedNotifyInterval = 10 * time.Minute
)

// Emitter is the single entry point for app notifications: enqueue to bounded channel, worker sends via Pushover and/or Telegram.
type Emitter struct {
	cfg             *config.Config
	policy          *Policy
	store           *Store
	client          *PushoverClient
	telegramClient  *TelegramClient
	globalStore     GlobalSentStore   // optional; when set, POSITION_OPENED/CLOSED use global dedup by deal_ref
	globalDedupTTL  time.Duration     // TTL for global dedup keys (e.g. 7 days)
	queue           chan NotifEvent
	done            chan struct{}
	wg              sync.WaitGroup
	dropped         atomic.Uint64
	lastDroppedAt   time.Time
	lastDroppedMu   sync.Mutex
}

const defaultGlobalDedupTTL = 7 * 24 * time.Hour

// NewEmitter builds an emitter from config. Active if Enabled and at least one sink (Pushover or Telegram) has valid credentials.
// globalStore is optional; when non-nil, POSITION_OPENED and POSITION_CLOSED are deduplicated globally by deal_ref.
func NewEmitter(cfg *config.Config, globalStore GlobalSentStore) *Emitter {
	if cfg == nil || cfg.Notifications == nil || !cfg.Notifications.Enabled {
		return &Emitter{cfg: cfg, policy: NewPolicy(nil), store: nil, client: nil, telegramClient: nil}
	}
	policy := NewPolicy(cfg.Notifications)
	var client *PushoverClient
	if cfg.Notifications.Provider == "pushover" && cfg.Notifications.Pushover != nil {
		token := strings.TrimSpace(cfg.Notifications.Pushover.Token)
		user := strings.TrimSpace(cfg.Notifications.Pushover.User)
		if token != "" && user != "" {
			client = NewPushoverClient(cfg.Notifications.Pushover)
		}
	}
	var telegramClient *TelegramClient
	if cfg.Notifications.Telegram != nil && cfg.Notifications.Telegram.Enabled {
		token := strings.TrimSpace(cfg.Notifications.Telegram.BotToken)
		chatID := strings.TrimSpace(cfg.Notifications.Telegram.ChatID)
		if token != "" && chatID != "" {
			telegramClient = NewTelegramClient(cfg.Notifications.Telegram)
		} else {
			logger.Warn("notifications.telegram enabled but missing bot_token or chat_id; telegram sink disabled")
		}
	}
	if client == nil && telegramClient == nil {
		return &Emitter{cfg: cfg, policy: policy, store: nil, client: nil, telegramClient: nil}
	}
	queueCap := cfg.Notifications.QueueCap
	if queueCap <= 0 {
		queueCap = defaultQueueCap
	}
	store := NewStore(policy.DedupWindow(), policy.AggregateWindow())
	globalDedupTTL := defaultGlobalDedupTTL
	if cfg.Notifications.GlobalDedup != nil && cfg.Notifications.GlobalDedup.TTLDays > 0 {
		globalDedupTTL = time.Duration(cfg.Notifications.GlobalDedup.TTLDays) * 24 * time.Hour
	}
	e := &Emitter{
		cfg:            cfg,
		policy:         policy,
		store:          store,
		client:         client,
		telegramClient: telegramClient,
		globalStore:    globalStore,
		globalDedupTTL: globalDedupTTL,
		queue:          make(chan NotifEvent, queueCap),
		done:           make(chan struct{}),
	}
	e.wg.Add(1)
	go e.worker()
	return e
}

// Active returns true if the emitter has at least one active sink (Pushover or Telegram).
func (e *Emitter) Active() bool {
	return e != nil && (e.client != nil || e.telegramClient != nil)
}

// Emit enqueues an event. If emitter is no-op (no sinks), returns immediately. If queue full, drops and increments counter.
func (e *Emitter) Emit(ev NotifEvent) {
	if e.client == nil && e.telegramClient == nil {
		return
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	select {
	case e.queue <- ev:
	default:
		e.dropped.Add(1)
	}
}

// Close stops the worker and drains the queue. Call on shutdown.
func (e *Emitter) Close() {
	if e.done == nil {
		return
	}
	close(e.done)
	e.wg.Wait()
	// Optional: drain remaining events (we don't send them after close)
	for {
		select {
		case <-e.queue:
		default:
			return
		}
	}
}

func (e *Emitter) worker() {
	defer e.wg.Done()
	apiMode := "demo"
	if e.cfg != nil {
		apiMode = e.cfg.API.Mode
	}
	aggFlushTicker := time.NewTicker(e.policy.AggregateWindow())
	defer aggFlushTicker.Stop()
	for {
		select {
		case <-e.done:
			// Flush dropped summary if any before exit
			if d := e.dropped.Load(); d > 0 {
				e.sendDroppedSummary(int(d))
			}
			return
		case ev, ok := <-e.queue:
			if !ok {
				return
			}
			e.processOne(ev, apiMode)
		case <-aggFlushTicker.C:
			e.flushRejectAggregate(apiMode)
			e.maybeSendDroppedSummary()
		}
	}
}

// telegramAllowed returns true if evType should be sent to Telegram (Events list empty = all allowed).
func (e *Emitter) telegramAllowed(evType string) bool {
	if e.cfg == nil || e.cfg.Notifications == nil || e.cfg.Notifications.Telegram == nil {
		return true
	}
	events := e.cfg.Notifications.Telegram.Events
	if len(events) == 0 {
		return true
	}
	for _, t := range events {
		if t == evType {
			return true
		}
	}
	return false
}

func (e *Emitter) processOne(ev NotifEvent, apiMode string) {
	if !e.policy.ShouldNotify(ev) {
		return
	}
	if ev.Type == TypeSweepConfirmed && !e.policy.SweepConfirmedShouldNotify(ev) {
		return
	}
	dedupKey := e.policy.DedupKey(ev)
	if e.store.ShouldDedup(dedupKey, ev) {
		return
	}
	minInterval := e.policy.MinInterval(ev.Type)
	if e.store.Throttled(ev.Type, minInterval) {
		// For aggregatable events, add to bucket instead of dropping
		if ev.Type == TypeSignalRejected {
			e.store.AddReject(ev)
		}
		return
	}
	if ev.Type == TypeSignalRejected {
		e.store.AddReject(ev)
		// Send aggregated on flush; don't send single reject when we aggregate
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if (ev.Type == TypePositionOpened || ev.Type == TypePositionClosed) && e.globalStore != nil {
		key := GlobalDedupKey(ev)
		if key != "" {
			first, err := e.globalStore.TryMarkSent(ctx, key, e.globalDedupTTL)
			if err != nil {
				logger.Warn("global dedup TryMarkSent: %v", err)
				return
			}
			if !first {
				return // already sent by another instance
			}
		}
	}
	title := BuildTitle(ev, e.cfg)
	msg := BuildMessage(ev, e.cfg)
	priority := e.policy.Priority(ev, apiMode)
	var successCount int
	if e.client != nil {
		if err := e.client.Send(ctx, title, msg, priority); err != nil {
			logger.Warn("pushover send: %v", err)
		} else {
			successCount++
		}
	}
	if e.telegramClient != nil && e.telegramAllowed(ev.Type) {
		titleTG := BuildTitleTelegram(ev, e.cfg)
		msgTG := BuildMessageTelegram(ev, e.cfg)
		textTG := titleTG + "\n" + msgTG
		tgCtx, tgCancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := e.telegramClient.Send(tgCtx, textTG); err != nil {
			logger.Warn("telegram send: %v", err)
		} else {
			successCount++
		}
		tgCancel()
	}
	if successCount > 0 {
		e.store.RecordSend(dedupKey, ev.Type, ev)
	} else if e.client != nil || e.telegramClient != nil {
		logger.Warn("notification send failed for all sinks (event=%s)", ev.Type)
	}
}

func (e *Emitter) flushRejectAggregate(apiMode string) {
	total, byReason, lastDir, lastScore, lastConf := e.store.FlushRejects()
	if total == 0 {
		return
	}
	ev := NotifEvent{
		Type:      TypeSignalRejected,
		Category:  CategorySignals,
		Severity:  SeverityInfo,
		Timestamp: time.Now().UTC(),
		Payload: map[string]any{
			"total":          total,
			"by_reason":      byReason,
			"last_direction": lastDir,
			"last_score":     lastScore,
			"last_conf":      lastConf,
		},
	}
	ev.DedupKey = e.policy.DedupKey(ev)
	title := BuildTitle(ev, e.cfg)
	msg := BuildMessage(ev, e.cfg)
	priority := e.policy.Priority(ev, apiMode)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if e.client != nil {
		if err := e.client.Send(ctx, title, msg, priority); err != nil {
			logger.Warn("pushover send (aggregate): %v", err)
		}
	}
	if e.telegramClient != nil && e.telegramAllowed(TypeSignalRejected) {
		titleTG := BuildTitleTelegram(ev, e.cfg)
		msgTG := BuildMessageTelegram(ev, e.cfg)
		textTG := titleTG + "\n" + msgTG
		tgCtx, tgCancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := e.telegramClient.Send(tgCtx, textTG); err != nil {
			logger.Warn("telegram send (aggregate): %v", err)
		}
		tgCancel()
	}
	// No RecordSend for aggregate; throttle is the ticker
}

func (e *Emitter) maybeSendDroppedSummary() {
	d := e.dropped.Load()
	if d == 0 {
		return
	}
	e.lastDroppedMu.Lock()
	if time.Since(e.lastDroppedAt) < droppedNotifyInterval {
		e.lastDroppedMu.Unlock()
		return
	}
	e.lastDroppedAt = time.Now()
	e.lastDroppedMu.Unlock()
	e.sendDroppedSummary(int(d))
	e.dropped.Store(0)
}

func (e *Emitter) sendDroppedSummary(dropped int) {
	ev := NotifEvent{
		Type:      TypeDroppedSummary,
		Category:  CategoryHealth,
		Severity:  SeverityInfo,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]any{"dropped": dropped},
	}
	title := BuildTitle(ev, e.cfg)
	msg := BuildMessage(ev, e.cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if e.client != nil {
		_ = e.client.Send(ctx, title, msg, 0)
	}
	if e.telegramClient != nil && e.telegramAllowed(TypeDroppedSummary) {
		titleTG := BuildTitleTelegram(ev, e.cfg)
		msgTG := BuildMessageTelegram(ev, e.cfg)
		textTG := titleTG + "\n" + msgTG
		tgCtx, tgCancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = e.telegramClient.Send(tgCtx, textTG)
		tgCancel()
	}
}
