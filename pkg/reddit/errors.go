package reddit

import (
	"errors"
	"fmt"
)

// RedditAPIError is the base error type for Reddit API errors.
type RedditAPIError struct {
	Message string
	Code    int
}

func (e *RedditAPIError) Error() string { return e.Message }

// NewRedditAPIError creates a new RedditAPIError.
func NewRedditAPIError(message string) *RedditAPIError {
	return &RedditAPIError{Message: message}
}

// SessionExpiredError indicates session cookies have expired.
type SessionExpiredError struct{ RedditAPIError }

// NewSessionExpiredError creates a SessionExpiredError.
func NewSessionExpiredError() *SessionExpiredError {
	return &SessionExpiredError{RedditAPIError{
		Message: "Session expired. Please re-login: reddit logout && reddit login",
		Code:    401,
	}}
}

// RateLimitError indicates Reddit rate-limited the request.
type RateLimitError struct {
	RedditAPIError
	RetryAfter float64
}

// NewRateLimitError creates a RateLimitError.
func NewRateLimitError(retryAfter float64) *RateLimitError {
	msg := "Rate limited by Reddit"
	if retryAfter > 0 {
		msg += fmt.Sprintf(" (retry after %.0fs)", retryAfter)
	}
	return &RateLimitError{
		RedditAPIError: RedditAPIError{Message: msg, Code: 429},
		RetryAfter:     retryAfter,
	}
}

// NotFoundError indicates a resource was not found.
type NotFoundError struct{ RedditAPIError }

// NewNotFoundError creates a NotFoundError.
func NewNotFoundError() *NotFoundError {
	return &NotFoundError{RedditAPIError{Message: "Resource not found", Code: 404}}
}

// ForbiddenError indicates access is forbidden.
type ForbiddenError struct{ RedditAPIError }

// NewForbiddenError creates a ForbiddenError.
func NewForbiddenError() *ForbiddenError {
	return &ForbiddenError{RedditAPIError{Message: "Access forbidden", Code: 403}}
}

// ErrorCodeFor maps an error to a stable error code string.
func ErrorCodeFor(err error) string {
	var sessionErr *SessionExpiredError
	if errors.As(err, &sessionErr) {
		return "not_authenticated"
	}
	var rateErr *RateLimitError
	if errors.As(err, &rateErr) {
		return "rate_limited"
	}
	var notFoundErr *NotFoundError
	if errors.As(err, &notFoundErr) {
		return "not_found"
	}
	var forbiddenErr *ForbiddenError
	if errors.As(err, &forbiddenErr) {
		return "forbidden"
	}
	var apiErr *RedditAPIError
	if errors.As(err, &apiErr) {
		return "api_error"
	}
	return "unknown_error"
}
