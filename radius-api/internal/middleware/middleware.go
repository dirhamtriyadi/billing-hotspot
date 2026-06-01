// Package middleware provides the radius-api HTTP middleware: correlation ids,
// structured logging, panic recovery, CORS and shared-key authentication.
package middleware

import (
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/dirhamt/billing-hotspot/radius-api/internal/config"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/response"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

// RequestID assigns/echoes a correlation id.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

func reqID(c *gin.Context) string {
	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Logger emits a structured entry per request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("request_id", reqID(c)),
		}
		switch s := c.Writer.Status(); {
		case s >= 500:
			slog.Error("request", attrs...)
		case s >= 400:
			slog.Warn("request", attrs...)
		default:
			slog.Info("request", attrs...)
		}
	}
}

// Recovery converts panics into a consistent 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered",
					slog.Any("error", r),
					slog.String("stack", string(debug.Stack())),
				)
				if !c.Writer.Written() {
					response.Error(c, 500, "INTERNAL_ERROR", "An unexpected error occurred")
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS configures cross-origin access.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "X-API-Key", requestIDHeader},
		MaxAge:       12 * time.Hour,
	})
}

// APIKey enforces the shared secret presented by the billing backend.
func APIKey(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key := c.GetHeader("X-API-Key"); key == "" || key != expected {
			response.Error(c, 401, "UNAUTHORIZED", "Missing or invalid API key")
			c.Abort()
			return
		}
		c.Next()
	}
}
