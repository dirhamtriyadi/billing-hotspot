package payment

import (
	"strconv"

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
)

// OverlayConfig returns a copy of base with any non-empty values from the
// key/value settings store applied on top. This lets operators manage gateway
// credentials from the admin UI while environment variables remain the
// fallback/default. Empty stored values never clobber the environment.
func OverlayConfig(base config.PaymentConfig, s map[string]string) config.PaymentConfig {
	out := base

	if v := s["payment_default_provider"]; v != "" {
		out.DefaultProvider = v
	}

	if v := s["midtrans_server_key"]; v != "" {
		out.Midtrans.ServerKey = v
	}
	if v := s["midtrans_client_key"]; v != "" {
		out.Midtrans.ClientKey = v
	}
	if v, ok := s["midtrans_is_production"]; ok && v != "" {
		out.Midtrans.IsProduction = parseBool(v)
	}

	if v := s["xendit_secret_key"]; v != "" {
		out.Xendit.SecretKey = v
	}
	if v := s["xendit_callback_token"]; v != "" {
		out.Xendit.CallbackToken = v
	}

	if v := s["tripay_api_key"]; v != "" {
		out.Tripay.APIKey = v
	}
	if v := s["tripay_private_key"]; v != "" {
		out.Tripay.PrivateKey = v
	}
	if v := s["tripay_merchant_code"]; v != "" {
		out.Tripay.MerchantCode = v
	}
	if v, ok := s["tripay_is_production"]; ok && v != "" {
		out.Tripay.IsProduction = parseBool(v)
	}

	return out
}

func parseBool(v string) bool {
	b, _ := strconv.ParseBool(v)
	return b
}
