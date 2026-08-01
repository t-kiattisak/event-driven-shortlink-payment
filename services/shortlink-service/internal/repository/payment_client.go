package repository

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PaymentClient interface {
	FetchCheckoutHTML(ctx context.Context, paymentNo string) ([]byte, int, error)
}

type paymentClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPaymentClient(baseURL string) PaymentClient {
	return &paymentClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *paymentClient) FetchCheckoutHTML(ctx context.Context, paymentNo string) ([]byte, int, error) {
	url := fmt.Sprintf("%s/checkout/%s", c.baseURL, paymentNo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return body, resp.StatusCode, nil
}
