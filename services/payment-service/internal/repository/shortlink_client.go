package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ShortlinkClient interface {
	CreateShortlink(ctx context.Context, paymentNo string) (string, error)
}

type shortlinkClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewShortlinkClient(baseURL string) ShortlinkClient {
	return &shortlinkClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *shortlinkClient) CreateShortlink(ctx context.Context, paymentNo string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/shortlinks", c.baseURL)
	payload, _ := json.Marshal(map[string]string{
		"payment_no": paymentNo,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to reach shortlink service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("shortlink service returned status: %d", resp.StatusCode)
	}

	var res struct {
		Data struct {
			Code string `json:"code"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Data.Code, nil
}
