package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type DBService struct {
	baseURL    string
	httpClient *http.Client
}

func NewDBService(baseURL string) *DBService {
	return &DBService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *DBService) PutRecord(ctx context.Context, payload PutRecordRequest) (*URLResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal put record payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/urls",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, fmt.Errorf("build put record request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("call db service put record: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted && res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("db service put record returned status %d", res.StatusCode)
	}

	var out URLResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode put record response: %w", err)
	}

	return &out, nil
}

func (c *DBService) GetRecord(ctx context.Context, shortCode string) (*URLResponse, error) {
	endpoint := c.baseURL + "/urls/" + url.PathEscape(shortCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build get record request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call db service get record: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("db service get record returned status %d", res.StatusCode)
	}

	var out URLResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode get record response: %w", err)
	}

	return &out, nil
}
