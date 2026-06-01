// Package payment abstracts the supported payment providers (Xendit, Midtrans,
// Tripay) behind a single Gateway interface. Each provider is implemented with
// plain net/http calls to its REST API so there is no SDK version coupling.
package payment

import (
	"context"
	"net/http"
)

// Normalised payment statuses used across providers.
const (
	StatusPending = "pending"
	StatusPaid    = "paid"
	StatusFailed  = "failed"
	StatusExpired = "expired"
)

// ChargeRequest is the provider-agnostic input to create a payment.
type ChargeRequest struct {
	OrderNumber   string
	Amount        int64 // whole IDR
	Description   string
	CustomerName  string
	CustomerEmail string
	CustomerPhone string

	// Channel is an optional provider-specific payment channel (used by Tripay,
	// e.g. "QRIS", "BRIVA"). Ignored by hosted-checkout providers.
	Channel string

	SuccessRedirectURL string
	FailureRedirectURL string
	CallbackURL        string
}

// ChargeResult is the provider-agnostic output of creating a payment.
type ChargeResult struct {
	Provider     string
	Reference    string // gateway transaction/invoice id
	PaymentURL   string // hosted checkout / redirect URL
	PaymentToken string // e.g. Midtrans Snap token (for embedded checkout)
	QRString     string // optional raw QRIS payload
	Status       string // normalised, usually StatusPending
	Raw          string // raw JSON response, persisted for debugging
}

// WebhookResult is the parsed + verified outcome of an inbound notification.
type WebhookResult struct {
	Provider    string
	Event       string
	Reference   string
	OrderNumber string // our order_number (external_id / merchant_ref / order_id)
	Status      string // normalised status
	Valid       bool   // signature/token verification result
	Raw         string
}

// Gateway is implemented by every payment provider.
type Gateway interface {
	// Name returns the provider identifier ("xendit"|"midtrans"|"tripay").
	Name() string
	// Configured reports whether the provider has the credentials it needs.
	Configured() bool
	// Charge creates a payment and returns checkout details.
	Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error)
	// HandleWebhook verifies and parses an inbound notification.
	HandleWebhook(headers http.Header, body []byte) (*WebhookResult, error)
}
