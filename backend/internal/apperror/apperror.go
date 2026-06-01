// Package apperror provides a typed error that carries an HTTP status and a
// stable error code. Services return these; the HTTP layer renders them through
// the consistent response envelope so business logic never touches gin.Context.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError is a domain error enriched with transport metadata.
type AppError struct {
	Status  int    // HTTP status to emit
	Code    string // stable machine-readable code, e.g. "NOT_FOUND"
	Message string // human-readable message safe to expose to clients
	Err     error  // optional wrapped cause (logged, never serialized)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap exposes the wrapped cause for errors.Is / errors.As.
func (e *AppError) Unwrap() error { return e.Err }

// WithCause attaches an underlying error for logging without changing the
// client-facing message.
func (e *AppError) WithCause(err error) *AppError {
	e.Err = err
	return e
}

// As extracts an *AppError from any error chain, falling back to a generic
// 500 when the error is not an AppError.
func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// New builds a custom AppError.
func New(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// BadRequest indicates a malformed or semantically invalid request.
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized indicates missing or invalid authentication.
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return New(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden indicates the caller is authenticated but not allowed.
func Forbidden(message string) *AppError {
	if message == "" {
		message = "You do not have permission to perform this action"
	}
	return New(http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound indicates a missing resource.
func NotFound(message string) *AppError {
	if message == "" {
		message = "Resource not found"
	}
	return New(http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict indicates a state conflict (e.g. duplicate unique key).
func Conflict(message string) *AppError {
	return New(http.StatusConflict, "CONFLICT", message)
}

// Unprocessable indicates a request that is well-formed but cannot be processed
// due to business rules.
func Unprocessable(message string) *AppError {
	return New(http.StatusUnprocessableEntity, "UNPROCESSABLE", message)
}

// Internal indicates an unexpected server-side failure.
func Internal(message string) *AppError {
	if message == "" {
		message = "An unexpected error occurred"
	}
	return New(http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// ServiceUnavailable indicates a dependency (e.g. radius-api, payment gateway)
// is unreachable.
func ServiceUnavailable(message string) *AppError {
	if message == "" {
		message = "A required service is temporarily unavailable"
	}
	return New(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}
