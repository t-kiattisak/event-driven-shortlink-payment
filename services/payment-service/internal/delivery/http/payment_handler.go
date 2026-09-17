package http

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/delivery/http/middleware"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/usecase"
)

type PaymentHandler struct {
	paymentUseCase usecase.PaymentUseCase
}

func NewPaymentHandler(app *fiber.App, paymentUseCase usecase.PaymentUseCase, rdb *redis.Client) {
	h := &PaymentHandler{paymentUseCase: paymentUseCase}

	// REST API Routes (JSON Output Only)
	api := app.Group("/api/v1/payments")

	// Apply Rate Limiting (60 req/min) & Idempotency Key Validation (30s lock, 24h cache)
	if rdb != nil {
		api.Post("/",
			middleware.NewRateLimiter(60, time.Minute),
			middleware.IdempotencyMiddleware(rdb, 30*time.Second, 24*time.Hour),
			h.CreatePayment,
		)
	} else {
		api.Post("/", middleware.NewRateLimiter(60, time.Minute), h.CreatePayment)
	}

	api.Get("/:paymentNo", h.GetPayment)
	api.Get("/code/:code", h.GetPaymentByCode)
	api.Patch("/:paymentNo/status", h.UpdateStatus)
}

func (h *PaymentHandler) CreatePayment(c *fiber.Ctx) error {
	var input domain.CreatePaymentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	payment, err := h.paymentUseCase.CreatePayment(c.UserContext(), input)
	if err != nil {
		if err == domain.ErrInvalidAmount {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "payment created successfully",
		"data":    payment,
	})
}

func (h *PaymentHandler) GetPayment(c *fiber.Ctx) error {
	paymentNo := c.Params("paymentNo")
	payment, err := h.paymentUseCase.GetPaymentByNo(c.UserContext(), paymentNo)
	if err != nil {
		if err == domain.ErrPaymentNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": payment})
}

func (h *PaymentHandler) GetPaymentByCode(c *fiber.Ctx) error {
	code := c.Params("code")
	payment, err := h.paymentUseCase.GetPaymentByShortCode(c.UserContext(), code)
	if err != nil {
		if err == domain.ErrPaymentNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": payment})
}

func (h *PaymentHandler) UpdateStatus(c *fiber.Ctx) error {
	paymentNo := c.Params("paymentNo")
	var req struct {
		Status domain.PaymentStatus `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := h.paymentUseCase.UpdatePaymentStatus(c.UserContext(), paymentNo, req.Status); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "status updated successfully"})
}
