// Package response defines the consistent JSON envelope for the radius-api,
// matching the contract used by the billing backend.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the canonical response shape.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// ErrorBody describes a failure.
type ErrorBody struct {
	Code string `json:"code"`
}

// OK writes a 200 success response.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: data})
}

// Created writes a 201 success response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Message: message, Data: data})
}

// NoContent acknowledges a successful mutation.
func NoContent(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message})
}

// Error writes a failure response.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Envelope{Success: false, Message: message, Error: &ErrorBody{Code: code}})
}
