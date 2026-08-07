package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/t-kiattisak/event-driven-shortlink-payment/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentGRPCClient interface {
	FetchCheckoutHTML(ctx context.Context, paymentNo string) ([]byte, int, error)
	Close() error
}

type paymentGRPCClient struct {
	conn   *grpc.ClientConn
	client paymentv1.PaymentServiceClient
}

func NewPaymentGRPCClient(targetAddress string) (PaymentGRPCClient, error) {
	conn, err := grpc.Dial(targetAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment gRPC server: %w", err)
	}

	client := paymentv1.NewPaymentServiceClient(conn)
	return &paymentGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *paymentGRPCClient) FetchCheckoutHTML(ctx context.Context, paymentNo string) ([]byte, int, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.FetchCheckoutHTML(ctxTimeout, &paymentv1.FetchCheckoutHTMLRequest{
		PaymentNo: paymentNo,
	})
	if err != nil {
		return nil, 500, fmt.Errorf("payment gRPC RPC FetchCheckoutHTML failed: %w", err)
	}

	return resp.HtmlContent, int(resp.StatusCode), nil
}

func (c *paymentGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
