package grpc

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/proto/payment/v1"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentGRPCServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	paymentUseCase usecase.PaymentUseCase
}

func NewPaymentGRPCServer(paymentUseCase usecase.PaymentUseCase) *PaymentGRPCServer {
	return &PaymentGRPCServer{
		paymentUseCase: paymentUseCase,
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

	// Simple payload return for gRPC content forwarding
	_ = payment

	return &paymentv1.FetchCheckoutHTMLResponse{
		StatusCode: 200,
	}, nil
}
