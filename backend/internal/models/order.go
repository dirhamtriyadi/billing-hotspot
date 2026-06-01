package models

import "time"

// Order is a single self-service purchase of a package. When paid (cash
// confirmation or gateway webhook) the system mints a Voucher and provisions it
// into FreeRADIUS.
type Order struct {
	Base
	OrderNumber string  `gorm:"size:40;uniqueIndex;not null" json:"order_number"`
	PackageID   uint    `gorm:"not null;index" json:"package_id"`
	Package     Package `gorm:"constraint:OnDelete:RESTRICT" json:"package,omitempty"`

	CustomerName  string `gorm:"size:120" json:"customer_name"`
	CustomerPhone string `gorm:"size:30" json:"customer_phone"`
	CustomerEmail string `gorm:"size:160" json:"customer_email"`

	Amount        int64  `gorm:"not null" json:"amount"`
	PaymentMethod string `gorm:"size:20;not null;index" json:"payment_method"` // cash|xendit|midtrans|tripay
	Status        string `gorm:"size:20;not null;default:pending;index" json:"status"`

	// Gateway artefacts (empty for cash).
	Reference    string `gorm:"size:120;index" json:"reference"` // gateway invoice/transaction id
	PaymentURL   string `gorm:"type:text" json:"payment_url"`    // hosted checkout / redirect URL
	PaymentToken string `gorm:"size:255" json:"payment_token"`   // e.g. Midtrans Snap token
	QRString     string `gorm:"type:text" json:"qr_string,omitempty"`
	RawResponse  string `gorm:"type:text" json:"-"` // raw gateway create response (debugging)

	PaidAt    *time.Time `json:"paid_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	VoucherID *uint    `gorm:"index" json:"voucher_id,omitempty"`
	Voucher   *Voucher `json:"voucher,omitempty"`
}
