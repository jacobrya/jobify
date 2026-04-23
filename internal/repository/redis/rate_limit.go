package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimitStore struct {
	client *redis.Client
}

func NewRateLimitStore(client *redis.Client) *RateLimitStore {
	return &RateLimitStore{client: client}
}

func (s *RateLimitStore) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

func (s *RateLimitStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Expire(ctx, key, ttl).Err()
}
