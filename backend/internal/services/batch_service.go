package services

import (
	"context"

	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"gorm.io/gorm"
)

// ListBatches returns a paginated list of voucher batches.
func (s *VoucherService) ListBatches(ctx context.Context, q dto.PageQuery) ([]models.VoucherBatch, int64, error) {
	q.Normalize()
	tx := s.db.WithContext(ctx).Model(&models.VoucherBatch{})
	if q.Search != "" {
		tx = tx.Where("name ILIKE ?", "%"+q.Search+"%")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, mapDBError(err)
	}

	var batches []models.VoucherBatch
	if err := tx.Preload("Package").Order("id DESC").
		Offset(q.Offset()).Limit(q.PerPage).Find(&batches).Error; err != nil {
		return nil, 0, mapDBError(err)
	}
	return batches, total, nil
}

// GetBatch returns a batch with its package and vouchers loaded.
func (s *VoucherService) GetBatch(ctx context.Context, id uint) (*models.VoucherBatch, error) {
	var batch models.VoucherBatch
	if err := s.db.WithContext(ctx).Preload("Package").Preload("Vouchers").First(&batch, id).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &batch, nil
}

// DeleteBatch revokes every voucher in the batch from FreeRADIUS (best effort)
// and soft-deletes the batch and its vouchers.
func (s *VoucherService) DeleteBatch(ctx context.Context, id uint) error {
	var vouchers []models.Voucher
	if err := s.db.WithContext(ctx).Where("batch_id = ?", id).Find(&vouchers).Error; err != nil {
		return mapDBError(err)
	}
	for _, v := range vouchers {
		if v.SyncedToRadius {
			s.revokeBestEffort(ctx, v.Code, "batch delete")
		}
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("batch_id = ?", id).Delete(&models.Voucher{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.VoucherBatch{}, id).Error
	})
}
