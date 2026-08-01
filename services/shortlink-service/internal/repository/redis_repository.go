package repository

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/domain"
)

type RedisRepository interface {
	Save(ctx context.Context, shortlink *domain.Shortlink) error
	Get(ctx context.Context, code string) (*domain.Shortlink, error)
}
