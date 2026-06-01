// Package server wires the radius-api Gin engine.
package server

import (
	"net/http"

	_ "github.com/dirhamt/billing-hotspot/radius-api/docs" // generated swagger spec
	"github.com/dirhamt/billing-hotspot/radius-api/internal/coa"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/config"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/handlers"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/middleware"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/radiusreload"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/radiussql"
	"github.com/dirhamt/billing-hotspot/radius-api/internal/response"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// New builds the radius-api engine.
func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(cfg.CORS),
	)

	disconnector := coa.New(cfg.CoA.Port, cfg.CoA.Timeout)
	svc := radiussql.New(db, disconnector)
	reloader := radiusreload.New(cfg.Reload.Container, cfg.Reload.DockerSocket)
	h := handlers.New(svc, reloader)

	r.GET("/health", func(c *gin.Context) {
		response.OK(c, "ok", gin.H{"status": "healthy", "service": cfg.App.Name})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	api.Use(middleware.APIKey(cfg.Auth.APIKey))
	{
		api.POST("/profiles", h.UpsertProfile)

		api.POST("/users", h.CreateUser)
		api.POST("/users/bulk", h.CreateUsers)
		api.GET("/users/:username", h.GetUser)
		api.DELETE("/users/:username", h.DeleteUser)
		api.POST("/users/:username/disconnect", h.DisconnectUser)

		api.GET("/sessions", h.ListSessions)

		api.GET("/nas", h.ListNAS)
		api.POST("/nas", h.UpsertNAS)
		api.DELETE("/nas/:id", h.DeleteNAS)
	}

	r.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "Route not found")
	})

	return r
}
