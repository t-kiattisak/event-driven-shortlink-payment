package repository

import (
	"context"
)

type PaymentClient interface {
	FetchCheckoutHTML(ctx context.Context, paymentNo string) ([]byte, int, error)
}

type paymentClient struct {
	grpcClient PaymentGRPCClient
}

func NewPaymentClient(grpcClient PaymentGRPCClient) PaymentClient {
	return &paymentClient{
		grpcClient: grpcClient,
	}
}

func (c *paymentClient) FetchCheckoutHTML(ctx context.Context, paymentNo string) ([]byte, int, error) {
	return c.grpcClient.FetchCheckoutHTML(ctx, paymentNo)
}
