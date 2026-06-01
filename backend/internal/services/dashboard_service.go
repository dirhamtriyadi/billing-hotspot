package services

import (
	"context"
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"gorm.io/gorm"
)

// DashboardService computes the operator dashboard summary.
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService builds a DashboardService.
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// Stats aggregates package, voucher, order and revenue figures.
func (s *DashboardService) Stats(ctx context.Context) (*dto.DashboardStats, error) {
	db := s.db.WithContext(ctx)
	stats := &dto.DashboardStats{VouchersByStatus: map[string]int64{}}

	if err := db.Model(&models.Package{}).Count(&stats.TotalPackages).Error; err != nil {
		return nil, mapDBError(err)
	}
	if err := db.Model(&models.Package{}).Where("is_active = ?", true).Count(&stats.ActivePackages).Error; err != nil {
		return nil, mapDBError(err)
	}
	if err := db.Model(&models.Voucher{}).Count(&stats.TotalVouchers).Error; err != nil {
		return nil, mapDBError(err)
	}

	var statusRows []struct {
		Status string
		Count  int64
	}
	if err := db.Model(&models.Voucher{}).Select("status, count(*) as count").Group("status").Scan(&statusRows).Error; err != nil {
		return nil, mapDBError(err)
	}
	for _, r := range statusRows {
		stats.VouchersByStatus[r.Status] = r.Count
	}

	if err := db.Model(&models.Order{}).Count(&stats.TotalOrders).Error; err != nil {
		return nil, mapDBError(err)
	}
	if err := db.Model(&models.Order{}).Where("status = ?", models.OrderPaid).Count(&stats.PaidOrders).Error; err != nil {
		return nil, mapDBError(err)
	}
	if err := db.Model(&models.Order{}).Where("status = ?", models.OrderPending).Count(&stats.PendingOrders).Error; err != nil {
		return nil, mapDBError(err)
	}

	now := time.Now()
	startToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	stats.RevenueTotal = s.revenue(ctx, nil)
	stats.RevenueToday = s.revenue(ctx, &startToday)
	stats.RevenueMonth = s.revenue(ctx, &startMonth)

	if err := db.Preload("Package").Preload("Voucher").Order("id DESC").Limit(5).Find(&stats.RecentOrders).Error; err != nil {
		return nil, mapDBError(err)
	}

	return stats, nil
}

// revenue sums paid order amounts, optionally since a given time (by paid_at).
func (s *DashboardService) revenue(ctx context.Context, since *time.Time) int64 {
	q := s.db.WithContext(ctx).Model(&models.Order{}).Where("status = ?", models.OrderPaid)
	if since != nil {
		q = q.Where("paid_at >= ?", *since)
	}
	var result struct{ Total int64 }
	q.Select("COALESCE(SUM(amount), 0) as total").Scan(&result)
	return result.Total
}
