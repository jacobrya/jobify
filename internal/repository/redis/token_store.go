package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TokenStore struct {
	client *redis.Client
}

func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

func (s *TokenStore) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Duration) error {
	return s.client.Set(ctx, "refresh:"+token, userID.String(), expiry).Err()
}

func (s *TokenStore) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	val, err := s.client.Get(ctx, "refresh:"+token).Result()
	if errors.Is(err, redis.Nil) {
		return uuid.Nil, domain.ErrNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("malformed user id in token store: %w", err)
	}
	return id, nil
}

func (s *TokenStore) DeleteRefreshToken(ctx context.Context, token string) error {
	return s.client.Del(ctx, "refresh:"+token).Err()
}
