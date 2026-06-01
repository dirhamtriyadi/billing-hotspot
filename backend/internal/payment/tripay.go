package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/config"
)

// Tripay integrates the closed-transaction API (hosted checkout page).
type Tripay struct {
	cfg config.TripayConfig
}

// NewTripay builds the Tripay gateway.
func NewTripay(cfg config.TripayConfig) *Tripay { return &Tripay{cfg: cfg} }

func (t *Tripay) Name() string { return "tripay" }
func (t *Tripay) Configured() bool {
	return t.cfg.APIKey != "" && t.cfg.PrivateKey != "" && t.cfg.MerchantCode != ""
}

func (t *Tripay) baseURL() string {
	if t.cfg.IsProduction {
		return "https://tripay.co.id/api"
	}
	return "https://tripay.co.id/api-sandbox"
}

// Charge creates a Tripay closed transaction and returns its checkout URL.
func (t *Tripay) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	channel := req.Channel
	if channel == "" {
		channel = "QRIS" // sensible default; admin can expose more channels later
	}
	amount := strconv.FormatInt(req.Amount, 10)
	signature := hmacSHA256Hex(t.cfg.MerchantCode+req.OrderNumber+amount, t.cfg.PrivateKey)

	payload := map[string]interface{}{
		"method":         channel,
		"merchant_ref":   req.OrderNumber,
		"amount":         req.Amount,
		"customer_name":  firstNonEmpty(req.CustomerName, "Pelanggan"),
		"customer_email": guestEmail(req.CustomerEmail, req.OrderNumber),
		"customer_phone": req.CustomerPhone,
		"order_items": []map[string]interface{}{{
			"sku":      req.OrderNumber,
			"name":     req.Description,
			"price":    req.Amount,
			"quantity": 1,
		}},
		"callback_url": req.CallbackURL,
		"return_url":   req.SuccessRedirectURL,
		"signature":    signature,
	}

	headers := map[string]string{"Authorization": "Bearer " + t.cfg.APIKey}
	status, body, err := postJSON(ctx, t.baseURL()+"/transaction/create", headers, payload)
	if err != nil {
		return nil, apperror.ServiceUnavailable("Tripay is unreachable").WithCause(err)
	}

	var out struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Reference   string `json:"reference"`
			CheckoutURL string `json:"checkout_url"`
			QRString    string `json:"qr_string"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &out)

	if status >= 300 || !out.Success || out.Data.CheckoutURL == "" {
		msg := "Tripay rejected the transaction"
		if out.Message != "" {
			msg = "Tripay: " + out.Message
		}
		return nil, apperror.Unprocessable(msg)
	}

	return &ChargeResult{
		Provider:   t.Name(),
		Reference:  out.Data.Reference,
		PaymentURL: out.Data.CheckoutURL,
		QRString:   out.Data.QRString,
		Status:     StatusPending,
		Raw:        string(body),
	}, nil
}

// HandleWebhook validates the HMAC callback signature and maps status.
func (t *Tripay) HandleWebhook(headers http.Header, body []byte) (*WebhookResult, error) {
	signature := headers.Get("X-Callback-Signature")
	expected := hmacSHA256Hex(string(body), t.cfg.PrivateKey)
	valid := secureEqual(signature, expected)

	var n struct {
		Reference   string `json:"reference"`
		MerchantRef string `json:"merchant_ref"`
		Status      string `json:"status"`
	}
	if err := json.Unmarshal(body, &n); err != nil {
		return nil, apperror.BadRequest("Invalid Tripay callback payload")
	}

	return &WebhookResult{
		Provider:    t.Name(),
		Event:       headers.Get("X-Callback-Event"),
		Reference:   n.Reference,
		OrderNumber: n.MerchantRef,
		Status:      mapTripayStatus(n.Status),
		Valid:       valid,
		Raw:         string(body),
	}, nil
}

func mapTripayStatus(status string) string {
	switch strings.ToUpper(status) {
	case "PAID":
		return StatusPaid
	case "EXPIRED":
		return StatusExpired
	case "FAILED", "REFUND":
		return StatusFailed
	default:
		return StatusPending
	}
}
