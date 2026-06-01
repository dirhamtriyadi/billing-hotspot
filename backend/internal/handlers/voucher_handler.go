package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// VoucherHandler exposes admin voucher management.
type VoucherHandler struct {
	svc *services.VoucherService
}

// NewVoucherHandler builds a VoucherHandler.
func NewVoucherHandler(svc *services.VoucherService) *VoucherHandler {
	return &VoucherHandler{svc: svc}
}

// List godoc
// @Summary  List vouchers
// @Tags     Vouchers
// @Produce  json
// @Security BearerAuth
// @Param    page       query int    false "Page"
// @Param    per_page   query int    false "Per page"
// @Param    search     query string false "Search by code"
// @Param    status     query string false "Filter by status"
// @Param    package_id query int    false "Filter by package"
// @Param    batch_id   query int    false "Filter by batch"
// @Success  200 {object} response.Envelope{data=[]models.Voucher,meta=response.Meta}
// @Router   /vouchers [get]
func (h *VoucherHandler) List(c *gin.Context) {
	var q dto.VoucherListQuery
	if !bindQuery(c, &q) {
		return
	}
	q.Normalize()
	vouchers, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		fail(c, err)
		return
	}
	response.Paginated(c, "OK", vouchers, response.NewMeta(q.Page, q.PerPage, total))
}

// Get godoc
// @Summary  Get a voucher
// @Tags     Vouchers
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Voucher ID"
// @Success  200 {object} response.Envelope{data=models.Voucher}
// @Router   /vouchers/{id} [get]
func (h *VoucherHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	v, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", v)
}

// UpdateStatus godoc
// @Summary  Enable or disable a voucher
// @Tags     Vouchers
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Voucher ID"
// @Param    payload body dto.UpdateVoucherStatusRequest true "Status"
// @Success  200 {object} response.Envelope{data=models.Voucher}
// @Router   /vouchers/{id}/status [patch]
func (h *VoucherHandler) UpdateStatus(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var req dto.UpdateVoucherStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	v, err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Voucher updated", v)
}

// Delete godoc
// @Summary  Delete a voucher
// @Tags     Vouchers
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Voucher ID"
// @Success  200 {object} response.Envelope
// @Router   /vouchers/{id} [delete]
func (h *VoucherHandler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}
	response.NoContent(c, "Voucher deleted")
}
