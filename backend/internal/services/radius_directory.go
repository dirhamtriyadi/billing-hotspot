package services

import (
	"context"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/radius"
	"gorm.io/gorm"
)

// RadiusEndpoint is one branch-local radius-api target.
type RadiusEndpoint struct {
	Name   string
	URL    string
	Client *radius.Client
}

// RadiusDirectory resolves all radius-api endpoints configured in NAS records.
// Empty per-NAS values fall back to the legacy single RADIUS_API_URL/API_KEY.
type RadiusDirectory struct {
	db  *gorm.DB
	cfg config.RadiusConfig
}

func NewRadiusDirectory(db *gorm.DB, cfg config.RadiusConfig) *RadiusDirectory {
	return &RadiusDirectory{db: db, cfg: cfg}
}

// Endpoints returns unique radius-api targets. If no NAS records exist yet, the
// legacy default endpoint is returned so older single-RADIUS installs keep
// working.
func (d *RadiusDirectory) Endpoints(ctx context.Context) ([]RadiusEndpoint, error) {
	var rows []models.NASHotspotConfig
	if err := d.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, mapDBError(err)
	}

	seen := map[string]struct{}{}
	out := make([]RadiusEndpoint, 0, len(rows)+1)
	for _, row := range rows {
		url := strings.TrimRight(defaultString(row.RadiusAPIURL, d.cfg.BaseURL), "/")
		key := defaultString(row.RadiusAPIKey, d.cfg.APIKey)
		if url == "" || key == "" {
			continue
		}
		dedupeKey := url + "\x00" + key
		if _, ok := seen[dedupeKey]; ok {
			continue
		}
		seen[dedupeKey] = struct{}{}
		out = append(out, RadiusEndpoint{
			Name:   defaultString(row.ShortName, row.NASName),
			URL:    url,
			Client: radius.NewClientWith(url, key, d.cfg.Timeout),
		})
	}

	if len(out) == 0 && d.cfg.BaseURL != "" && d.cfg.APIKey != "" {
		url := strings.TrimRight(d.cfg.BaseURL, "/")
		out = append(out, RadiusEndpoint{
			Name:   "default",
			URL:    url,
			Client: radius.NewClientWith(url, d.cfg.APIKey, d.cfg.Timeout),
		})
	}

	return out, nil
}
