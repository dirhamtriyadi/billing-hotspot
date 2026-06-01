package dto

// ReportQuery selects the inclusive date window for a revenue report. Dates are
// "YYYY-MM-DD"; when omitted the service defaults to the trailing 30 days.
type ReportQuery struct {
	Start string `form:"start" json:"start"`
	End   string `form:"end" json:"end"`
}

// ReportSummary holds the headline figures for the selected window.
type ReportSummary struct {
	Start          string `json:"start"`
	End            string `json:"end"`
	RevenueTotal   int64  `json:"revenue_total"`
	PaidOrders     int64  `json:"paid_orders"`
	TotalOrders    int64  `json:"total_orders"`
	VouchersIssued int64  `json:"vouchers_issued"`
	AvgOrderValue  int64  `json:"avg_order_value"`
}

// RevenuePoint is one day on the revenue trend line. The series is gap-filled so
// days without sales appear as zero.
type RevenuePoint struct {
	Date    string `json:"date"` // YYYY-MM-DD
	Revenue int64  `json:"revenue"`
	Orders  int64  `json:"orders"`
}

// RevenueByMethod breaks revenue down by payment method.
type RevenueByMethod struct {
	Method  string `json:"method"`
	Revenue int64  `json:"revenue"`
	Orders  int64  `json:"orders"`
}

// RevenueByPackage breaks revenue down by package.
type RevenueByPackage struct {
	PackageID   uint   `json:"package_id"`
	PackageName string `json:"package_name"`
	Revenue     int64  `json:"revenue"`
	Orders      int64  `json:"orders"`
}

// Report is the full revenue report payload.
type Report struct {
	Summary   ReportSummary      `json:"summary"`
	Series    []RevenuePoint     `json:"series"`
	ByMethod  []RevenueByMethod  `json:"by_method"`
	ByPackage []RevenueByPackage `json:"by_package"`
}
