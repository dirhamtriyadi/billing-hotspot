// Package handlers contains the Gin HTTP handlers. They translate requests into
// service calls and render results through the consistent response envelope.
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/middleware"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/validatorx"
	"github.com/gin-gonic/gin"
)

// bindJSON binds and validates the request body into target. On failure it
// writes the appropriate error envelope (422 for validation, 400 for malformed
// JSON) and returns false so the caller can stop processing.
func bindJSON(c *gin.Context, target interface{}) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		if details, ok := validatorx.FieldErrors(err); ok {
			response.ValidationError(c, "Validation failed", details)
			return false
		}
		if validatorx.IsSyntaxError(err) {
			response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Malformed JSON request body")
			return false
		}
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return false
	}
	return true
}

// bindQuery binds and validates query parameters into target.
func bindQuery(c *gin.Context, target interface{}) bool {
	if err := c.ShouldBindQuery(target); err != nil {
		if details, ok := validatorx.FieldErrors(err); ok {
			response.ValidationError(c, "Validation failed", details)
			return false
		}
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return false
	}
	return true
}

// fail renders a service error. Typed AppErrors map to their status/code; any
// other error becomes a logged 500 so internals never leak to the client.
func fail(c *gin.Context, err error) {
	if ae, ok := apperror.As(err); ok {
		if ae.Err != nil {
			slog.Error("handled error",
				slog.String("code", ae.Code),
				slog.String("request_id", middleware.GetRequestID(c)),
				slog.Any("cause", ae.Err),
			)
		}
		response.Error(c, ae.Status, ae.Code, ae.Message)
		return
	}
	slog.Error("unhandled error",
		slog.String("request_id", middleware.GetRequestID(c)),
		slog.Any("error", err),
	)
	response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}
