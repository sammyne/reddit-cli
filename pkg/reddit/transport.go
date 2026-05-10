package reddit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ReadTransport handles HTTP GET requests with jitter, retry, and cookie merging.
type ReadTransport struct {
	client          *http.Client
	session         *SessionState
	fingerprint     BrowserFingerprint
	config          RuntimeConfig
	requestDelay    time.Duration
	lastRequestTime time.Time
	requestCount    int
}

// NewReadTransport creates a ReadTransport.
func NewReadTransport(session *SessionState, config RuntimeConfig, fp BrowserFingerprint) *ReadTransport {
	return &ReadTransport{
		client:       &http.Client{Timeout: config.Timeout},
		session:      session,
		fingerprint:  fp,
		config:       config,
		requestDelay: config.ReadRequestDelay,
	}
}

// Close releases transport resources.
func (t *ReadTransport) Close() {
	t.client.CloseIdleConnections()
}

// RequestCount returns the number of requests made.
func (t *ReadTransport) RequestCount() int {
	return t.requestCount
}

// Request performs an HTTP request with jitter, retry, and cookie merging.
func (t *ReadTransport) Request(method, path string, params map[string]string) (map[string]interface{}, error) {
	t.rateLimitDelay()
	var lastErr error

	for attempt := range t.config.MaxRetries {
		result, err := t.doRequest(method, path, params)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if isRetryable(err) {
			wait := time.Duration(math.Pow(2, float64(attempt)))*time.Second + time.Duration(rand.Float64()*float64(time.Second))
			log.Printf("HTTP error: %v, retrying in %v", err, wait)
			time.Sleep(wait)
			continue
		}
		return nil, err
	}
	if lastErr != nil {
		return nil, &RedditAPIError{Message: fmt.Sprintf("Request failed after %d retries: %v", t.config.MaxRetries, lastErr)}
	}
	return nil, &RedditAPIError{Message: fmt.Sprintf("Request failed after %d retries", t.config.MaxRetries)}
}

func (t *ReadTransport) doRequest(method, path string, params map[string]string) (map[string]interface{}, error) {
	fullURL := BaseURL + path
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		fullURL += "?" + q.Encode()
	}

	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, &RedditAPIError{Message: fmt.Sprintf("failed to create request: %v", err)}
	}

	// Set headers
	for k, v := range t.fingerprint.ReadHeaders() {
		req.Header.Set(k, v)
	}
	// Set cookies as raw header to avoid Go's strict cookie value validation.
	// Reddit cookies may contain characters like `"` that net/http rejects.
	if cookieStr := buildCookieHeader(t.session.Cookies); cookieStr != "" {
		req.Header.Set("Cookie", cookieStr)
	}

	t0 := time.Now()
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, &RedditAPIError{Message: fmt.Sprintf("network error: %v", err)}
	}
	defer resp.Body.Close()

	elapsed := time.Since(t0)
	t.mergeResponseCookies(resp)
	t.requestCount++
	t.lastRequestTime = time.Now()

	log.Printf("[#%d] %s %s -> %d (%.2fs)", t.requestCount, method, truncate(path, 80), resp.StatusCode, elapsed.Seconds())

	switch resp.StatusCode {
	case 429:
		retryAfter, _ := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64)
		if retryAfter == 0 {
			retryAfter = 5
		}
		return nil, NewRateLimitError(retryAfter)
	case 401:
		return nil, NewSessionExpiredError()
	case 403:
		return nil, NewForbiddenError()
	case 404:
		return nil, NewNotFoundError()
	}
	if resp.StatusCode >= 500 {
		return nil, &RedditAPIError{Message: fmt.Sprintf("server error: HTTP %d", resp.StatusCode)}
	}
	if resp.StatusCode >= 400 {
		return nil, &RedditAPIError{Message: fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &RedditAPIError{Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	text := strings.TrimSpace(string(body))
	if text == "" {
		return map[string]interface{}{}, nil
	}
	if strings.HasPrefix(text, "<") {
		return nil, &RedditAPIError{Message: "Received HTML instead of JSON (possible auth redirect)"}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &RedditAPIError{Message: fmt.Sprintf("failed to parse JSON: %v", err)}
	}
	return result, nil
}

func (t *ReadTransport) rateLimitDelay() {
	if t.requestDelay <= 0 {
		return
	}
	elapsed := time.Since(t.lastRequestTime)
	if elapsed < t.requestDelay {
		jitter := time.Duration(math.Max(0, rand.NormFloat64()*0.15+0.3) * float64(time.Second))
		// 5% chance of random long pause (2-5s)
		if rand.Float64() < 0.05 {
			jitter += time.Duration((2.0 + rand.Float64()*3.0) * float64(time.Second))
		}
		time.Sleep(t.requestDelay - elapsed + jitter)
	}
}

func (t *ReadTransport) mergeResponseCookies(resp *http.Response) {
	for _, c := range resp.Cookies() {
		if c.Value != "" {
			t.session.Cookies[c.Name] = c.Value
		}
	}
	t.session.RefreshCapabilities()
}

func isRetryable(err error) bool {
	var apiErr *RedditAPIError
	if !errors.As(err, &apiErr) {
		return true // network errors are retryable
	}
	// 5xx and rate limit are retryable
	return apiErr.Code >= 500 || apiErr.Code == 429
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func buildCookieHeader(cookies map[string]string) string {
	if len(cookies) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cookies))
	for k, v := range cookies {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "; ")
}
