package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// GatewayHandler exposes admin payment-gateway credential management.
type GatewayHandler struct {
	svc *services.GatewayService
}

// NewGatewayHandler builds a GatewayHandler.
func NewGatewayHandler(svc *services.GatewayService) *GatewayHandler {
	return &GatewayHandler{svc: svc}
}

// Get godoc
// @Summary  Get payment gateway settings (secrets masked)
// @Tags     Payment Gateways
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} response.Envelope{data=dto.GatewaySettings}
// @Router   /payment-gateways [get]
func (h *GatewayHandler) Get(c *gin.Context) {
	cfg, err := h.svc.Get(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", cfg)
}

// Update godoc
// @Summary  Update payment gateway credentials
// @Tags     Payment Gateways
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    payload body dto.GatewayUpdate true "Gateway settings"
// @Success  200 {object} response.Envelope{data=dto.GatewaySettings}
// @Router   /payment-gateways [put]
func (h *GatewayHandler) Update(c *gin.Context) {
	var in dto.GatewayUpdate
	if !bindJSON(c, &in) {
		return
	}
	cfg, err := h.svc.Update(c.Request.Context(), in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Pengaturan gateway disimpan", cfg)
}
