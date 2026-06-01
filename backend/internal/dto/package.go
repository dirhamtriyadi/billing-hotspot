package dto

import (
	"fmt"

	"github.com/dirhamt/billing-hotspot/backend/internal/models"
)

// PackageRequest creates or replaces a package (admin). Used for both POST and
// PUT — a PUT carries the full desired state.
type PackageRequest struct {
	Name               string `json:"name" binding:"required,max=120"`
	Description        string `json:"description" binding:"max=1000"`
	Price              int64  `json:"price" binding:"gte=0"`
	RateDownKbps       int    `json:"rate_down_kbps" binding:"required,min=64"`
	RateUpKbps         int    `json:"rate_up_kbps" binding:"required,min=64"`
	BurstEnabled       bool   `json:"burst_enabled"`
	ValidityValue      int    `json:"validity_value" binding:"required,min=1"`
	ValidityUnit       string `json:"validity_unit" binding:"required,oneof=minute hour day month"`
	SessionTimeoutSecs int    `json:"session_timeout_secs" binding:"gte=0"`
	DataQuotaMB        int64  `json:"data_quota_mb" binding:"gte=0"`
	SimultaneousUse    int    `json:"simultaneous_use" binding:"required,min=1,max=100"`
	Highlight          bool   `json:"highlight"`
	BadgeText          string `json:"badge_text" binding:"max=40"`
	Color              string `json:"color" binding:"max=20"`
	Icon               string `json:"icon" binding:"max=40"`
	SortOrder          int    `json:"sort_order"`
	IsActive           *bool  `json:"is_active"`
}

// PublicPackage is the marketing-friendly projection shown on the landing page.
type PublicPackage struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Description   string  `json:"description"`
	Price         int64   `json:"price"`
	DownloadMbps  float64 `json:"download_mbps"`
	UploadMbps    float64 `json:"upload_mbps"`
	Validity      string  `json:"validity"`
	ValidityValue int     `json:"validity_value"`
	ValidityUnit  string  `json:"validity_unit"`
	DataQuotaMB   int64   `json:"data_quota_mb"`
	Highlight     bool    `json:"highlight"`
	BadgeText     string  `json:"badge_text"`
	Color         string  `json:"color"`
	Icon          string  `json:"icon"`
}

var validityUnitLabel = map[string]string{
	models.UnitMinute: "Menit",
	models.UnitHour:   "Jam",
	models.UnitDay:    "Hari",
	models.UnitMonth:  "Bulan",
}

// ToPublicPackage maps a package model into its public projection.
func ToPublicPackage(p models.Package) PublicPackage {
	return PublicPackage{
		ID:            p.ID,
		Name:          p.Name,
		Slug:          p.Slug,
		Description:   p.Description,
		Price:         p.Price,
		DownloadMbps:  float64(p.RateDownKbps) / 1000.0,
		UploadMbps:    float64(p.RateUpKbps) / 1000.0,
		Validity:      fmt.Sprintf("%d %s", p.ValidityValue, validityUnitLabel[p.ValidityUnit]),
		ValidityValue: p.ValidityValue,
		ValidityUnit:  p.ValidityUnit,
		DataQuotaMB:   p.DataQuotaMB,
		Highlight:     p.Highlight,
		BadgeText:     p.BadgeText,
		Color:         p.Color,
		Icon:          p.Icon,
	}
}

// ToPublicPackages maps a slice of packages.
func ToPublicPackages(in []models.Package) []PublicPackage {
	out := make([]PublicPackage, 0, len(in))
	for _, p := range in {
		out = append(out, ToPublicPackage(p))
	}
	return out
}
