package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/delivery/kafka"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/repository"
)

type ShortlinkUseCase interface {
	CreateShortlink(ctx context.Context, input domain.CreateShortlinkInput) (*domain.Shortlink, error)
	ResolveShortlink(ctx context.Context, code string, userAgent, ip, referer string) ([]byte, int, error)
}

type shortlinkUseCase struct {
	redisRepo     repository.RedisRepository
	paymentClient repository.PaymentClient
	eventProducer kafka.EventProducer
}

func NewShortlinkUseCase(
	redisRepo repository.RedisRepository,
	paymentClient repository.PaymentClient,
	eventProducer kafka.EventProducer,
) ShortlinkUseCase {
	return &shortlinkUseCase{
		redisRepo:     redisRepo,
		paymentClient: paymentClient,
		eventProducer: eventProducer,
	}
}

func (u *shortlinkUseCase) CreateShortlink(ctx context.Context, input domain.CreateShortlinkInput) (*domain.Shortlink, error) {
	if input.PaymentNo == "" && input.TargetURL == "" {
		return nil, domain.ErrInvalidURL
	}

	code := input.CustomCode
	if code == "" {
		code = uuid.New().String()[:6]
	}

	targetURL := input.TargetURL
	if targetURL == "" && input.PaymentNo != "" {
		targetURL = fmt.Sprintf("/checkout/%s", input.PaymentNo)
	}

	now := time.Now()
	expiresAt := now.Add(domain.DefaultTTL) // 7 Days Expiration Policy

	shortlink := &domain.Shortlink{
		Code:      code,
		PaymentNo: input.PaymentNo,
		TargetURL: targetURL,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	// 1. Save to Redis Cache with 7-Day TTL
	if err := u.redisRepo.Save(ctx, shortlink); err != nil {
		return nil, fmt.Errorf("failed to save shortlink to redis: %w", err)
	}

	// 2. Publish shortlink.created event to Kafka
	_ = u.eventProducer.PublishShortlinkCreated(ctx, shortlink)

	return shortlink, nil
}

func (u *shortlinkUseCase) ResolveShortlink(ctx context.Context, code string, userAgent, ip, referer string) ([]byte, int, error) {
	// 1. Fetch & Validate from Redis
	shortlink, err := u.redisRepo.Get(ctx, code)
	if err != nil {
		if err == domain.ErrCodeNotFound {
			return []byte("<h1>404 - Shortlink Not Found</h1>"), 404, nil
		}
		if err == domain.ErrCodeExpired {
			return []byte("<h1>410 - Link Expired</h1><p>This shortlink has expired after 7 days.</p>"), 410, nil
		}
		return nil, 500, err
	}

	// 2. Fire-and-Forget Click Event to Kafka
	clickEvent := &domain.ClickEvent{
		EventID:   uuid.New().String(),
		Code:      code,
		PaymentNo: shortlink.PaymentNo,
		UserAgent: userAgent,
		IPAddress: ip,
		Referer:   referer,
		Timestamp: time.Now(),
	}
	_ = u.eventProducer.PublishShortlinkClicked(ctx, clickEvent)

	// 3. Content Forwarding (Proxy HTML UI from payment-service)
	htmlBody, statusCode, err := u.paymentClient.FetchCheckoutHTML(ctx, shortlink.PaymentNo)
	if err != nil {
		return []byte("<h1>502 - Payment Service Unavailable</h1>"), 502, nil
	}

	return htmlBody, statusCode, nil
}
