package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/config"
)

// Xendit integrates the hosted Invoice API.
type Xendit struct {
	cfg config.XenditConfig
}

// NewXendit builds the Xendit gateway.
func NewXendit(cfg config.XenditConfig) *Xendit { return &Xendit{cfg: cfg} }

func (x *Xendit) Name() string     { return "xendit" }
func (x *Xendit) Configured() bool { return x.cfg.SecretKey != "" }

// Charge creates a Xendit invoice and returns its hosted checkout URL.
func (x *Xendit) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	payload := map[string]interface{}{
		"external_id":          req.OrderNumber,
		"amount":               req.Amount,
		"description":          req.Description,
		"currency":             "IDR",
		"payer_email":          guestEmail(req.CustomerEmail, req.OrderNumber),
		"success_redirect_url": req.SuccessRedirectURL,
		"failure_redirect_url": req.FailureRedirectURL,
		"customer": map[string]interface{}{
			"given_names":   firstNonEmpty(req.CustomerName, "Pelanggan"),
			"email":         guestEmail(req.CustomerEmail, req.OrderNumber),
			"mobile_number": req.CustomerPhone,
		},
	}

	headers := map[string]string{"Authorization": basicAuth(x.cfg.SecretKey, "")}
	status, body, err := postJSON(ctx, "https://api.xendit.co/v2/invoices", headers, payload)
	if err != nil {
		return nil, apperror.ServiceUnavailable("Xendit is unreachable").WithCause(err)
	}

	var out struct {
		ID         string `json:"id"`
		InvoiceURL string `json:"invoice_url"`
		Status     string `json:"status"`
		Message    string `json:"message"`
	}
	_ = json.Unmarshal(body, &out)

	if status >= 300 || out.InvoiceURL == "" {
		msg := "Xendit rejected the invoice"
		if out.Message != "" {
			msg = "Xendit: " + out.Message
		}
		return nil, apperror.Unprocessable(msg)
	}

	return &ChargeResult{
		Provider:   x.Name(),
		Reference:  out.ID,
		PaymentURL: out.InvoiceURL,
		Status:     StatusPending,
		Raw:        string(body),
	}, nil
}

// HandleWebhook validates the x-callback-token header and maps invoice status.
func (x *Xendit) HandleWebhook(headers http.Header, body []byte) (*WebhookResult, error) {
	token := headers.Get("X-Callback-Token")
	valid := x.cfg.CallbackToken != "" && secureEqual(token, x.cfg.CallbackToken)

	var n struct {
		ID         string `json:"id"`
		ExternalID string `json:"external_id"`
		Status     string `json:"status"`
	}
	if err := json.Unmarshal(body, &n); err != nil {
		return nil, apperror.BadRequest("Invalid Xendit callback payload")
	}

	return &WebhookResult{
		Provider:    x.Name(),
		Event:       n.Status,
		Reference:   n.ID,
		OrderNumber: n.ExternalID,
		Status:      mapXenditStatus(n.Status),
		Valid:       valid,
		Raw:         string(body),
	}, nil
}

func mapXenditStatus(status string) string {
	switch strings.ToUpper(status) {
	case "PAID", "SETTLED":
		return StatusPaid
	case "EXPIRED":
		return StatusExpired
	case "PENDING":
		return StatusPending
	default:
		return StatusFailed
	}
}
