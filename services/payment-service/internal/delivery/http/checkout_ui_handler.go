package http

import (
	"html/template"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/usecase"
)

type CheckoutUIHandler struct {
	paymentUseCase usecase.PaymentUseCase
}

func NewCheckoutUIHandler(app *fiber.App, paymentUseCase usecase.PaymentUseCase) {
	h := &CheckoutUIHandler{paymentUseCase: paymentUseCase}
	app.Get("/checkout/:paymentNo", h.RenderCheckoutUI)
}

func (h *CheckoutUIHandler) RenderCheckoutUI(c *fiber.Ctx) error {
	paymentNo := c.Params("paymentNo")
	payment, err := h.paymentUseCase.GetPaymentByNo(c.UserContext(), paymentNo)
	if err != nil {
		return c.Status(http.StatusNotFound).SendString("Payment Invoice Not Found")
	}

	return c.Render("checkout", fiber.Map{
		"PaymentNo":  payment.PaymentNo,
		"Amount":     payment.Amount,
		"Currency":   payment.Currency,
		"Status":     payment.Status,
		"QRCodeData": template.URL(payment.QRCodeData),
		"CreatedAt":  payment.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}
