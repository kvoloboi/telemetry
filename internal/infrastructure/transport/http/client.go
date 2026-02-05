package transporthttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	headers    http.Header
}

type Option func(*Client) error

func WithTimeout(d time.Duration) Option {
	return func(c *Client) error {
		c.httpClient.Timeout = d
		return nil
	}
}

func WithBaseURL(u string) Option {
	return func(c *Client) error {
		parsed, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("invalid base url: %w", err)
		}
		c.baseURL = parsed
		return nil
	}
}

func WithHeaders(h http.Header) Option {
	return func(c *Client) error {
		for k, v := range h {
			c.headers[k] = append(c.headers[k], v...)
		}
		return nil
	}
}

func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: &http.Client{},
		headers:    make(http.Header),
	}

	// Default to JSON content
	c.headers.Set("Content-Type", "application/json")
	c.headers.Set("Accept", "application/json")

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if c.baseURL == nil {
		return nil, errors.New("base URL is required")
	}

	return c, nil
}

// Post sends a POST request with `in` as the JSON body and decodes response into `out`.
func (c *Client) Post(ctx context.Context, path string, in any, out any) error {
	return c.doRequest(ctx, http.MethodPost, path, in, out)
}

// Get sends a GET request and decodes response into `out`.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, out)
}

// Put sends a PUT request with `in` as the JSON body and decodes response into `out`.
func (c *Client) Put(ctx context.Context, path string, in any, out any) error {
	return c.doRequest(ctx, http.MethodPut, path, in, out)
}

// Delete sends a DELETE request and decodes response into `out`.
func (c *Client) Delete(ctx context.Context, path string, out any) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, out)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, out any) error {
	// Build URL
	rel, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %w", path, err)
	}
	fullURL := c.baseURL.ResolveReference(rel)

	// Encode body if present
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), bodyReader)

	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Set default headers
	maps.Copy(req.Header, c.headers)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		const errBodySize = 1 << 10
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, errBodySize))
		return fmt.Errorf("http %d: %s", resp.StatusCode, payload)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
