package repository

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AnalyticsRepository interface {
	RecordClick(ctx context.Context, click *domain.ClickAnalytics) error
	UpsertPayment(ctx context.Context, payment *domain.PaymentAnalytics) error
}

type analyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) RecordClick(ctx context.Context, click *domain.ClickAnalytics) error {
	return r.db.WithContext(ctx).Create(click).Error
}

func (r *analyticsRepository) UpsertPayment(ctx context.Context, payment *domain.PaymentAnalytics) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "payment_no"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "last_updated_at"}),
		}).
		Create(payment).Error
}
