package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	httpRequestTimeout = 10 * time.Second
	maxHTTPRespBody    = 1 << 20 // 1 MiB
)

type httpRequestConfig struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

// runHTTPRequest performs a single HTTP call and returns JSON output
// { "status_code": N, "body": ... } where body is decoded JSON or a string.
// Success is HTTP 2xx; non-2xx returns the same output shape plus a non-nil error.
func runHTTPRequest(ctx context.Context, config []byte) ([]byte, error) {
	if len(config) == 0 {
		return nil, fmt.Errorf("http_request: empty config")
	}
	var cfg httpRequestConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}
	method := strings.ToUpper(strings.TrimSpace(cfg.Method))
	if method == "" {
		return nil, fmt.Errorf("http_request: method is required")
	}
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
	default:
		return nil, fmt.Errorf("http_request: unsupported method %q", cfg.Method)
	}
	urlStr := strings.TrimSpace(cfg.URL)
	if urlStr == "" {
		return nil, fmt.Errorf("http_request: url is required")
	}

	var body io.Reader
	if (method == http.MethodPost || method == http.MethodPut) && len(cfg.Body) > 0 {
		body = bytes.NewReader(cfg.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}
	for k, v := range cfg.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: httpRequestTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPRespBody))
	if err != nil {
		return nil, err
	}
	out, err := json.Marshal(httpResponsePayload(resp.StatusCode, raw))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return out, &HTTPStatusError{StatusCode: resp.StatusCode}
	}
	return out, nil
}

// HTTPStatusError is returned when an http_request receives a non-2xx response.
type HTTPStatusError struct {
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("http_request: status %d", e.StatusCode)
}

// NonRetriable is true for 4xx client errors except 429 (rate limit), which may succeed after a wait.
func (e *HTTPStatusError) NonRetriable() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500 && e.StatusCode != http.StatusTooManyRequests
}

func httpResponsePayload(statusCode int, raw []byte) map[string]any {
	m := map[string]any{"status_code": statusCode}
	if json.Valid(raw) {
		var v any
		if err := json.Unmarshal(raw, &v); err == nil {
			m["body"] = v
			return m
		}
	}
	m["body"] = string(raw)
	return m
}
