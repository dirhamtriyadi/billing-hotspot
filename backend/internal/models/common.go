// Package models defines the GORM entities for the billing database.
// The schema itself is owned by the goose migrations under /migrations; these
// structs mirror it for type-safe queries.
package models

import (
	"time"

	"gorm.io/gorm"
)

// Base is embedded by every entity to provide an identity and audit timestamps.
type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Roles for admin users.
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
)

// Voucher lifecycle states.
const (
	VoucherUnused   = "unused"   // generated, not yet used
	VoucherActive   = "active"   // synced to radius and usable
	VoucherUsed     = "used"     // consumed (session-timeout / data exhausted)
	VoucherExpired  = "expired"  // past validity window
	VoucherDisabled = "disabled" // manually revoked
)

// Order payment methods / providers.
const (
	MethodCash     = "cash"
	MethodXendit   = "xendit"
	MethodMidtrans = "midtrans"
	MethodTripay   = "tripay"
)

// Order lifecycle states.
const (
	OrderPending   = "pending"
	OrderPaid      = "paid"
	OrderFailed    = "failed"
	OrderExpired   = "expired"
	OrderCancelled = "cancelled"
)

// Package validity units.
const (
	UnitMinute = "minute"
	UnitHour   = "hour"
	UnitDay    = "day"
	UnitMonth  = "month"
)
