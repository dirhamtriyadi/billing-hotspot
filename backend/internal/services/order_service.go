package services

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/models"
	"github.com/dirhamt/billing-hotspot/backend/internal/payment"
	"github.com/dirhamt/billing-hotspot/backend/internal/util"
	"gorm.io/gorm"
)

// OrderService orchestrates self-service checkout, payment gateway charges,
// webhook reconciliation, and voucher issuance on payment success.
type OrderService struct {
	db       *gorm.DB
	payments *payment.Registry
	vouchers *VoucherService
	app      config.AppConfig
}

// NewOrderService builds an OrderService.
func NewOrderService(db *gorm.DB, payments *payment.Registry, vouchers *VoucherService, app config.AppConfig) *OrderService {
	return &OrderService{db: db, payments: payments, vouchers: vouchers, app: app}
}

// Checkout creates an order. For "cash" it stays pending until an operator
// confirms payment; for a gateway it charges immediately and returns the
// checkout URL/token.
func (s *OrderService) Checkout(ctx context.Context, req dto.CheckoutRequest) (*models.Order, error) {
	var p models.Package
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", req.PackageID, true).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("Package is not available")
		}
		return nil, mapDBError(err)
	}

	order := models.Order{
		OrderNumber:   util.OrderNumber(),
		PackageID:     p.ID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		CustomerEmail: req.CustomerEmail,
		Amount:        p.Price,
		PaymentMethod: req.PaymentMethod,
		Status:        models.OrderPending,
	}
	if err := s.db.WithContext(ctx).Create(&order).Error; err != nil {
		return nil, mapDBError(err)
	}

	if req.PaymentMethod == models.MethodCash {
		// Awaiting operator confirmation; no gateway interaction.
		return s.loadOrder(ctx, order.ID)
	}

	gw, err := s.payments.Get(req.PaymentMethod)
	if err != nil {
		s.db.WithContext(ctx).Model(&order).Update("status", models.OrderFailed)
		return nil, err
	}

	res, err := gw.Charge(ctx, payment.ChargeRequest{
		OrderNumber:        order.OrderNumber,
		Amount:             order.Amount,
		Description:        p.Name,
		CustomerName:       req.CustomerName,
		CustomerEmail:      req.CustomerEmail,
		CustomerPhone:      req.CustomerPhone,
		Channel:            req.Channel,
		SuccessRedirectURL: s.app.FrontendURL + "/payment/" + order.OrderNumber,
		FailureRedirectURL: s.app.FrontendURL + "/payment/" + order.OrderNumber,
		CallbackURL:        s.app.BaseURL + "/api/v1/webhooks/" + req.PaymentMethod,
	})
	if err != nil {
		s.db.WithContext(ctx).Model(&order).Update("status", models.OrderFailed)
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&order).Updates(map[string]interface{}{
		"reference":     res.Reference,
		"payment_url":   res.PaymentURL,
		"payment_token": res.PaymentToken,
		"qr_string":     res.QRString,
		"raw_response":  res.Raw,
	}).Error; err != nil {
		return nil, mapDBError(err)
	}

	return s.loadOrder(ctx, order.ID)
}

// List returns a filtered, paginated order list (admin).
func (s *OrderService) List(ctx context.Context, q dto.OrderListQuery) ([]models.Order, int64, error) {
	q.Normalize()
	tx := s.db.WithContext(ctx).Model(&models.Order{})
	if q.Search != "" {
		tx = tx.Where("order_number ILIKE ? OR customer_name ILIKE ? OR customer_phone ILIKE ?",
			"%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%")
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	if q.PaymentMethod != "" {
		tx = tx.Where("payment_method = ?", q.PaymentMethod)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, mapDBError(err)
	}

	var orders []models.Order
	if err := tx.Preload("Package").Preload("Voucher").Order("id DESC").
		Offset(q.Offset()).Limit(q.PerPage).Find(&orders).Error; err != nil {
		return nil, 0, mapDBError(err)
	}
	return orders, total, nil
}

// Get fetches an order by id (admin).
func (s *OrderService) Get(ctx context.Context, id uint) (*models.Order, error) {
	return s.loadOrder(ctx, id)
}

// GetByNumber fetches an order by its public order number (status page).
func (s *OrderService) GetByNumber(ctx context.Context, number string) (*models.Order, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Package").Preload("Voucher").
		Where("order_number = ?", number).First(&order).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &order, nil
}

// ConfirmCash marks a pending cash order as paid and issues its voucher.
func (s *OrderService) ConfirmCash(ctx context.Context, id uint) (*models.Order, error) {
	order, err := s.loadOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.PaymentMethod != models.MethodCash {
		return nil, apperror.BadRequest("Only cash orders can be confirmed manually")
	}
	if order.Status == models.OrderPaid {
		return order, nil // idempotent
	}
	if order.Status != models.OrderPending {
		return nil, apperror.BadRequest("Order is not awaiting payment")
	}
	if err := s.markPaid(ctx, order); err != nil {
		return nil, err
	}
	return s.loadOrder(ctx, id)
}

// MarkPaidManual force-settles a gateway (non-cash) order that the webhook
// never reconciled — e.g. the customer paid but the gateway callback failed or
// the payment page was closed. The operator confirms payment out-of-band (checks
// the gateway dashboard / mutation) and the system then issues the voucher just
// like a normal settlement. Idempotent if already paid; cash orders should use
// ConfirmCash instead.
func (s *OrderService) MarkPaidManual(ctx context.Context, id uint) (*models.Order, error) {
	order, err := s.loadOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if order.PaymentMethod == models.MethodCash {
		return nil, apperror.BadRequest("Use cash confirmation for cash orders")
	}
	if order.Status == models.OrderPaid {
		return order, nil // idempotent
	}
	// Allow settling orders that are stuck pending OR ended in failed/expired
	// (gateway error) — the operator has verified the money actually arrived.
	switch order.Status {
	case models.OrderPending, models.OrderFailed, models.OrderExpired:
		if err := s.markPaid(ctx, order); err != nil {
			return nil, err
		}
	default:
		return nil, apperror.BadRequest("Order cannot be settled from its current status")
	}
	return s.loadOrder(ctx, id)
}

// HandleWebhook verifies and applies a gateway notification. It is idempotent
// and records every callback in payment_logs.
func (s *OrderService) HandleWebhook(ctx context.Context, provider string, headers http.Header, body []byte) error {
	gw, err := s.payments.Get(provider)
	if err != nil {
		return err
	}
	res, err := gw.HandleWebhook(headers, body)
	if err != nil {
		return err
	}

	logEntry := models.PaymentLog{
		Provider:  provider,
		Event:     res.Event,
		Reference: res.Reference,
		Status:    res.Status,
		Valid:     res.Valid,
		Payload:   res.Raw,
	}

	var order models.Order
	orderErr := s.db.WithContext(ctx).Where("order_number = ?", res.OrderNumber).First(&order).Error
	if orderErr == nil {
		logEntry.OrderID = &order.ID
	}
	if err := s.db.WithContext(ctx).Create(&logEntry).Error; err != nil {
		slog.Warn("failed to persist payment log", slog.Any("error", err))
	}

	if !res.Valid {
		return apperror.Unauthorized("Invalid webhook signature")
	}
	if orderErr != nil {
		if errors.Is(orderErr, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Order not found for this notification")
		}
		return mapDBError(orderErr)
	}

	if order.Status == models.OrderPaid {
		return nil // already settled — idempotent no-op
	}

	switch res.Status {
	case payment.StatusPaid:
		return s.markPaid(ctx, &order)
	case payment.StatusExpired:
		return s.db.WithContext(ctx).Model(&order).Update("status", models.OrderExpired).Error
	case payment.StatusFailed:
		return s.db.WithContext(ctx).Model(&order).Update("status", models.OrderFailed).Error
	default:
		return nil // still pending
	}
}

// markPaid issues the voucher and transitions the order to paid.
func (s *OrderService) markPaid(ctx context.Context, order *models.Order) error {
	voucher, err := s.vouchers.IssueForOrder(ctx, order)
	if err != nil {
		return err
	}
	now := time.Now()
	return s.db.WithContext(ctx).Model(order).Updates(map[string]interface{}{
		"status":     models.OrderPaid,
		"paid_at":    now,
		"voucher_id": voucher.ID,
	}).Error
}

func (s *OrderService) loadOrder(ctx context.Context, id uint) (*models.Order, error) {
	var order models.Order
	if err := s.db.WithContext(ctx).Preload("Package").Preload("Voucher").First(&order, id).Error; err != nil {
		return nil, mapDBError(err)
	}
	return &order, nil
}
