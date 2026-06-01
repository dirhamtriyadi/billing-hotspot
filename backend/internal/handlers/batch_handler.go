package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/middleware"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// BatchHandler exposes voucher batch generation and listing.
type BatchHandler struct {
	svc *services.VoucherService
}

// NewBatchHandler builds a BatchHandler.
func NewBatchHandler(svc *services.VoucherService) *BatchHandler { return &BatchHandler{svc: svc} }

// Create godoc
// @Summary  Generate a batch of vouchers
// @Tags     Batches
// @Accept   json
// @Produce  json
// @Security BearerAuth
// @Param    payload body dto.CreateBatchRequest true "Batch"
// @Success  201 {object} response.Envelope{data=models.VoucherBatch}
// @Router   /batches [post]
func (h *BatchHandler) Create(c *gin.Context) {
	var req dto.CreateBatchRequest
	if !bindJSON(c, &req) {
		return
	}
	batch, err := h.svc.GenerateBatch(c.Request.Context(), req, middleware.CurrentUserID(c))
	if err != nil {
		fail(c, err)
		return
	}
	response.Created(c, "Vouchers generated", batch)
}

// List godoc
// @Summary  List voucher batches
// @Tags     Batches
// @Produce  json
// @Security BearerAuth
// @Param    page     query int    false "Page"
// @Param    per_page query int    false "Per page"
// @Param    search   query string false "Search by name"
// @Success  200 {object} response.Envelope{data=[]models.VoucherBatch,meta=response.Meta}
// @Router   /batches [get]
func (h *BatchHandler) List(c *gin.Context) {
	var q dto.PageQuery
	if !bindQuery(c, &q) {
		return
	}
	q.Normalize()
	batches, total, err := h.svc.ListBatches(c.Request.Context(), q)
	if err != nil {
		fail(c, err)
		return
	}
	response.Paginated(c, "OK", batches, response.NewMeta(q.Page, q.PerPage, total))
}

// Get godoc
// @Summary  Get a batch with its vouchers
// @Tags     Batches
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Batch ID"
// @Success  200 {object} response.Envelope{data=models.VoucherBatch}
// @Router   /batches/{id} [get]
func (h *BatchHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	batch, err := h.svc.GetBatch(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", batch)
}

// Delete godoc
// @Summary  Delete a batch and its vouchers
// @Tags     Batches
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Batch ID"
// @Success  200 {object} response.Envelope
// @Router   /batches/{id} [delete]
func (h *BatchHandler) Delete(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteBatch(c.Request.Context(), id); err != nil {
		fail(c, err)
		return
	}
	response.NoContent(c, "Batch deleted")
}
