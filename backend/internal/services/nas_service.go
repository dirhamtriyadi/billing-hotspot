package services

import (
	"context"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/radius"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NasService manages NAS / RADIUS clients (the Mikrotik routers). The records
// live in the FreeRADIUS database owned by the radius-api microservice. Local
// hotspot deployment settings are stored in the billing database and merged into
// the admin-facing response.
type NasService struct {
	db     *gorm.DB
	app    config.AppConfig
	radius config.RadiusConfig
}

// NewNasService builds a NasService.
func NewNasService(db *gorm.DB, appCfg config.AppConfig, radiusCfg config.RadiusConfig) *NasService {
	return &NasService{db: db, app: appCfg, radius: radiusCfg}
}

// List returns every registered NAS.
func (s *NasService) List(ctx context.Context) ([]dto.NASOutput, error) {
	var configs []models.NASHotspotConfig
	err := s.db.WithContext(ctx).Order("id ASC").Find(&configs).Error
	if err != nil {
		return nil, mapDBError(err)
	}

	out := make([]dto.NASOutput, 0, len(configs))
	for _, cfg := range configs {
		out = append(out, outputFromConfig(cfg))
	}
	return out, nil
}

// Upsert registers or updates a NAS keyed by nasname.
func (s *NasService) Upsert(ctx context.Context, in dto.NASInput) (*dto.NASOutput, error) {
	cfg := s.configFromInput(in.NASName, in)
	var previous models.NASHotspotConfig
	hadPrevious := s.db.WithContext(ctx).Where("nasname = ?", in.NASName).First(&previous).Error == nil

	client := s.clientForConfig(cfg)
	if _, err := client.UpsertNAS(ctx, radius.NASInput{
		NASName:     in.NASName,
		ShortName:   in.ShortName,
		Secret:      in.Secret,
		Type:        in.Type,
		Description: in.Description,
		Ports:       in.Ports,
	}); err != nil {
		return nil, err
	}
	if hadPrevious && s.endpointChanged(previous, cfg) {
		s.deleteRemoteNASBestEffort(ctx, previous)
	}

	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "nasname"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"shortname",
			"type",
			"ports",
			"secret",
			"description",
			"radius_api_url",
			"radius_api_key",
			"radius_ip",
			"frontend_url",
			"backend_url",
			"frontend_host",
			"coa_port",
			"wan_interface",
			"hotspot_interface",
			"bridge_ports",
			"hotspot_network",
			"hotspot_gateway",
			"hotspot_pool_range",
			"hotspot_dns",
			"updated_at",
			"deleted_at",
		}),
	}).Create(&cfg).Error; err != nil {
		return nil, mapDBError(err)
	}
	if err := s.db.WithContext(ctx).Where("nasname = ?", in.NASName).First(&cfg).Error; err != nil {
		return nil, mapDBError(err)
	}

	out := outputFromConfig(cfg)
	return &out, nil
}

// Delete removes a NAS by id.
func (s *NasService) Delete(ctx context.Context, id uint) error {
	var cfg models.NASHotspotConfig
	if err := s.db.WithContext(ctx).First(&cfg, id).Error; err != nil {
		return mapDBError(err)
	}
	if err := s.deleteRemoteNAS(ctx, cfg); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Delete(&cfg).Error; err != nil {
		return mapDBError(err)
	}
	return nil
}

func (s *NasService) configFromInput(nasname string, in dto.NASInput) models.NASHotspotConfig {
	hotspot := in.HotspotConfig
	return models.NASHotspotConfig{
		NASName:          nasname,
		ShortName:        in.ShortName,
		Type:             defaultString(in.Type, "mikrotik"),
		Ports:            in.Ports,
		Secret:           in.Secret,
		Description:      in.Description,
		RadiusAPIURL:     strings.TrimRight(hotspot.RadiusAPIURL, "/"),
		RadiusAPIKey:     hotspot.RadiusAPIKey,
		RadiusIP:         hotspot.RadiusIP,
		FrontendURL:      strings.TrimRight(defaultString(hotspot.FrontendURL, s.app.FrontendURL), "/"),
		BackendURL:       strings.TrimRight(defaultString(hotspot.BackendURL, s.app.BaseURL), "/"),
		FrontendHost:     hotspot.FrontendHost,
		CoAPort:          defaultString(hotspot.CoAPort, "3799"),
		WANInterface:     defaultString(hotspot.WANInterface, "ether1"),
		HotspotInterface: defaultString(hotspot.HotspotInterface, "bridge-hotspot"),
		BridgePorts:      defaultString(hotspot.BridgePorts, "wlan1,wlan2"),
		HotspotNetwork:   defaultString(hotspot.HotspotNetwork, "10.5.50.0/24"),
		HotspotGateway:   defaultString(hotspot.HotspotGateway, "10.5.50.1"),
		HotspotPoolRange: defaultString(hotspot.HotspotPoolRange, "10.5.50.10-10.5.50.254"),
		HotspotDNS:       defaultString(hotspot.HotspotDNS, "8.8.8.8,1.1.1.1"),
	}
}

func outputFromConfig(cfg models.NASHotspotConfig) dto.NASOutput {
	return dto.NASOutput{
		ID:          cfg.ID,
		NASName:     cfg.NASName,
		ShortName:   cfg.ShortName,
		Type:        cfg.Type,
		Ports:       cfg.Ports,
		Secret:      cfg.Secret,
		Description: cfg.Description,
		HotspotConfig: dto.NASHotspotConfigOutput{
			RadiusAPIURL:     cfg.RadiusAPIURL,
			RadiusAPIKey:     cfg.RadiusAPIKey,
			RadiusIP:         cfg.RadiusIP,
			FrontendURL:      cfg.FrontendURL,
			BackendURL:       cfg.BackendURL,
			FrontendHost:     cfg.FrontendHost,
			CoAPort:          cfg.CoAPort,
			WANInterface:     cfg.WANInterface,
			HotspotInterface: cfg.HotspotInterface,
			BridgePorts:      cfg.BridgePorts,
			HotspotNetwork:   cfg.HotspotNetwork,
			HotspotGateway:   cfg.HotspotGateway,
			HotspotPoolRange: cfg.HotspotPoolRange,
			HotspotDNS:       cfg.HotspotDNS,
		},
	}
}

func (s *NasService) clientForConfig(cfg models.NASHotspotConfig) *radius.Client {
	url := defaultString(cfg.RadiusAPIURL, s.radius.BaseURL)
	key := defaultString(cfg.RadiusAPIKey, s.radius.APIKey)
	return radius.NewClientWith(strings.TrimRight(url, "/"), key, s.radius.Timeout)
}

func (s *NasService) deleteRemoteNAS(ctx context.Context, cfg models.NASHotspotConfig) error {
	client := s.clientForConfig(cfg)
	list, err := client.ListNAS(ctx)
	if err != nil {
		return err
	}
	for _, n := range list {
		if n.NASName == cfg.NASName {
			return client.DeleteNAS(ctx, n.ID)
		}
	}
	return nil
}

func (s *NasService) deleteRemoteNASBestEffort(ctx context.Context, cfg models.NASHotspotConfig) {
	_ = s.deleteRemoteNAS(ctx, cfg)
}

func (s *NasService) endpointChanged(a, b models.NASHotspotConfig) bool {
	aURL := strings.TrimRight(defaultString(a.RadiusAPIURL, s.radius.BaseURL), "/")
	bURL := strings.TrimRight(defaultString(b.RadiusAPIURL, s.radius.BaseURL), "/")
	aKey := defaultString(a.RadiusAPIKey, s.radius.APIKey)
	bKey := defaultString(b.RadiusAPIKey, s.radius.APIKey)
	return aURL != bURL || aKey != bKey
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
