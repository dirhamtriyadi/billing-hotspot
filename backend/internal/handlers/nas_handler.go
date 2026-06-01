package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// NasHandler exposes admin NAS / RADIUS client management.
type NasHandler struct {
	svc *services.NasService
}

// NewNasHandler builds a NasHandler.
func NewNasHandler(svc *services.NasService) *NasHandler {
	return &NasHandler{svc: svc}
}

// List godoc
// @Summary  List NAS / RADIUS clients
// @Tags     NAS
// @Produce  json
// @Security BearerAuth
// @Success  200 {object} response.Envelope{data=[]radius.NAS}
// @Router   /nas [get]
func (h *NasHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", list)
}

// Upsert godoc
// @Summary  Register or update a NAS / RADIUS client
// @Tags     NAS
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    payload body dto.NASInput true "NAS"
// @Success  200 {object} response.Envelope{data=radius.NAS}
// @Router   /nas [post]
func (h *NasHandler) Upsert(c *gin.Context) {
	var in dto.NASInput
	if !bindJSON(c, &in) {
		return
	}
	nas, err := h.svc.Upsert(c.Request.Context(), in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Router disimpan", nas)
}

// Delete godoc
// @Summary  Remove a NAS / RADIUS client
// @Tags     NAS
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "NAS ID"
// @Success  200 {object} response.Envelope
// @Router   /nas/{id} [delete]
func (h *NasHandler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}
	response.NoContent(c, "Router dihapus")
}
