package services

import (
	"context"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"gorm.io/gorm"
)

// RadiusServerService manages branch-local radius-api endpoints.
type RadiusServerService struct {
	db *gorm.DB
}

func NewRadiusServerService(db *gorm.DB) *RadiusServerService {
	return &RadiusServerService{db: db}
}

func (s *RadiusServerService) List(ctx context.Context) ([]dto.RadiusServerOutput, error) {
	var rows []models.RadiusServer
	if err := s.db.WithContext(ctx).Order("is_default DESC, name ASC").Find(&rows).Error; err != nil {
		return nil, mapDBError(err)
	}
	out := make([]dto.RadiusServerOutput, 0, len(rows))
	for _, row := range rows {
		out = append(out, radiusServerOutput(row))
	}
	return out, nil
}

func (s *RadiusServerService) Create(ctx context.Context, in dto.RadiusServerInput) (*dto.RadiusServerOutput, error) {
	row := radiusServerFromInput(in)
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if row.IsDefault {
			if err := tx.Model(&models.RadiusServer{}).Where("is_default = true").Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&row).Error
	}); err != nil {
		return nil, mapDBError(err)
	}
	out := radiusServerOutput(row)
	return &out, nil
}

func (s *RadiusServerService) Update(ctx context.Context, id uint, in dto.RadiusServerInput) (*dto.RadiusServerOutput, error) {
	row := radiusServerFromInput(in)
	row.ID = id
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if row.IsDefault {
			if err := tx.Model(&models.RadiusServer{}).Where("id <> ? AND is_default = true", id).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.RadiusServer{}).Where("id = ?", id).Updates(map[string]interface{}{
			"name":        row.Name,
			"api_url":     row.APIURL,
			"api_key":     row.APIKey,
			"radius_ip":   row.RadiusIP,
			"coa_port":    row.CoAPort,
			"description": row.Description,
			"is_default":  row.IsDefault,
		}).Error
	}); err != nil {
		return nil, mapDBError(err)
	}
	if err := s.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, mapDBError(err)
	}
	out := radiusServerOutput(row)
	return &out, nil
}

func (s *RadiusServerService) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Delete(&models.RadiusServer{}, id).Error; err != nil {
		return mapDBError(err)
	}
	return nil
}

func radiusServerFromInput(in dto.RadiusServerInput) models.RadiusServer {
	return models.RadiusServer{
		Name:        in.Name,
		APIURL:      strings.TrimRight(in.APIURL, "/"),
		APIKey:      in.APIKey,
		RadiusIP:    in.RadiusIP,
		CoAPort:     defaultString(in.CoAPort, "3799"),
		Description: in.Description,
		IsDefault:   in.IsDefault,
	}
}

func radiusServerOutput(row models.RadiusServer) dto.RadiusServerOutput {
	return dto.RadiusServerOutput{
		ID:          row.ID,
		Name:        row.Name,
		APIURL:      row.APIURL,
		APIKey:      row.APIKey,
		RadiusIP:    row.RadiusIP,
		CoAPort:     row.CoAPort,
		Description: row.Description,
		IsDefault:   row.IsDefault,
	}
}
