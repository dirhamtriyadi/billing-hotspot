package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the canonical header carrying the correlation id.
const RequestIDHeader = "X-Request-ID"

// contextKeyRequestID is the gin context key under which the id is stored.
const contextKeyRequestID = "request_id"

// RequestID ensures every request has a correlation id, honouring an inbound
// X-Request-ID header or minting a new UUID, and echoes it on the response.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(contextKeyRequestID, id)
		c.Header(RequestIDHeader, id)
		c.Next()
	}
}

// GetRequestID returns the correlation id stored on the context.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(contextKeyRequestID); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
