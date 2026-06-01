package dto

import "github.com/dirhamt/billing-hotspot/backend/internal/models"

// DashboardStats is the operator dashboard summary.
type DashboardStats struct {
	TotalPackages    int64            `json:"total_packages"`
	ActivePackages   int64            `json:"active_packages"`
	TotalVouchers    int64            `json:"total_vouchers"`
	VouchersByStatus map[string]int64 `json:"vouchers_by_status"`
	TotalOrders      int64            `json:"total_orders"`
	PaidOrders       int64            `json:"paid_orders"`
	PendingOrders    int64            `json:"pending_orders"`
	RevenueTotal     int64            `json:"revenue_total"`
	RevenueToday     int64            `json:"revenue_today"`
	RevenueMonth     int64            `json:"revenue_month"`
	RecentOrders     []models.Order   `json:"recent_orders"`
}
