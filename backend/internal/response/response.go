// Package response defines the single, consistent JSON envelope returned by
// every endpoint in the billing backend — success and error alike — including
// structured validation errors. Keeping this in one place guarantees the SPA
// can rely on a stable contract.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the canonical response shape.
//
// Successful responses populate Data (and optionally Meta); failed responses
// populate Error. Success is always present so clients can branch on a single
// boolean regardless of HTTP status.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// Meta carries pagination information for list endpoints.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ErrorBody describes a failure. Code is a stable machine-readable identifier;
// Details enumerates field-level validation problems when applicable.
type ErrorBody struct {
	Code    string       `json:"code"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError is a single validation failure tied to a request field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
}

// NewMeta computes pagination metadata.
func NewMeta(page, perPage int, total int64) *Meta {
	if perPage <= 0 {
		perPage = 1
	}
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	return &Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
}

// OK writes a 200 success response.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: data})
}

// Created writes a 201 success response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Message: message, Data: data})
}

// Paginated writes a 200 success response with list metadata.
func Paginated(c *gin.Context, message string, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: data, Meta: meta})
}

// NoContent acknowledges a successful mutation with no payload.
func NoContent(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message})
}

// Error writes a failure response with the given HTTP status and error code.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Envelope{
		Success: false,
		Message: message,
		Error:   &ErrorBody{Code: code},
	})
}

// ValidationError writes a 422 with structured field errors.
func ValidationError(c *gin.Context, message string, details []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Message: message,
		Error:   &ErrorBody{Code: "VALIDATION_ERROR", Details: details},
	})
}
