package repository

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/domain"
	"gorm.io/gorm"
)

type OutboxRepository interface {
	FetchPending(ctx context.Context, limit int) ([]domain.Outbox, error)
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, err error) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) FetchPending(ctx context.Context, limit int) ([]domain.Outbox, error) {
	var records []domain.Outbox
	// Fetch PENDING events OR FAILED events with retry_count < 5
	err := r.db.WithContext(ctx).
		Where("status = ? OR (status = ? AND retry_count < ?)", domain.OutboxPending, domain.OutboxFailed, 5).
		Order("created_at ASC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *outboxRepository) MarkProcessed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Outbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     domain.OutboxProcessed,
			"last_error": nil,
		}).Error
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id string, failureErr error) error {
	errMsg := ""
	if failureErr != nil {
		errMsg = failureErr.Error()
	}

	return r.db.WithContext(ctx).
		Model(&domain.Outbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      domain.OutboxFailed,
			"retry_count": gorm.Expr("retry_count + ?", 1),
			"last_error":  errMsg,
		}).Error
}
