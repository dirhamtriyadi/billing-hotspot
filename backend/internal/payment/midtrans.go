package payment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/config"
)

// Midtrans integrates the Snap hosted-checkout API.
type Midtrans struct {
	cfg config.MidtransConfig
}

// NewMidtrans builds the Midtrans gateway.
func NewMidtrans(cfg config.MidtransConfig) *Midtrans { return &Midtrans{cfg: cfg} }

func (m *Midtrans) Name() string     { return "midtrans" }
func (m *Midtrans) Configured() bool { return m.cfg.ServerKey != "" }

func (m *Midtrans) snapBaseURL() string {
	if m.cfg.IsProduction {
		return "https://app.midtrans.com"
	}
	return "https://app.sandbox.midtrans.com"
}

// Charge creates a Snap transaction and returns its token + redirect URL.
func (m *Midtrans) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	itemName := req.Description
	if len(itemName) > 50 {
		itemName = itemName[:50] // Midtrans item name limit
	}
	payload := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     req.OrderNumber,
			"gross_amount": req.Amount,
		},
		"customer_details": map[string]interface{}{
			"first_name": firstNonEmpty(req.CustomerName, "Pelanggan"),
			"email":      guestEmail(req.CustomerEmail, req.OrderNumber),
			"phone":      req.CustomerPhone,
		},
		"item_details": []map[string]interface{}{{
			"id":       req.OrderNumber,
			"price":    req.Amount,
			"quantity": 1,
			"name":     itemName,
		}},
		"callbacks": map[string]interface{}{
			"finish": req.SuccessRedirectURL,
		},
	}

	headers := map[string]string{"Authorization": basicAuth(m.cfg.ServerKey, "")}
	status, body, err := postJSON(ctx, m.snapBaseURL()+"/snap/v1/transactions", headers, payload)
	if err != nil {
		return nil, apperror.ServiceUnavailable("Midtrans is unreachable").WithCause(err)
	}

	var out struct {
		Token         string   `json:"token"`
		RedirectURL   string   `json:"redirect_url"`
		ErrorMessages []string `json:"error_messages"`
	}
	_ = json.Unmarshal(body, &out)

	if status >= 300 || out.Token == "" {
		msg := "Midtrans rejected the transaction"
		if len(out.ErrorMessages) > 0 {
			msg = "Midtrans: " + out.ErrorMessages[0]
		}
		return nil, apperror.Unprocessable(msg).WithCause(nil)
	}

	return &ChargeResult{
		Provider:     m.Name(),
		Reference:    req.OrderNumber, // transaction_id only arrives via webhook
		PaymentURL:   out.RedirectURL,
		PaymentToken: out.Token,
		Status:       StatusPending,
		Raw:          string(body),
	}, nil
}

// HandleWebhook verifies the Midtrans notification signature and maps status.
func (m *Midtrans) HandleWebhook(_ http.Header, body []byte) (*WebhookResult, error) {
	var n struct {
		OrderID           string `json:"order_id"`
		StatusCode        string `json:"status_code"`
		GrossAmount       string `json:"gross_amount"`
		SignatureKey      string `json:"signature_key"`
		TransactionStatus string `json:"transaction_status"`
		FraudStatus       string `json:"fraud_status"`
		TransactionID     string `json:"transaction_id"`
	}
	if err := json.Unmarshal(body, &n); err != nil {
		return nil, apperror.BadRequest("Invalid Midtrans notification payload")
	}

	expected := sha512Hex(n.OrderID + n.StatusCode + n.GrossAmount + m.cfg.ServerKey)
	valid := secureEqual(expected, n.SignatureKey)

	return &WebhookResult{
		Provider:    m.Name(),
		Event:       n.TransactionStatus,
		Reference:   firstNonEmpty(n.TransactionID, n.OrderID),
		OrderNumber: n.OrderID,
		Status:      mapMidtransStatus(n.TransactionStatus, n.FraudStatus),
		Valid:       valid,
		Raw:         string(body),
	}, nil
}

func mapMidtransStatus(transactionStatus, fraudStatus string) string {
	switch transactionStatus {
	case "settlement":
		return StatusPaid
	case "capture":
		if fraudStatus == "accept" || fraudStatus == "" {
			return StatusPaid
		}
		return StatusPending // challenge -> awaiting manual review
	case "pending":
		return StatusPending
	case "expire":
		return StatusExpired
	case "deny", "cancel", "failure", "refund", "partial_refund", "chargeback":
		return StatusFailed
	default:
		return StatusPending
	}
}
