package http

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/delivery/http/middleware"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/usecase"
)

type ShortlinkHandler struct {
	useCase usecase.ShortlinkUseCase
}

func NewShortlinkHandler(app *fiber.App, useCase usecase.ShortlinkUseCase) {
	h := &ShortlinkHandler{useCase: useCase}

	// Content Forwarding Endpoint (URL remains /s/:code) - Rate limit: 120 requests / minute per IP
	app.Get("/s/:code", middleware.NewRateLimiter(120, time.Minute), h.ResolveShortlink)

	// REST API Endpoint - Rate limit: 60 requests / minute per IP
	api := app.Group("/api/v1/shortlinks", middleware.NewRateLimiter(60, time.Minute))
	api.Post("/", h.CreateShortlink)
}

func (h *ShortlinkHandler) CreateShortlink(c *fiber.Ctx) error {
	var input domain.CreateShortlinkInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	shortlink, err := h.useCase.CreateShortlink(c.UserContext(), input)
	if err != nil {
		if err == domain.ErrInvalidURL {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "shortlink created successfully",
		"data":    shortlink,
	})
}

func (h *ShortlinkHandler) ResolveShortlink(c *fiber.Ctx) error {
	code := c.Params("code")
	userAgent := c.Get("User-Agent")
	ip := c.IP()
	referer := c.Get("Referer")

	htmlBody, statusCode, err := h.useCase.ResolveShortlink(c.UserContext(), code, userAgent, ip, referer)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
	}

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Status(statusCode).Send(htmlBody)
}
