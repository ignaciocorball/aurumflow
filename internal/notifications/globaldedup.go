package notifications

import (
	"context"
	"time"
)

// GlobalSentStore registers in a shared backend that an event was already sent.
// TryMarkSent attempts to register the key; returns true if this instance registered it (and should send).
type GlobalSentStore interface {
	TryMarkSent(ctx context.Context, key string, ttl time.Duration) (first bool, err error)
}

// NoopGlobalStore is a no-op implementation: TryMarkSent always returns (true, nil).
// Use when global dedup is not configured so behavior is unchanged.
type NoopGlobalStore struct{}

// TryMarkSent always returns (true, nil) so the caller always sends.
func (NoopGlobalStore) TryMarkSent(ctx context.Context, key string, ttl time.Duration) (first bool, err error) {
	return true, nil
}

// GlobalDedupKey returns the global dedup key for POSITION_OPENED and POSITION_CLOSED events.
// Format: "POSITION_OPENED:{deal_ref}" or "POSITION_CLOSED:{deal_ref}".
// Returns empty string if ev is not one of these types or deal_ref is missing.
func GlobalDedupKey(ev NotifEvent) string {
	if ev.Type != TypePositionOpened && ev.Type != TypePositionClosed {
		return ""
	}
	dealRef, _ := ev.Payload["deal_ref"].(string)
	if dealRef == "" {
		return ""
	}
	return ev.Type + ":" + dealRef
}
