package middleware

import (
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS configures cross-origin access for the SPA frontend.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", RequestIDHeader},
		ExposeHeaders:    []string{"Content-Length", RequestIDHeader},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
