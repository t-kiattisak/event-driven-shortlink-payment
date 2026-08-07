package grpc

import (
	"bytes"
	"context"
	"html/template"

	"github.com/t-kiattisak/event-driven-shortlink-payment/proto/payment/v1"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGRPCServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	paymentUseCase usecase.PaymentUseCase
	tmpl           *template.Template
}

func NewPaymentGRPCServer(paymentUseCase usecase.PaymentUseCase) *PaymentGRPCServer {
	tmpl, err := template.ParseFiles("./internal/delivery/views/checkout.html")
	if err != nil {
		tmpl = nil
	}
	return &PaymentGRPCServer{
		paymentUseCase: paymentUseCase,
		tmpl:           tmpl,
	}
}

func (s *PaymentGRPCServer) FetchCheckoutHTML(ctx context.Context, req *paymentv1.FetchCheckoutHTMLRequest) (*paymentv1.FetchCheckoutHTMLResponse, error) {
	if req.PaymentNo == "" {
		return nil, status.Error(codes.InvalidArgument, "payment_no is required")
	}

	payment, err := s.paymentUseCase.GetPaymentByNo(ctx, req.PaymentNo)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "payment invoice not found: %v", err)
	}

	if s.tmpl == nil {
		var parseErr error
		s.tmpl, parseErr = template.ParseFiles("./internal/delivery/views/checkout.html")
		if parseErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to parse template: %v", parseErr)
		}
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"PaymentNo":  payment.PaymentNo,
		"Amount":     payment.Amount,
		"Currency":   payment.Currency,
		"Status":     payment.Status,
		"QRCodeData": template.URL(payment.QRCodeData),
		"CreatedAt":  payment.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if err := s.tmpl.Execute(&buf, data); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to execute template: %v", err)
	}

	return &paymentv1.FetchCheckoutHTMLResponse{
		HtmlContent: buf.Bytes(),
		StatusCode:  200,
	}, nil
}
