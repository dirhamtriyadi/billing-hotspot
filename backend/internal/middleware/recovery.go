package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/gin-gonic/gin"
)

// Recovery converts an unhandled panic into a consistent 500 envelope and logs
// the stack trace, so a single bad request never takes the process down.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered",
					slog.Any("error", r),
					slog.String("request_id", GetRequestID(c)),
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
