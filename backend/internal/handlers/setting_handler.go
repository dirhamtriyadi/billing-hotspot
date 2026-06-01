package handlers

import (
	"net/http"

	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// SettingHandler exposes admin business-settings management.
type SettingHandler struct {
	svc *services.SettingService
}

// NewSettingHandler builds a SettingHandler.
func NewSettingHandler(svc *services.SettingService) *SettingHandler {
	return &SettingHandler{svc: svc}
}

// Get godoc
// @Summary  Get all settings
// @Tags     Settings
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} response.Envelope
// @Router   /settings [get]
func (h *SettingHandler) Get(c *gin.Context) {
	all, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", all)
}

// Update godoc
// @Summary  Update settings (key/value map)
// @Tags     Settings
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    payload body map[string]string true "Settings to upsert"
// @Success  200 {object} response.Envelope
// @Router   /settings [put]
func (h *SettingHandler) Update(c *gin.Context) {
	var values map[string]string
	if err := c.ShouldBindJSON(&values); err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Expected a JSON object of string values")
		return
	}
	all, err := h.svc.Update(c.Request.Context(), values)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Settings updated", all)
}
