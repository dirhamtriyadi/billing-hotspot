package middleware

import (
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/token"
	"github.com/gin-gonic/gin"
)

// Context keys for authenticated identity.
const (
	ContextUserID   = "auth_user_id"
	ContextUsername = "auth_username"
	ContextRole     = "auth_role"
)

// Auth validates the Bearer JWT and stores the caller's identity on the context.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, 401, "UNAUTHORIZED", "Missing or malformed Authorization header")
			c.Abort()
			return
		}

		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := token.Parse(secret, raw)
		if err != nil {
			response.Error(c, 401, "UNAUTHORIZED", "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// CurrentUserID returns the authenticated user id, or 0 if unauthenticated.
func CurrentUserID(c *gin.Context) uint {
	if v, ok := c.Get(ContextUserID); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}
