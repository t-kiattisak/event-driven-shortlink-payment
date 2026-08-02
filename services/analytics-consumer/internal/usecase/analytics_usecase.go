package usecase

import (
	"context"
	"time"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/analytics-consumer/internal/repository"
)

type AnalyticsUseCase interface {
	ProcessClickEvent(ctx context.Context, event *domain.ShortlinkClickedEvent) error
	ProcessPaymentCreated(ctx context.Context, event *domain.PaymentCreatedEvent) error
	ProcessPaymentStatusUpdated(ctx context.Context, event *domain.PaymentStatusUpdatedEvent) error
}

type analyticsUseCase struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsUseCase(repo repository.AnalyticsRepository) AnalyticsUseCase {
	return &analyticsUseCase{repo: repo}
}

func (u *analyticsUseCase) ProcessClickEvent(ctx context.Context, event *domain.ShortlinkClickedEvent) error {
	click := &domain.ClickAnalytics{
		Code:      event.Code,
		PaymentNo: event.PaymentNo,
		IPAddress: event.IPAddress,
		UserAgent: event.UserAgent,
		Referer:   event.Referer,
		ClickedAt: event.ClickedAt,
	}
	if click.ClickedAt.IsZero() {
		click.ClickedAt = time.Now()
	}
	return u.repo.RecordClick(ctx, click)
}

func (u *analyticsUseCase) ProcessPaymentCreated(ctx context.Context, event *domain.PaymentCreatedEvent) error {
	payment := &domain.PaymentAnalytics{
		PaymentNo:     event.PaymentNo,
		Amount:        event.Amount,
		Currency:      event.Currency,
		Status:        event.Status,
		CreatedAt:     event.CreatedAt,
		LastUpdatedAt: event.CreatedAt,
	}
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = time.Now()
		payment.LastUpdatedAt = time.Now()
	}
	return u.repo.UpsertPayment(ctx, payment)
}

func (u *analyticsUseCase) ProcessPaymentStatusUpdated(ctx context.Context, event *domain.PaymentStatusUpdatedEvent) error {
	payment := &domain.PaymentAnalytics{
		PaymentNo:     event.PaymentNo,
		Status:        event.NewStatus,
		LastUpdatedAt: event.UpdatedAt,
	}
	if payment.LastUpdatedAt.IsZero() {
		payment.LastUpdatedAt = time.Now()
	}
	return u.repo.UpsertPayment(ctx, payment)
}
