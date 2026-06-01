package handlers

import (
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/payment"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// PublicHandler exposes the unauthenticated storefront endpoints consumed by the
// landing page and checkout flow.
type PublicHandler struct {
	packages *services.PackageService
	orders   *services.OrderService
	settings *services.SettingService
	payments *payment.Registry
}

// NewPublicHandler builds a PublicHandler.
func NewPublicHandler(p *services.PackageService, o *services.OrderService, s *services.SettingService, pay *payment.Registry) *PublicHandler {
	return &PublicHandler{packages: p, orders: o, settings: s, payments: pay}
}

// ListPackages godoc
// @Summary  List active packages (public)
// @Tags     Public
// @Produce  json
// @Success  200 {object} response.Envelope{data=[]dto.PublicPackage}
// @Router   /public/packages [get]
func (h *PublicHandler) ListPackages(c *gin.Context) {
	packages, err := h.packages.ListPublic(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", dto.ToPublicPackages(packages))
}

// GetPackage godoc
// @Summary  Get an active package by slug (public)
// @Tags     Public
// @Produce  json
// @Param    slug path string true "Package slug"
// @Success  200 {object} response.Envelope{data=dto.PublicPackage}
// @Router   /public/packages/{slug} [get]
func (h *PublicHandler) GetPackage(c *gin.Context) {
	p, err := h.packages.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", dto.ToPublicPackage(*p))
}

// Checkout godoc
// @Summary  Create an order and start payment (public)
// @Tags     Public
// @Accept   json
// @Produce  json
// @Param    payload body dto.CheckoutRequest true "Checkout"
// @Success  201 {object} response.Envelope{data=models.Order}
// @Failure  422 {object} response.Envelope
// @Router   /public/checkout [post]
func (h *PublicHandler) Checkout(c *gin.Context) {
	var req dto.CheckoutRequest
	if !bindJSON(c, &req) {
		return
	}
	order, err := h.orders.Checkout(c.Request.Context(), req)
	if err != nil {
		fail(c, err)
		return
	}
	response.Created(c, "Order created", order)
}

// OrderStatus godoc
// @Summary  Check an order's status by order number (public)
// @Tags     Public
// @Produce  json
// @Param    number path string true "Order number"
// @Success  200 {object} response.Envelope{data=models.Order}
// @Router   /public/orders/{number} [get]
func (h *PublicHandler) OrderStatus(c *gin.Context) {
	order, err := h.orders.GetByNumber(c.Request.Context(), c.Param("number"))
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", order)
}

// Settings godoc
// @Summary  Public storefront settings + available payment methods
// @Tags     Public
// @Produce  json
// @Success  200 {object} response.Envelope
// @Router   /public/settings [get]
func (h *PublicHandler) Settings(c *gin.Context) {
	all, err := h.settings.GetAll(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}

	methods := h.payments.Available()
	if strings.EqualFold(all["enable_cash"], "true") {
		methods = append([]string{"cash"}, methods...)
	}

	response.OK(c, "OK", gin.H{
		"site_name":        all["site_name"],
		"site_subtitle":    all["site_subtitle"],
		"site_description": all["site_description"],
		"contact_whatsapp": all["contact_whatsapp"],
		"currency":         all["currency"],
		"payment_methods":  methods,
	})
}
