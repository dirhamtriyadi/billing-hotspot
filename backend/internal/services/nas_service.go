package services

import (
	"context"

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
	radius *radius.Client
	db     *gorm.DB
}

// NewNasService builds a NasService.
func NewNasService(client *radius.Client, db *gorm.DB) *NasService {
	return &NasService{radius: client, db: db}
}

// List returns every registered NAS.
func (s *NasService) List(ctx context.Context) ([]dto.NASOutput, error) {
	list, err := s.radius.ListNAS(ctx)
	if err != nil {
		return nil, err
	}

	configs, err := s.configsByNASName(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]dto.NASOutput, 0, len(list))
	for _, n := range list {
		out = append(out, mergeNAS(n, configs[n.NASName]))
	}
	return out, nil
}

// Upsert registers or updates a NAS keyed by nasname.
func (s *NasService) Upsert(ctx context.Context, in dto.NASInput) (*dto.NASOutput, error) {
	nas, err := s.radius.UpsertNAS(ctx, radius.NASInput{
		NASName:     in.NASName,
		ShortName:   in.ShortName,
		Secret:      in.Secret,
		Type:        in.Type,
		Description: in.Description,
		Ports:       in.Ports,
	})
	if err != nil {
		return nil, err
	}

	cfg := configFromInput(in.NASName, in.HotspotConfig)
	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "nasname"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"radius_ip",
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

	out := mergeNAS(*nas, &cfg)
	return &out, nil
}

// Delete removes a NAS by id.
func (s *NasService) Delete(ctx context.Context, id uint) error {
	list, err := s.radius.ListNAS(ctx)
	if err != nil {
		return err
	}
	nasname := ""
	for _, n := range list {
		if n.ID == id {
			nasname = n.NASName
			break
		}
	}

	if err := s.radius.DeleteNAS(ctx, id); err != nil {
		return err
	}
	if nasname == "" {
		return nil
	}
	if err := s.db.WithContext(ctx).
		Where("nasname = ?", nasname).
		Delete(&models.NASHotspotConfig{}).Error; err != nil {
		return mapDBError(err)
	}
	return nil
}

func (s *NasService) configsByNASName(ctx context.Context) (map[string]*models.NASHotspotConfig, error) {
	var rows []models.NASHotspotConfig
	if err := s.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, mapDBError(err)
	}
	out := make(map[string]*models.NASHotspotConfig, len(rows))
	for i := range rows {
		out[rows[i].NASName] = &rows[i]
	}
	return out, nil
}

func configFromInput(nasname string, in dto.NASHotspotConfigInput) models.NASHotspotConfig {
	return models.NASHotspotConfig{
		NASName:          nasname,
		RadiusIP:         in.RadiusIP,
		FrontendHost:     in.FrontendHost,
		CoAPort:          defaultString(in.CoAPort, "3799"),
		WANInterface:     defaultString(in.WANInterface, "ether1"),
		HotspotInterface: defaultString(in.HotspotInterface, "bridge-hotspot"),
		BridgePorts:      defaultString(in.BridgePorts, "wlan1,wlan2"),
		HotspotNetwork:   defaultString(in.HotspotNetwork, "10.5.50.0/24"),
		HotspotGateway:   defaultString(in.HotspotGateway, "10.5.50.1"),
		HotspotPoolRange: defaultString(in.HotspotPoolRange, "10.5.50.10-10.5.50.254"),
		HotspotDNS:       defaultString(in.HotspotDNS, "8.8.8.8,1.1.1.1"),
	}
}

func mergeNAS(n radius.NAS, cfg *models.NASHotspotConfig) dto.NASOutput {
	out := dto.NASOutput{
		ID:          n.ID,
		NASName:     n.NASName,
		ShortName:   n.ShortName,
		Type:        n.Type,
		Ports:       n.Ports,
		Secret:      n.Secret,
		Server:      n.Server,
		Community:   n.Community,
		Description: n.Description,
		HotspotConfig: dto.NASHotspotConfigOutput{
			CoAPort:          "3799",
			WANInterface:     "ether1",
			HotspotInterface: "bridge-hotspot",
			BridgePorts:      "wlan1,wlan2",
			HotspotNetwork:   "10.5.50.0/24",
			HotspotGateway:   "10.5.50.1",
			HotspotPoolRange: "10.5.50.10-10.5.50.254",
			HotspotDNS:       "8.8.8.8,1.1.1.1",
		},
	}
	if cfg == nil {
		return out
	}
	out.HotspotConfig = dto.NASHotspotConfigOutput{
		RadiusIP:         cfg.RadiusIP,
		FrontendHost:     cfg.FrontendHost,
		CoAPort:          cfg.CoAPort,
		WANInterface:     cfg.WANInterface,
		HotspotInterface: cfg.HotspotInterface,
		BridgePorts:      cfg.BridgePorts,
		HotspotNetwork:   cfg.HotspotNetwork,
		HotspotGateway:   cfg.HotspotGateway,
		HotspotPoolRange: cfg.HotspotPoolRange,
		HotspotDNS:       cfg.HotspotDNS,
	}
	return out
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
