package handlers

import (
	"encoding/csv"
	"strconv"

	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/response"
	"github.com/dirhamt/billing-hotspot/backend/internal/services"
	"github.com/gin-gonic/gin"
)

// ReportHandler exposes revenue reporting and CSV export.
type ReportHandler struct {
	svc *services.ReportService
}

// NewReportHandler builds a ReportHandler.
func NewReportHandler(svc *services.ReportService) *ReportHandler {
	return &ReportHandler{svc: svc}
}

// Revenue godoc
// @Summary  Revenue report for a date window
// @Tags     Reports
// @Produce  json
// @Security BearerAuth
// @Param    start query string false "Start date (YYYY-MM-DD)"
// @Param    end   query string false "End date (YYYY-MM-DD)"
// @Success  200 {object} response.Envelope{data=dto.Report}
// @Router   /reports/revenue [get]
func (h *ReportHandler) Revenue(c *gin.Context) {
	var q dto.ReportQuery
	if !bindQuery(c, &q) {
		return
	}
	report, err := h.svc.Generate(c.Request.Context(), q)
	if err != nil {
		fail(c, err)
		return
	}
	response.OK(c, "OK", report)
}

// Export godoc
// @Summary  Export orders in a date window as CSV
// @Tags     Reports
// @Produce  text/csv
// @Security BearerAuth
// @Param    start query string false "Start date (YYYY-MM-DD)"
// @Param    end   query string false "End date (YYYY-MM-DD)"
// @Success  200 {string} string "CSV file"
// @Router   /reports/export [get]
func (h *ReportHandler) Export(c *gin.Context) {
	var q dto.ReportQuery
	if !bindQuery(c, &q) {
		return
	}
	orders, start, end, err := h.svc.ExportOrders(c.Request.Context(), q)
	if err != nil {
		fail(c, err)
		return
	}

	filename := "laporan-" + start.Format("20060102") + "-" + end.Format("20060102") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{
		"No. Pesanan", "Tanggal Dibuat", "Tanggal Bayar", "Pelanggan", "No. HP",
		"Paket", "Jumlah", "Metode", "Status",
	})
	for _, o := range orders {
		pkgName := ""
		if o.Package.ID != 0 {
			pkgName = o.Package.Name
		}
		paidAt := ""
		if o.PaidAt != nil {
			paidAt = o.PaidAt.Format("2006-01-02 15:04")
		}
		_ = w.Write([]string{
			o.OrderNumber,
			o.CreatedAt.Format("2006-01-02 15:04"),
			paidAt,
			o.CustomerName,
			o.CustomerPhone,
			pkgName,
			strconv.FormatInt(o.Amount, 10),
			o.PaymentMethod,
			o.Status,
		})
	}
	w.Flush()
}
