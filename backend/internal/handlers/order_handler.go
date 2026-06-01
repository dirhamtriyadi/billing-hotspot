package handlers

import (
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// OrderHandler exposes admin order management.
type OrderHandler struct {
	svc *services.OrderService
}

// NewOrderHandler builds an OrderHandler.
func NewOrderHandler(svc *services.OrderService) *OrderHandler { return &OrderHandler{svc: svc} }

// List godoc
// @Summary  List orders
// @Tags     Orders
// @Produce  json
// @Security BearerAuth
// @Param    page           query int    false "Page"
// @Param    per_page       query int    false "Per page"
// @Param    search         query string false "Search by number/customer"
// @Param    status         query string false "Filter by status"
// @Param    payment_method query string false "Filter by payment method"
// @Success  200 {object} response.Envelope{data=[]models.Order,meta=response.Meta}
// @Router   /orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	var q dto.OrderListQuery
	if !bindQuery(c, &q) {
		return
	}
	q.Normalize()
	orders, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		fail(c, err)
		return
	}
	response.Paginated(c, "OK", orders, response.NewMeta(q.Page, q.PerPage, total))
}

// Get godoc
// @Summary  Get an order
// @Tags     Orders
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Order ID"
// @Success  200 {object} response.Envelope{data=models.Order}
// @Router   /orders/{id} [get]
func (h *OrderHandler) Get(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	order, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", order)
}

// ConfirmCash godoc
// @Summary  Confirm cash payment for an order (issues the voucher)
// @Tags     Orders
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Order ID"
// @Success  200 {object} response.Envelope{data=models.Order}
// @Router   /orders/{id}/confirm-cash [post]
func (h *OrderHandler) ConfirmCash(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	order, err := h.svc.ConfirmCash(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Payment confirmed, voucher issued", order)
}

// MarkPaid godoc
// @Summary  Manually settle a gateway order (webhook failed / payment page closed)
// @Tags     Orders
// @Produce  json
// @Security BearerAuth
// @Param    id path int true "Order ID"
// @Success  200 {object} response.Envelope{data=models.Order}
// @Router   /orders/{id}/mark-paid [post]
func (h *OrderHandler) MarkPaid(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	order, err := h.svc.MarkPaidManual(c.Request.Context(), id)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "Pesanan dilunaskan manual, voucher diterbitkan", order)
}
