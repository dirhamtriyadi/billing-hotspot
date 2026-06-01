package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// DashboardHandler exposes the operator dashboard summary.
type DashboardHandler struct {
	svc *services.DashboardService
}

// NewDashboardHandler builds a DashboardHandler.
func NewDashboardHandler(svc *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

// Stats godoc
// @Summary  Dashboard statistics
// @Tags     Dashboard
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} response.Envelope{data=dto.DashboardStats}
// @Router   /dashboard/stats [get]
func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", stats)
}
