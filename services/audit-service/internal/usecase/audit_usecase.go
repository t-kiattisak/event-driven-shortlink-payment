package usecase

import (
	"context"
	"time"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/repository"
)

type AuditUseCase interface {
	RecordEvent(ctx context.Context, eventType, key, rawPayload string) error
}

type auditUseCase struct {
	repo repository.AuditRepository
}

func NewAuditUseCase(repo repository.AuditRepository) AuditUseCase {
	return &auditUseCase{repo: repo}
}

func (u *auditUseCase) RecordEvent(ctx context.Context, eventType, key, rawPayload string) error {
	log := &domain.AuditLog{
		EventID:   key,
		EventType: eventType,
		Actor:     "system",
		Payload:   rawPayload,
		CreatedAt: time.Now(),
	}
	return u.repo.SaveLog(ctx, log)
}
