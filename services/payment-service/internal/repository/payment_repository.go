package repository

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/domain"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreateWithOutbox(ctx context.Context, payment *domain.Payment, outbox *domain.Outbox) error
	FindByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error)
	FindByShortCode(ctx context.Context, shortCode string) (*domain.Payment, error)
	UpdateStatus(ctx context.Context, paymentNo string, status domain.PaymentStatus, outbox *domain.Outbox) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) CreateWithOutbox(ctx context.Context, payment *domain.Payment, outbox *domain.Outbox) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payment).Error; err != nil {
			return err
		}
		if err := tx.Create(outbox).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *paymentRepository) FindByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) FindByShortCode(ctx context.Context, shortCode string) (*domain.Payment, error) {
	var payment domain.Payment
	if err := r.db.WithContext(ctx).Where("short_code = ?", shortCode).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, paymentNo string, status domain.PaymentStatus, outbox *domain.Outbox) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.Payment{}).Where("payment_no = ?", paymentNo).Update("status", status).Error; err != nil {
			return err
		}
		if err := tx.Create(outbox).Error; err != nil {
			return err
		}
		return nil
	})
}
