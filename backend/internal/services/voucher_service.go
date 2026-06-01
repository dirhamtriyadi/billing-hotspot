package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/radius"
	"github.com/dirhamt/billing-hotspot/backend/internal/util"
	"gorm.io/gorm"
)

// VoucherService manages voucher lifecycle and FreeRADIUS provisioning.
type VoucherService struct {
	db     *gorm.DB
	radius *radius.Client
}

// NewVoucherService builds a VoucherService.
func NewVoucherService(db *gorm.DB, r *radius.Client) *VoucherService {
	return &VoucherService{db: db, radius: r}
}

// GenerateBatch creates a named batch of vouchers for a package and provisions
// them into FreeRADIUS in a single bulk call.
func (s *VoucherService) GenerateBatch(ctx context.Context, req dto.CreateBatchRequest, createdBy uint) (*models.VoucherBatch, error) {
	var p models.Package
	if err := s.db.WithContext(ctx).First(&p, req.PackageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Package not found")
		}
		return nil, mapDBError(err)
	}

	codeLen := req.CodeLength
	if codeLen == 0 {
		codeLen = 8
	}
	name := req.Name
	if name == "" {
		name = "Batch " + time.Now().Format("2006-01-02 15:04")
	}

	// Pre-generate codes, deduped within the batch and against existing rows.
	codes, err := s.generateUniqueCodes(ctx, req.Prefix, codeLen, req.Quantity)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := computeExpiry(p, now)

	batch := models.VoucherBatch{
		Name:       name,
		PackageID:  p.ID,
		Prefix:     req.Prefix,
		Quantity:   req.Quantity,
		CodeLength: codeLen,
		CreatedBy:  createdBy,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}
		vouchers := make([]models.Voucher, 0, len(codes))
		for _, code := range codes {
			vouchers = append(vouchers, models.Voucher{
				Code:      code,
				PackageID: p.ID,
				BatchID:   &batch.ID,
				Status:    models.VoucherUnused,
				Profile:   p.Profile,
				Price:     p.Price,
				ExpiresAt: expiresAt,
			})
		}
		return tx.CreateInBatches(&vouchers, 200).Error
	})
	if err != nil {
		return nil, mapDBError(err)
	}

	// Best-effort bulk provisioning. On failure the vouchers persist as "unused"
	// and can be re-synced individually; we surface the problem in the log.
	s.provisionBatch(ctx, &batch, p)

	if err := s.db.WithContext(ctx).Preload("Package").Preload("Vouchers").First(&batch, batch.ID).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &batch, nil
}

// List returns a filtered, paginated voucher list.
func (s *VoucherService) List(ctx context.Context, q dto.VoucherListQuery) ([]models.Voucher, int64, error) {
	q.Normalize()
	tx := s.db.WithContext(ctx).Model(&models.Voucher{})
	if q.Search != "" {
		tx = tx.Where("code ILIKE ?", "%"+q.Search+"%")
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.PackageID != 0 {
		tx = tx.Where("package_id = ?", q.PackageID)
	}
	if q.BatchID != 0 {
		tx = tx.Where("batch_id = ?", q.BatchID)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, mapDBError(err)
	}

	var vouchers []models.Voucher
	if err := tx.Preload("Package").Order("id DESC").
		Offset(q.Offset()).Limit(q.PerPage).Find(&vouchers).Error; err != nil {
		return nil, 0, mapDBError(err)
	}
	return vouchers, total, nil
}

// Get fetches a voucher by id.
func (s *VoucherService) Get(ctx context.Context, id uint) (*models.Voucher, error) {
	var v models.Voucher
	if err := s.db.WithContext(ctx).Preload("Package").First(&v, id).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &v, nil
}

// UpdateStatus enables (re-provisions) or disables (revokes) a voucher.
func (s *VoucherService) UpdateStatus(ctx context.Context, id uint, status string) (*models.Voucher, error) {
	v, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	switch status {
	case models.VoucherActive:
		if err := s.provision(ctx, v, &v.Package); err != nil {
			return nil, err
		}
	case models.VoucherDisabled:
		if err := s.revoke(ctx, v); err != nil {
			return nil, err
		}
	default:
		return nil, apperror.BadRequest("Unsupported voucher status: " + status)
	}
	return v, nil
}

// Delete revokes a voucher in FreeRADIUS (best effort) and soft-deletes the row.
func (s *VoucherService) Delete(ctx context.Context, id uint) error {
	v, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if v.SyncedToRadius {
		_ = s.radius.DisconnectUser(ctx, v.Code)
		if err := s.radius.DeleteUser(ctx, v.Code); err != nil {
			slog.Warn("failed to delete radius user during voucher delete",
				slog.String("code", v.Code), slog.Any("error", err))
		}
	}
	if err := s.db.WithContext(ctx).Delete(&models.Voucher{}, id).Error; err != nil {
		return mapDBError(err)
	}
	return nil
}

// IssueForOrder mints a single voucher for a paid order and provisions it. A
// radius provisioning failure is logged (the payment already succeeded) but the
// voucher row is still returned and linked to the order.
func (s *VoucherService) IssueForOrder(ctx context.Context, order *models.Order) (*models.Voucher, error) {
	var p models.Package
	if err := s.db.WithContext(ctx).First(&p, order.PackageID).Error; err != nil {
		return nil, mapDBError(err)
	}

	codes, err := s.generateUniqueCodes(ctx, "", 8, 1)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	v := models.Voucher{
		Code:      codes[0],
		PackageID: p.ID,
		OrderID:   &order.ID,
		Status:    models.VoucherUnused,
		Profile:   p.Profile,
		Price:     p.Price,
		ExpiresAt: computeExpiry(p, now),
	}
	if err := s.db.WithContext(ctx).Create(&v).Error; err != nil {
		return nil, mapDBError(err)
	}

	if err := s.provision(ctx, &v, &p); err != nil {
		slog.Error("voucher provisioning failed after payment — needs manual re-sync",
			slog.String("code", v.Code),
			slog.String("order", order.OrderNumber),
			slog.Any("error", err),
		)
	}
	return &v, nil
}

// provision upserts the package profile and creates the radius user, then marks
// the voucher active + synced.
func (s *VoucherService) provision(ctx context.Context, v *models.Voucher, p *models.Package) error {
	if err := s.radius.UpsertProfile(ctx, buildProfile(*p)); err != nil {
		return err
	}
	if err := s.radius.CreateUser(ctx, radius.User{
		Username:  v.Code,
		Password:  v.Code,
		Profile:   p.Profile,
		ExpiresAt: v.ExpiresAt,
	}); err != nil {
		return err
	}
	v.SyncedToRadius = true
	v.Status = models.VoucherActive
	return s.db.WithContext(ctx).Model(v).
		Updates(map[string]interface{}{"synced_to_radius": true, "status": models.VoucherActive}).Error
}

// revoke deletes the radius user and marks the voucher disabled.
func (s *VoucherService) revoke(ctx context.Context, v *models.Voucher) error {
	_ = s.radius.DisconnectUser(ctx, v.Code)
	if err := s.radius.DeleteUser(ctx, v.Code); err != nil {
		return err
	}
	v.SyncedToRadius = false
	v.Status = models.VoucherDisabled
	return s.db.WithContext(ctx).Model(v).
		Updates(map[string]interface{}{"synced_to_radius": false, "status": models.VoucherDisabled}).Error
}

// provisionBatch bulk-provisions a freshly created batch.
func (s *VoucherService) provisionBatch(ctx context.Context, batch *models.VoucherBatch, p models.Package) {
	var vouchers []models.Voucher
	if err := s.db.WithContext(ctx).Where("batch_id = ?", batch.ID).Find(&vouchers).Error; err != nil {
		slog.Warn("failed to load batch vouchers for provisioning", slog.Any("error", err))
		return
	}

	if err := s.radius.UpsertProfile(ctx, buildProfile(p)); err != nil {
		slog.Warn("batch profile sync failed", slog.String("profile", p.Profile), slog.Any("error", err))
		return
	}

	users := make([]radius.User, 0, len(vouchers))
	ids := make([]uint, 0, len(vouchers))
	for _, v := range vouchers {
		users = append(users, radius.User{Username: v.Code, Password: v.Code, Profile: p.Profile, ExpiresAt: v.ExpiresAt})
		ids = append(ids, v.ID)
	}

	if err := s.radius.CreateUsers(ctx, users); err != nil {
		slog.Error("batch radius provisioning failed — vouchers need re-sync",
			slog.Uint64("batch_id", uint64(batch.ID)), slog.Any("error", err))
		return
	}

	if err := s.db.WithContext(ctx).Model(&models.Voucher{}).Where("id IN ?", ids).
		Updates(map[string]interface{}{"synced_to_radius": true, "status": models.VoucherActive}).Error; err != nil {
		slog.Warn("failed to mark batch vouchers active", slog.Any("error", err))
	}
}

// generateUniqueCodes returns n codes guaranteed unique within the batch and
// against existing (including soft-deleted) voucher codes.
func (s *VoucherService) generateUniqueCodes(ctx context.Context, prefix string, length, n int) ([]string, error) {
	seen := make(map[string]struct{}, n)
	out := make([]string, 0, n)
	attempts := 0
	maxAttempts := n*5 + 50

	for len(out) < n {
		attempts++
		if attempts > maxAttempts {
			return nil, apperror.Internal("Failed to generate unique voucher codes; try a longer code length")
		}
		code, err := util.GenerateCode(prefix, length)
		if err != nil {
			return nil, apperror.Internal("Failed to generate voucher code").WithCause(err)
		}
		if _, dup := seen[code]; dup {
			continue
		}
		var count int64
		if err := s.db.WithContext(ctx).Unscoped().Model(&models.Voucher{}).Where("code = ?", code).Count(&count).Error; err != nil {
			return nil, mapDBError(err)
		}
		if count > 0 {
			continue
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}
	return out, nil
}
