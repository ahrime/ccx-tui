package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type APIClient struct {
	baseURL    string
	accessKey  string
	httpClient *http.Client
}

func NewAPIClient(baseURL, accessKey string) *APIClient {
	return &APIClient{
		baseURL:   baseURL,
		accessKey: accessKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) BaseURL() string { return c.baseURL }

func (c *APIClient) buildURL(path string) string {
	return c.baseURL + path
}

func (c *APIClient) setAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.accessKey)
	req.Header.Set("Content-Type", "application/json")
}

func (c *APIClient) Get(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path), nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	c.setAuth(req)
	return c.do(req, result)
}

func (c *APIClient) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doWithBody(ctx, http.MethodPost, path, body, result)
}

func (c *APIClient) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doWithBody(ctx, http.MethodPut, path, body, result)
}

func (c *APIClient) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doWithBody(ctx, http.MethodPatch, path, body, result)
}

func (c *APIClient) Delete(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.buildURL(path), nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	c.setAuth(req)
	return c.do(req, result)
}

func (c *APIClient) doWithBody(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("encode body: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path), &buf)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	c.setAuth(req)
	return c.do(req, result)
}

func (c *APIClient) do(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *APIClient) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := c.Get(ctx, "/health", &result)
	return result, err
}
