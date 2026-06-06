package services

import (
	"context"
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"gorm.io/gorm"
)

const reportDateLayout = "2006-01-02"

// ReportService computes revenue analytics over a date window from paid orders.
type ReportService struct {
	db *gorm.DB
}

// NewReportService builds a ReportService.
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

// resolveRange parses the query into an inclusive [start, end] day window,
// defaulting to the trailing 30 days. The returned `until` is exclusive
// (end + 1 day) for use in half-open SQL range filters.
func resolveRange(q dto.ReportQuery) (start, end, until time.Time) {
	loc := time.Now().Location()
	today := time.Now().In(loc)
	end = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, loc)
	if t, err := time.ParseInLocation(reportDateLayout, q.End, loc); err == nil {
		end = t
	}
	start = end.AddDate(0, 0, -29)
	if t, err := time.ParseInLocation(reportDateLayout, q.Start, loc); err == nil {
		start = t
	}
	if start.After(end) {
		start, end = end, start
	}
	until = end.AddDate(0, 0, 1)
	return start, end, until
}

// Generate produces the full revenue report for the selected window.
func (s *ReportService) Generate(ctx context.Context, q dto.ReportQuery) (*dto.Report, error) {
	db := s.db.WithContext(ctx)
	start, end, until := resolveRange(q)

	report := &dto.Report{
		Summary: dto.ReportSummary{
			Start: start.Format(reportDateLayout),
			End:   end.Format(reportDateLayout),
		},
		Series:    []dto.RevenuePoint{},
		ByMethod:  []dto.RevenueByMethod{},
		ByPackage: []dto.RevenueByPackage{},
	}

	// Headline figures — revenue is based on paid orders by paid_at.
	var paid struct {
		Revenue int64
		Count   int64
	}
	if err := db.Model(&models.Order{}).
		Where("status = ? AND paid_at >= ? AND paid_at < ?", models.OrderPaid, start, until).
		Select("COALESCE(SUM(amount),0) as revenue, COUNT(*) as count").
		Scan(&paid).Error; err != nil {
		return nil, mapDBError(err)
	}
	report.Summary.RevenueTotal = paid.Revenue
	report.Summary.PaidOrders = paid.Count
	if paid.Count > 0 {
		report.Summary.AvgOrderValue = paid.Revenue / paid.Count
	}

	if err := db.Model(&models.Order{}).
		Where("created_at >= ? AND created_at < ?", start, until).
		Count(&report.Summary.TotalOrders).Error; err != nil {
		return nil, mapDBError(err)
	}
	if err := db.Model(&models.Voucher{}).
		Where("created_at >= ? AND created_at < ?", start, until).
		Count(&report.Summary.VouchersIssued).Error; err != nil {
		return nil, mapDBError(err)
	}

	// Daily trend, gap-filled so every day in the window is present.
	var dayRows []struct {
		Day     string
		Revenue int64
		Orders  int64
	}
	if err := db.Model(&models.Order{}).
		Where("status = ? AND paid_at >= ? AND paid_at < ?", models.OrderPaid, start, until).
		Select("to_char(paid_at, 'YYYY-MM-DD') as day, COALESCE(SUM(amount),0) as revenue, COUNT(*) as orders").
		Group("day").Scan(&dayRows).Error; err != nil {
		return nil, mapDBError(err)
	}
	byDay := make(map[string]struct {
		Revenue int64
		Orders  int64
	}, len(dayRows))
	for _, r := range dayRows {
		byDay[r.Day] = struct {
			Revenue int64
			Orders  int64
		}{r.Revenue, r.Orders}
	}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format(reportDateLayout)
		v := byDay[key]
		report.Series = append(report.Series, dto.RevenuePoint{
			Date:    key,
			Revenue: v.Revenue,
			Orders:  v.Orders,
		})
	}

	// Breakdown by payment method.
	if err := db.Model(&models.Order{}).
		Where("status = ? AND paid_at >= ? AND paid_at < ?", models.OrderPaid, start, until).
		Select("payment_method as method, COALESCE(SUM(amount),0) as revenue, COUNT(*) as orders").
		Group("payment_method").Order("revenue DESC").
		Scan(&report.ByMethod).Error; err != nil {
		return nil, mapDBError(err)
	}

	// Breakdown by package.
	if err := db.Model(&models.Order{}).
		Joins("JOIN packages ON packages.id = orders.package_id").
		Where("orders.status = ? AND orders.paid_at >= ? AND orders.paid_at < ?", models.OrderPaid, start, until).
		Select("packages.id as package_id, packages.name as package_name, COALESCE(SUM(orders.amount),0) as revenue, COUNT(*) as orders").
		Group("packages.id, packages.name").Order("revenue DESC").
		Scan(&report.ByPackage).Error; err != nil {
		return nil, mapDBError(err)
	}

	return report, nil
}

// ExportOrders returns the orders created within the window (all statuses), with
// their package preloaded, for CSV export.
func (s *ReportService) ExportOrders(ctx context.Context, q dto.ReportQuery) ([]models.Order, time.Time, time.Time, error) {
	start, end, until := resolveRange(q)
	var orders []models.Order
	if err := s.db.WithContext(ctx).Preload("Package").
		Where("created_at >= ? AND created_at < ?", start, until).
		Order("id DESC").Find(&orders).Error; err != nil {
		return nil, start, end, mapDBError(err)
	}
	return orders, start, end, nil
}
