package models

import "time"

// Package is a sellable internet plan. Its radius-facing attributes (speed,
// validity, quotas) are provisioned into FreeRADIUS as a profile/group when a
// voucher for this package is created. The presentation fields drive the
// attractive package-selection UI on the public landing page.
type Package struct {
	Base
	Name        string `gorm:"size:120;not null" json:"name"`
	Slug        string `gorm:"size:140;uniqueIndex;not null" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	Price       int64  `gorm:"not null" json:"price"` // IDR, whole rupiah

	// Radius profile name (FreeRADIUS group). Derived from slug.
	Profile string `gorm:"size:80;not null" json:"profile"`

	// Bandwidth (kbps). RateUp = client upload, RateDown = client download.
	RateDownKbps int  `gorm:"not null" json:"rate_down_kbps"`
	RateUpKbps   int  `gorm:"not null" json:"rate_up_kbps"`
	BurstEnabled bool `gorm:"not null;default:false" json:"burst_enabled"`

	// Validity window, e.g. 1 day. Combined into ValidityUnit/Value.
	ValidityValue int    `gorm:"not null;default:1" json:"validity_value"`
	ValidityUnit  string `gorm:"size:10;not null;default:day" json:"validity_unit"`

	// SessionTimeoutSecs caps total online time (0 = unlimited).
	SessionTimeoutSecs int `gorm:"not null;default:0" json:"session_timeout_secs"`
	// DataQuotaMB caps total transfer in megabytes (0 = unlimited).
	DataQuotaMB int64 `gorm:"not null;default:0" json:"data_quota_mb"`
	// SimultaneousUse caps concurrent logins for one voucher.
	SimultaneousUse int `gorm:"not null;default:1" json:"simultaneous_use"`

	// Presentation / merchandising for the package-selection UI.
	Highlight bool   `gorm:"not null;default:false" json:"highlight"`
	BadgeText string `gorm:"size:40" json:"badge_text"`
	Color     string `gorm:"size:20;default:'#2563eb'" json:"color"`
	Icon      string `gorm:"size:40;default:'wifi'" json:"icon"`
	SortOrder int    `gorm:"not null;default:0" json:"sort_order"`
	IsActive  bool   `gorm:"not null;default:true" json:"is_active"`
}

// Validity returns the package validity as a duration. Returns 0 for an
// unknown unit (treated as unlimited validity).
func (p Package) Validity() time.Duration {
	v := time.Duration(p.ValidityValue)
	switch p.ValidityUnit {
	case UnitMinute:
		return v * time.Minute
	case UnitHour:
		return v * time.Hour
	case UnitDay:
		return v * 24 * time.Hour
	case UnitMonth:
		return v * 30 * 24 * time.Hour
	default:
		return 0
	}
}
