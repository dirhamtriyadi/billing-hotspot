package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/util"
	"gorm.io/gorm"
)

// PackageService manages sellable packages and keeps their FreeRADIUS profiles
// in sync via the radius-api.
type PackageService struct {
	db     *gorm.DB
	radius *RadiusDirectory
}

// NewPackageService builds a PackageService.
func NewPackageService(db *gorm.DB, r *RadiusDirectory) *PackageService {
	return &PackageService{db: db, radius: r}
}

// List returns a paginated, optionally searched list of packages (admin view,
// includes inactive packages).
func (s *PackageService) List(ctx context.Context, q dto.PageQuery) ([]models.Package, int64, error) {
	q.Normalize()
	tx := s.db.WithContext(ctx).Model(&models.Package{})
	if q.Search != "" {
		tx = tx.Where("name ILIKE ?", "%"+q.Search+"%")
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, mapDBError(err)
	}

	var packages []models.Package
	if err := tx.Order("sort_order ASC, id ASC").
		Offset(q.Offset()).Limit(q.PerPage).Find(&packages).Error; err != nil {
		return nil, 0, mapDBError(err)
	}
	return packages, total, nil
}

// ListPublic returns active packages ordered for the landing page.
func (s *PackageService) ListPublic(ctx context.Context) ([]models.Package, error) {
	var packages []models.Package
	if err := s.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, price ASC").
		Find(&packages).Error; err != nil {
		return nil, mapDBError(err)
	}
	return packages, nil
}

// Get fetches a package by id.
func (s *PackageService) Get(ctx context.Context, id uint) (*models.Package, error) {
	var p models.Package
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &p, nil
}

// GetBySlug fetches an active package by slug (public).
func (s *PackageService) GetBySlug(ctx context.Context, slug string) (*models.Package, error) {
	var p models.Package
	if err := s.db.WithContext(ctx).Where("slug = ? AND is_active = ?", slug, true).First(&p).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &p, nil
}

// Create persists a new package and best-effort syncs its radius profile.
func (s *PackageService) Create(ctx context.Context, req dto.PackageRequest) (*models.Package, error) {
	slug, err := s.uniqueSlug(ctx, util.Slugify(req.Name))
	if err != nil {
		return nil, err
	}

	p := models.Package{
		Name:               req.Name,
		Slug:               slug,
		Description:        req.Description,
		Price:              req.Price,
		Profile:            slug,
		RateDownKbps:       req.RateDownKbps,
		RateUpKbps:         req.RateUpKbps,
		BurstEnabled:       req.BurstEnabled,
		ValidityValue:      req.ValidityValue,
		ValidityUnit:       req.ValidityUnit,
		SessionTimeoutSecs: req.SessionTimeoutSecs,
		DataQuotaMB:        req.DataQuotaMB,
		SimultaneousUse:    req.SimultaneousUse,
		Highlight:          req.Highlight,
		BadgeText:          req.BadgeText,
		Color:              defaultStr(req.Color, "#2563eb"),
		Icon:               defaultStr(req.Icon, "wifi"),
		SortOrder:          req.SortOrder,
		IsActive:           req.IsActive == nil || *req.IsActive,
	}

	if err := s.db.WithContext(ctx).Create(&p).Error; err != nil {
		return nil, mapDBError(err)
	}

	s.syncProfile(ctx, p)
	return &p, nil
}

// Update replaces the editable fields of a package. Slug/profile are immutable
// so existing vouchers keep referencing the same radius profile.
func (s *PackageService) Update(ctx context.Context, id uint, req dto.PackageRequest) (*models.Package, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	p.Name = req.Name
	p.Description = req.Description
	p.Price = req.Price
	p.RateDownKbps = req.RateDownKbps
	p.RateUpKbps = req.RateUpKbps
	p.BurstEnabled = req.BurstEnabled
	p.ValidityValue = req.ValidityValue
	p.ValidityUnit = req.ValidityUnit
	p.SessionTimeoutSecs = req.SessionTimeoutSecs
	p.DataQuotaMB = req.DataQuotaMB
	p.SimultaneousUse = req.SimultaneousUse
	p.Highlight = req.Highlight
	p.BadgeText = req.BadgeText
	p.Color = defaultStr(req.Color, "#2563eb")
	p.Icon = defaultStr(req.Icon, "wifi")
	p.SortOrder = req.SortOrder
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}

	if err := s.db.WithContext(ctx).Save(p).Error; err != nil {
		return nil, mapDBError(err)
	}

	s.syncProfile(ctx, *p)
	return p, nil
}

// Delete soft-deletes a package. It refuses deletion when vouchers reference it
// to preserve sales history (admins should deactivate instead).
func (s *PackageService) Delete(ctx context.Context, id uint) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Voucher{}).Where("package_id = ?", id).Count(&count).Error; err != nil {
		return mapDBError(err)
	}
	if count > 0 {
		return apperror.Conflict("This package has vouchers and cannot be deleted; deactivate it instead")
	}
	if err := s.db.WithContext(ctx).Delete(&models.Package{}, id).Error; err != nil {
		return mapDBError(err)
	}
	return nil
}

// uniqueSlug ensures slug uniqueness by appending a numeric suffix on collision.
func (s *PackageService) uniqueSlug(ctx context.Context, base string) (string, error) {
	if base == "" {
		base = "paket"
	}
	slug := base
	for i := 2; ; i++ {
		var count int64
		if err := s.db.WithContext(ctx).Unscoped().Model(&models.Package{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
			return "", mapDBError(err)
		}
		if count == 0 {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// syncProfile pushes the package's attributes to the radius-api. Failures are
// logged but not fatal: provisioning a voucher re-upserts the profile anyway.
func (s *PackageService) syncProfile(ctx context.Context, p models.Package) {
	endpoints, err := s.radius.Endpoints(ctx)
	if err != nil {
		slog.Warn("failed to sync radius profile",
			slog.String("profile", p.Profile),
			slog.Any("error", err),
		)
		return
	}
	for _, endpoint := range endpoints {
		if err := endpoint.Client.UpsertProfile(ctx, buildProfile(p)); err != nil {
			slog.Warn("failed to sync radius profile",
				slog.String("profile", p.Profile),
				slog.String("radius_api", endpoint.URL),
				slog.Any("error", err),
			)
		}
	}
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
