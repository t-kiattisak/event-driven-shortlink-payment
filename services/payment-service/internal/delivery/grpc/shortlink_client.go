package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/t-kiattisak/event-driven-shortlink-payment/proto/shortlink/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ShortlinkGRPCClient interface {
	CreateShortlink(ctx context.Context, paymentNo string) (string, error)
	Close() error
}

type shortlinkGRPCClient struct {
	conn   *grpc.ClientConn
	client shortlinkv1.ShortlinkServiceClient
}

func NewShortlinkGRPCClient(targetAddress string) (ShortlinkGRPCClient, error) {
	conn, err := grpc.Dial(targetAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to shortlink gRPC server: %w", err)
	}

	client := shortlinkv1.NewShortlinkServiceClient(conn)
	return &shortlinkGRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *shortlinkGRPCClient) CreateShortlink(ctx context.Context, paymentNo string) (string, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.CreateShortlink(ctxTimeout, &shortlinkv1.CreateShortlinkRequest{
		PaymentNo: paymentNo,
	})
	if err != nil {
		return "", fmt.Errorf("shortlink gRPC RPC CreateShortlink failed: %w", err)
	}

	return resp.Code, nil
}

func (c *shortlinkGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
