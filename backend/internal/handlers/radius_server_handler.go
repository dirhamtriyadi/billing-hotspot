package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type RadiusServerHandler struct {
	svc *services.RadiusServerService
}

func NewRadiusServerHandler(svc *services.RadiusServerService) *RadiusServerHandler {
	return &RadiusServerHandler{svc: svc}
}

func (h *RadiusServerHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", list)
}

func (h *RadiusServerHandler) Create(c *gin.Context) {
	var in dto.RadiusServerInput
	if !bindJSON(c, &in) {
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Radius server disimpan", out)
}

func (h *RadiusServerHandler) Update(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var in dto.RadiusServerInput
	if !bindJSON(c, &in) {
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Radius server diperbarui", out)
}

func (h *RadiusServerHandler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}
	response.NoContent(c, "Radius server dihapus")
}
