// Package server wires configuration, the database, services and handlers into
// a ready-to-run Gin engine with all routes and middleware registered.
package server

import (
	"context"
	"log/slog"
	"net/http"

	_ "github.com/dirhamt/billing-hotspot/backend/docs" // generated swagger spec
	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/dirhamt/billing-hotspot/backend/internal/handlers"
	"github.com/dirhamt/billing-hotspot/backend/internal/middleware"
	"github.com/dirhamt/billing-hotspot/backend/internal/payment"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// New builds the fully-wired Gin engine.
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

	// External clients and services.
	radiusDirectory := services.NewRadiusDirectory(db, cfg.Radius)
	paymentRegistry := payment.NewRegistry(cfg.Payment)

	authSvc := services.NewAuthService(db, cfg.JWT)
	pkgSvc := services.NewPackageService(db, radiusDirectory)
	voucherSvc := services.NewVoucherService(db, radiusDirectory)
	orderSvc := services.NewOrderService(db, paymentRegistry, voucherSvc, cfg.App)
	dashSvc := services.NewDashboardService(db)
	settingSvc := services.NewSettingService(db)
	gatewaySvc := services.NewGatewayService(settingSvc, cfg.Payment, paymentRegistry)
	nasSvc := services.NewNasService(db, cfg.Radius)
	reportSvc := services.NewReportService(db)

	// Apply any DB-stored gateway credentials on top of the env defaults so the
	// registry reflects what operators configured from the admin UI.
	if err := gatewaySvc.Reload(context.Background()); err != nil {
		slog.Warn("could not load gateway settings; using environment defaults", slog.Any("error", err))
	}

	// Handlers.
	authH := handlers.NewAuthHandler(authSvc)
	pkgH := handlers.NewPackageHandler(pkgSvc)
	voucherH := handlers.NewVoucherHandler(voucherSvc)
	batchH := handlers.NewBatchHandler(voucherSvc)
	orderH := handlers.NewOrderHandler(orderSvc)
	dashH := handlers.NewDashboardHandler(dashSvc)
	settingH := handlers.NewSettingHandler(settingSvc)
	gatewayH := handlers.NewGatewayHandler(gatewaySvc)
	nasH := handlers.NewNasHandler(nasSvc)
	reportH := handlers.NewReportHandler(reportSvc)
	hotspotH := handlers.NewHotspotHandler(cfg.App)
	publicH := handlers.NewPublicHandler(pkgSvc, orderSvc, settingSvc, paymentRegistry)
	webhookH := handlers.NewWebhookHandler(orderSvc)

	// Health + API docs.
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, "ok", gin.H{"status": "healthy", "service": cfg.App.Name})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")

	// --- Public (unauthenticated) ---
	pub := api.Group("/public")
	{
		pub.GET("/packages", publicH.ListPackages)
		pub.GET("/packages/:slug", publicH.GetPackage)
		pub.POST("/checkout", publicH.Checkout)
		pub.GET("/orders/:number", publicH.OrderStatus)
		pub.GET("/settings", publicH.Settings)
		pub.GET("/hotspot/login.html", hotspotH.LoginPage)
	}

	// Payment gateway callbacks (signature-verified inside the service).
	// Configure the gateway dashboard notification URL as:
	//   {APP_BASE_URL}/api/v1/webhooks/{provider}
	api.POST("/webhooks/:provider", webhookH.Handle)

	// Unauthenticated auth endpoints.
	api.POST("/auth/login", authH.Login)

	// --- Protected (JWT) ---
	authed := api.Group("")
	authed.Use(middleware.Auth(cfg.JWT.Secret))
	{
		authed.GET("/auth/me", authH.Me)
		authed.POST("/auth/change-password", authH.ChangePassword)

		authed.GET("/dashboard/stats", dashH.Stats)

		authed.GET("/reports/revenue", reportH.Revenue)
		authed.GET("/reports/export", reportH.Export)

		authed.GET("/nas", nasH.List)
		authed.POST("/nas", nasH.Upsert)
		authed.DELETE("/nas/:id", nasH.Delete)

		authed.GET("/packages", pkgH.List)
		authed.POST("/packages", pkgH.Create)
		authed.GET("/packages/:id", pkgH.Get)
		authed.PUT("/packages/:id", pkgH.Update)
		authed.DELETE("/packages/:id", pkgH.Delete)

		authed.GET("/vouchers", voucherH.List)
		authed.GET("/vouchers/:id", voucherH.Get)
		authed.PATCH("/vouchers/:id/status", voucherH.UpdateStatus)
		authed.DELETE("/vouchers/:id", voucherH.Delete)

		authed.POST("/batches", batchH.Create)
		authed.GET("/batches", batchH.List)
		authed.GET("/batches/:id", batchH.Get)
		authed.DELETE("/batches/:id", batchH.Delete)

		authed.GET("/orders", orderH.List)
		authed.GET("/orders/:id", orderH.Get)
		authed.POST("/orders/:id/confirm-cash", orderH.ConfirmCash)
		authed.POST("/orders/:id/mark-paid", orderH.MarkPaid)

		authed.GET("/settings", settingH.Get)
		authed.PUT("/settings", settingH.Update)

		authed.GET("/payment-gateways", gatewayH.Get)
		authed.PUT("/payment-gateways", gatewayH.Update)
	}

	r.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "Route not found")
	})

	return r
}
