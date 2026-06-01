package services

import (
	"context"

	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SettingService manages the editable key/value business settings.
type SettingService struct {
	db *gorm.DB
}

// NewSettingService builds a SettingService.
func NewSettingService(db *gorm.DB) *SettingService {
	return &SettingService{db: db}
}

// GetAll returns every setting as a flat map.
func (s *SettingService) GetAll(ctx context.Context) (map[string]string, error) {
	var rows []models.Setting
	if err := s.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, mapDBError(err)
	}
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		out[r.Key] = r.Value
	}
	return out, nil
}

// Update upserts the provided settings and returns the full set.
func (s *SettingService) Update(ctx context.Context, values map[string]string) (map[string]string, error) {
	for k, v := range values {
		if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&models.Setting{Key: k, Value: v}).Error; err != nil {
			return nil, mapDBError(err)
		}
	}
	return s.GetAll(ctx)
}
