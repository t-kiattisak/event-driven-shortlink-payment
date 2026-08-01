package grpc

import (
	"context"

	"github.com/t-kiattisak/event-driven-shortlink-payment/proto/shortlink/v1"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/domain"
	"github.com/t-kiattisak/event-driven-shortlink-payment/services/shortlink-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShortlinkGRPCServer struct {
	shortlinkv1.UnimplementedShortlinkServiceServer
	shortlinkUseCase usecase.ShortlinkUseCase
}

func NewShortlinkGRPCServer(shortlinkUseCase usecase.ShortlinkUseCase) *ShortlinkGRPCServer {
	return &ShortlinkGRPCServer{
		shortlinkUseCase: shortlinkUseCase,
	}
}

func (s *ShortlinkGRPCServer) CreateShortlink(ctx context.Context, req *shortlinkv1.CreateShortlinkRequest) (*shortlinkv1.CreateShortlinkResponse, error) {
	if req.PaymentNo == "" {
		return nil, status.Error(codes.InvalidArgument, "payment_no is required")
	}

	input := domain.CreateShortlinkInput{
		PaymentNo:  req.PaymentNo,
		TargetURL:  req.TargetUrl,
		CustomCode: req.CustomCode,
	}

	shortlink, err := s.shortlinkUseCase.CreateShortlink(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create shortlink: %v", err)
	}

	return &shortlinkv1.CreateShortlinkResponse{
		Code:      shortlink.Code,
		PaymentNo: shortlink.PaymentNo,
		TargetUrl: shortlink.TargetURL,
		ExpiresAt: shortlink.ExpiresAt.Unix(),
	}, nil
}
