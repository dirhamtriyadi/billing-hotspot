package models

import "time"

// Voucher is a single hotspot credential. Per the chosen login style the Code
// doubles as both username and password in FreeRADIUS. A voucher is created
// either by an admin batch or by a paid self-service order.
type Voucher struct {
	Base
	Code      string  `gorm:"size:40;uniqueIndex;not null" json:"code"`
	PackageID uint    `gorm:"not null;index" json:"package_id"`
	Package   Package `gorm:"constraint:OnDelete:RESTRICT" json:"package,omitempty"`

	BatchID *uint `gorm:"index" json:"batch_id,omitempty"`
	OrderID *uint `gorm:"index" json:"order_id,omitempty"`

	Status  string `gorm:"size:20;not null;default:unused;index" json:"status"`
	Profile string `gorm:"size:80;not null" json:"profile"` // snapshot of package profile
	Price   int64  `gorm:"not null" json:"price"`           // snapshot of package price

	SyncedToRadius bool `gorm:"not null;default:false" json:"synced_to_radius"`

	ActivatedAt *time.Time `json:"activated_at,omitempty"` // first login (best-effort)
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`   // validity deadline
	UsedAt      *time.Time `json:"used_at,omitempty"`

	Note string `gorm:"size:255" json:"note"`
}
