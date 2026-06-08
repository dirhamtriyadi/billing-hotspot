package services

import (
	"context"
	"strings"
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/radius"
	"gorm.io/gorm"
)

const defaultRadiusAPITimeout = 10 * time.Second

// RadiusEndpoint is one branch-local radius-api target.
type RadiusEndpoint struct {
	Name   string
	URL    string
	Client *radius.Client
}

// RadiusDirectory resolves all radius-api endpoints configured in Radius Server
// master data and NAS records.
type RadiusDirectory struct {
	db *gorm.DB
}

func NewRadiusDirectory(db *gorm.DB) *RadiusDirectory {
	return &RadiusDirectory{db: db}
}

// Endpoints returns unique radius-api targets from managed Radius Server rows
// and per-NAS manual overrides.
func (d *RadiusDirectory) Endpoints(ctx context.Context) ([]RadiusEndpoint, error) {
	seen := map[string]struct{}{}
	out := make([]RadiusEndpoint, 0)

	var servers []models.RadiusServer
	if err := d.db.WithContext(ctx).Find(&servers).Error; err != nil {
		return nil, mapDBError(err)
	}
	for _, srv := range servers {
		d.addEndpoint(&out, seen, srv.Name, srv.APIURL, srv.APIKey, srv.Timeout)
	}

	var rows []models.NASHotspotConfig
	if err := d.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, mapDBError(err)
	}

	for _, row := range rows {
		d.addEndpoint(&out, seen, defaultString(row.ShortName, row.NASName), row.RadiusAPIURL, row.RadiusAPIKey, "")
	}

	return out, nil
}

func (d *RadiusDirectory) addEndpoint(out *[]RadiusEndpoint, seen map[string]struct{}, name, url, key, timeoutRaw string) {
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
		Client: radius.NewClientWith(url, key, parseRadiusTimeout(timeoutRaw)),
	})
}

func parseRadiusTimeout(raw string) time.Duration {
	timeout, err := time.ParseDuration(defaultString(raw, "10s"))
	if err != nil || timeout <= 0 {
		return defaultRadiusAPITimeout
	}
	return timeout
}
