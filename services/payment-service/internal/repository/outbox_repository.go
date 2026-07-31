package repository

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/domain"
	"gorm.io/gorm"
)

type OutboxRepository interface {
	FetchPending(ctx context.Context, limit int) ([]domain.Outbox, error)
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) FetchPending(ctx context.Context, limit int) ([]domain.Outbox, error) {
	var records []domain.Outbox
	err := r.db.WithContext(ctx).
		Where("status = ?", domain.OutboxPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *outboxRepository) MarkProcessed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Outbox{}).
		Where("id = ?", id).
		Update("status", domain.OutboxProcessed).Error
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Outbox{}).
		Where("id = ?", id).
		Update("status", domain.OutboxFailed).Error
}
