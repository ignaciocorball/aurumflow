package notifications

import (
	"context"
	"time"

	"aurumflow/config"
	"aurumflow/internal/logger"

	"github.com/redis/go-redis/v9"
)

// RedisGlobalStore implements GlobalSentStore using Redis SET NX EX.
type RedisGlobalStore struct {
	client *redis.Client
}

// NewRedisGlobalStore creates a Redis-backed global dedup store. Caller must call Close when done.
func NewRedisGlobalStore(cfg *config.RedisDedupConfig) (*RedisGlobalStore, error) {
	if cfg == nil || cfg.Addr == "" {
		return nil, nil
	}
	opt := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &RedisGlobalStore{client: client}, nil
}

// TryMarkSent sets key only if it does not exist (NX), with TTL. Returns (true, nil) if we set it (first sender), (false, nil) if key already exists, (_, err) on error.
func (r *RedisGlobalStore) TryMarkSent(ctx context.Context, key string, ttl time.Duration) (first bool, err error) {
	if r == nil || r.client == nil {
		return true, nil
	}
	ok, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		logger.Warn("redis SetNX: %v", err)
		return false, err
	}
	return ok, nil
}

// Close closes the Redis client.
func (r *RedisGlobalStore) Close() error {
	if r == nil || r.client == nil {
		return nil
	}
	return r.client.Close()
}
