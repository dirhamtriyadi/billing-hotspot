package handlers

import (
	"net/http"

	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// WebhookHandler receives payment gateway notifications. Routes are public but
// each provider's signature/token is verified inside the service.
type WebhookHandler struct {
	orders *services.OrderService
}

// NewWebhookHandler builds a WebhookHandler.
func NewWebhookHandler(orders *services.OrderService) *WebhookHandler {
	return &WebhookHandler{orders: orders}
}

// Handle godoc
// @Summary  Payment gateway webhook/callback
// @Tags     Webhooks
// @Accept   json
// @Produce  json
// @Param    provider path string true "Provider" Enums(xendit, midtrans, tripay)
// @Success  200 {object} response.Envelope
// @Failure  401 {object} response.Envelope
// @Router   /webhooks/{provider} [post]
func (h *WebhookHandler) Handle(c *gin.Context) {
	provider := c.Param("provider")
	body, err := c.GetRawData()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "Unable to read request body")
		return
	}
	if err := h.orders.HandleWebhook(c.Request.Context(), provider, c.Request.Header, body); err != nil {
		fail(c, err)
		return
	}
	// A 200 with {"success": true} satisfies all three providers' ack contracts.
	response.OK(c, "Webhook processed", nil)
}
