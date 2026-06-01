package services

import (
	"context"
	"strconv"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/dirhamt/billing-hotspot/backend/internal/dto"
	"github.com/dirhamt/billing-hotspot/backend/internal/payment"
)

// GatewayService manages payment-gateway credentials from the admin UI. Values
// are persisted in the key/value settings store and overlaid on top of the
// environment-provided defaults; on every change the live payment Registry is
// hot-reloaded so the new credentials take effect immediately.
type GatewayService struct {
	settings *SettingService
	base     config.PaymentConfig
	registry *payment.Registry
}

// NewGatewayService builds a GatewayService.
func NewGatewayService(settings *SettingService, base config.PaymentConfig, registry *payment.Registry) *GatewayService {
	return &GatewayService{settings: settings, base: base, registry: registry}
}

// Reload rebuilds the payment registry from the stored settings overlaid on the
// environment defaults. Called once at startup and after every update.
func (s *GatewayService) Reload(ctx context.Context) error {
	all, err := s.settings.GetAll(ctx)
	if err != nil {
		return err
	}
	s.registry.Reload(payment.OverlayConfig(s.base, all))
	return nil
}

// Get returns the masked gateway configuration view.
func (s *GatewayService) Get(ctx context.Context) (*dto.GatewaySettings, error) {
	all, err := s.settings.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return s.view(all), nil
}

// Update applies the provided changes, persists them, hot-reloads the registry,
// and returns the refreshed masked view.
func (s *GatewayService) Update(ctx context.Context, in dto.GatewayUpdate) (*dto.GatewaySettings, error) {
	all, err := s.settings.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	changes := map[string]string{}
	set := func(key, val string) { changes[key] = val }
	// setSecret keeps the existing value when the incoming one is blank so the
	// masked preview shown in the UI can be submitted back harmlessly.
	setSecret := func(key string, val *string) {
		if val != nil && strings.TrimSpace(*val) != "" {
			changes[key] = *val
		}
	}

	if in.DefaultProvider != nil {
		set("payment_default_provider", *in.DefaultProvider)
	}
	if in.EnableCash != nil {
		set("enable_cash", strconv.FormatBool(*in.EnableCash))
	}

	// Track which providers are offered to customers (enabled_providers CSV).
	enabled := parseProviders(all["enabled_providers"])
	if in.Midtrans != nil {
		setSecret("midtrans_server_key", in.Midtrans.ServerKey)
		if in.Midtrans.ClientKey != nil {
			set("midtrans_client_key", *in.Midtrans.ClientKey)
		}
		if in.Midtrans.Production != nil {
			set("midtrans_is_production", strconv.FormatBool(*in.Midtrans.Production))
		}
		if in.Midtrans.Enabled != nil {
			enabled["midtrans"] = *in.Midtrans.Enabled
		}
	}
	if in.Xendit != nil {
		setSecret("xendit_secret_key", in.Xendit.SecretKey)
		setSecret("xendit_callback_token", in.Xendit.CallbackToken)
		if in.Xendit.Enabled != nil {
			enabled["xendit"] = *in.Xendit.Enabled
		}
	}
	if in.Tripay != nil {
		setSecret("tripay_api_key", in.Tripay.APIKey)
		setSecret("tripay_private_key", in.Tripay.PrivateKey)
		if in.Tripay.MerchantCode != nil {
			set("tripay_merchant_code", *in.Tripay.MerchantCode)
		}
		if in.Tripay.Production != nil {
			set("tripay_is_production", strconv.FormatBool(*in.Tripay.Production))
		}
		if in.Tripay.Enabled != nil {
			enabled["tripay"] = *in.Tripay.Enabled
		}
	}
	set("enabled_providers", joinProviders(enabled))

	if _, err := s.settings.Update(ctx, changes); err != nil {
		return nil, err
	}
	if err := s.Reload(ctx); err != nil {
		return nil, err
	}
	return s.Get(ctx)
}

// view builds the masked DTO from the effective (overlaid) configuration.
func (s *GatewayService) view(all map[string]string) *dto.GatewaySettings {
	cfg := payment.OverlayConfig(s.base, all)
	enabled := parseProviders(all["enabled_providers"])

	midtransConfigured := cfg.Midtrans.ServerKey != ""
	xenditConfigured := cfg.Xendit.SecretKey != ""
	tripayConfigured := cfg.Tripay.APIKey != "" && cfg.Tripay.PrivateKey != "" && cfg.Tripay.MerchantCode != ""

	return &dto.GatewaySettings{
		DefaultProvider: cfg.DefaultProvider,
		EnableCash:      strings.EqualFold(all["enable_cash"], "true"),
		Midtrans: dto.GatewayMidtrans{
			Enabled:    enabled["midtrans"],
			Configured: midtransConfigured,
			Production: cfg.Midtrans.IsProduction,
			ServerKey:  mask(cfg.Midtrans.ServerKey),
			ClientKey:  cfg.Midtrans.ClientKey,
		},
		Xendit: dto.GatewayXendit{
			Enabled:       enabled["xendit"],
			Configured:    xenditConfigured,
			SecretKey:     mask(cfg.Xendit.SecretKey),
			CallbackToken: mask(cfg.Xendit.CallbackToken),
		},
		Tripay: dto.GatewayTripay{
			Enabled:      enabled["tripay"],
			Configured:   tripayConfigured,
			Production:   cfg.Tripay.IsProduction,
			APIKey:       mask(cfg.Tripay.APIKey),
			PrivateKey:   mask(cfg.Tripay.PrivateKey),
			MerchantCode: cfg.Tripay.MerchantCode,
		},
	}
}

// mask returns a short, non-reversible preview of a secret (its last 4 chars).
func mask(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return "••••"
	}
	return "••••" + s[len(s)-4:]
}

func parseProviders(csv string) map[string]bool {
	out := map[string]bool{}
	for _, p := range strings.Split(csv, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out[p] = true
		}
	}
	return out
}

func joinProviders(m map[string]bool) string {
	order := []string{"midtrans", "xendit", "tripay"}
	parts := make([]string, 0, len(order))
	for _, p := range order {
		if m[p] {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, ",")
}
