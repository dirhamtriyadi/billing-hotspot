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

// RadiusDirectory resolves all radius-api endpoints configured in Radius Server
// master data and NAS records. Empty per-NAS values fall back to the legacy
// single RADIUS_API_URL/API_KEY.
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
	seen := map[string]struct{}{}
	out := make([]RadiusEndpoint, 0)

	var servers []models.RadiusServer
	if err := d.db.WithContext(ctx).Find(&servers).Error; err != nil {
		return nil, mapDBError(err)
	}
	for _, srv := range servers {
		d.addEndpoint(&out, seen, srv.Name, srv.APIURL, srv.APIKey)
	}

	var rows []models.NASHotspotConfig
	if err := d.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, mapDBError(err)
	}

	for _, row := range rows {
		url := strings.TrimRight(defaultString(row.RadiusAPIURL, d.cfg.BaseURL), "/")
		key := defaultString(row.RadiusAPIKey, d.cfg.APIKey)
		d.addEndpoint(&out, seen, defaultString(row.ShortName, row.NASName), url, key)
	}

	if len(out) == 0 && d.cfg.BaseURL != "" && d.cfg.APIKey != "" {
		url := strings.TrimRight(d.cfg.BaseURL, "/")
		d.addEndpoint(&out, seen, "default", url, d.cfg.APIKey)
	}

	return out, nil
}

func (d *RadiusDirectory) addEndpoint(out *[]RadiusEndpoint, seen map[string]struct{}, name, url, key string) {
	url = strings.TrimRight(url, "/")
	if url == "" || key == "" {
		return
	}
	dedupeKey := url + "\x00" + key
	if _, ok := seen[dedupeKey]; ok {
		return
	}
	seen[dedupeKey] = struct{}{}
	*out = append(*out, RadiusEndpoint{
		Name:   defaultString(name, url),
		URL:    url,
		Client: radius.NewClientWith(url, key, d.cfg.Timeout),
	})
}
