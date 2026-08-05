package repository

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/audit-service/internal/domain"
	"gorm.io/gorm"
)

type AuditRepository interface {
	SaveLog(ctx context.Context, log *domain.AuditLog) error
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) SaveLog(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
