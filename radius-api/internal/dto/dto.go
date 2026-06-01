// Package dto holds the radius-api request/response payloads.
package dto

import (
	"time"

	"github.com/dirhamt/billing-hotspot/radius-api/internal/models"
)

// ProfileRequest upserts a FreeRADIUS group with bandwidth/quota attributes.
type ProfileRequest struct {
	Name            string `json:"name" binding:"required"`
	RateDownKbps    int    `json:"rate_down_kbps"`
	RateUpKbps      int    `json:"rate_up_kbps"`
	BurstEnabled    bool   `json:"burst_enabled"`
	SessionTimeout  int    `json:"session_timeout_secs"`
	DataQuotaMB     int64  `json:"data_quota_mb"`
	SimultaneousUse int    `json:"simultaneous_use"`
}

// UserRequest provisions a single credential.
type UserRequest struct {
	Username  string     `json:"username" binding:"required"`
	Password  string     `json:"password" binding:"required"`
	Profile   string     `json:"profile" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// BulkUsersRequest provisions many credentials at once.
type BulkUsersRequest struct {
	Users []UserRequest `json:"users" binding:"required,min=1,dive"`
}

// NASRequest registers (or updates) a NAS / RADIUS client.
type NASRequest struct {
	NASName     string `json:"nasname" binding:"required"` // IP address or hostname
	ShortName   string `json:"shortname"`
	Secret      string `json:"secret" binding:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Ports       *int   `json:"ports"`
}

// UserDetail is the read view of a provisioned credential.
type UserDetail struct {
	Username   string            `json:"username"`
	Groups     []string          `json:"groups"`
	Attributes []models.RadCheck `json:"attributes"`
}
