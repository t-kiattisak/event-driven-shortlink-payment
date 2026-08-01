package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/domain"
)

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) RedisRepository {
	return &redisRepository{client: client}
}

func (r *redisRepository) Save(ctx context.Context, shortlink *domain.Shortlink) error {
	key := fmt.Sprintf("shortlink:%s", shortlink.Code)
	data, err := json.Marshal(shortlink)
	if err != nil {
		return err
	}

	ttl := time.Until(shortlink.ExpiresAt)
	if ttl <= 0 {
		ttl = domain.DefaultTTL
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisRepository) Get(ctx context.Context, code string) (*domain.Shortlink, error) {
	key := fmt.Sprintf("shortlink:%s", code)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, domain.ErrCodeNotFound
		}
		return nil, err
	}

	var shortlink domain.Shortlink
	if err := json.Unmarshal([]byte(val), &shortlink); err != nil {
		return nil, err
	}

	if time.Now().After(shortlink.ExpiresAt) {
		return nil, domain.ErrCodeExpired
	}

	return &shortlink, nil
}
