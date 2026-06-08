// Package config loads and exposes application configuration sourced from
// environment variables (optionally seeded from a .env file).
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config is the root configuration aggregate for the billing backend.
type Config struct {
	App     AppConfig
	DB      DBConfig
	JWT     JWTConfig
	Payment PaymentConfig
	CORS    CORSConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name        string
	Env         string // development | production
	Port        string
	BaseURL     string // public URL of THIS backend (used to build payment callback URLs)
	FrontendURL string // public URL of the SPA (used for redirect after payment)
}

// IsProduction reports whether the app runs in a production environment.
func (a AppConfig) IsProduction() bool { return strings.EqualFold(a.Env, "production") }

// DBConfig holds PostgreSQL connection settings for the billing database.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

// DSN renders a GORM/pgx compatible connection string.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.TimeZone,
	)
}

// JWTConfig holds admin authentication token settings.
type JWTConfig struct {
	Secret    string
	Issuer    string
	AccessTTL time.Duration
}

// PaymentConfig aggregates every supported payment provider's credentials.
type PaymentConfig struct {
	DefaultProvider string // xendit | midtrans | tripay
	Xendit          XenditConfig
	Midtrans        MidtransConfig
	Tripay          TripayConfig
}

// XenditConfig holds Xendit credentials. CallbackToken verifies invoice webhooks.
type XenditConfig struct {
	SecretKey     string
	CallbackToken string
}

// MidtransConfig holds Midtrans Snap credentials.
type MidtransConfig struct {
	ServerKey    string
	ClientKey    string
	IsProduction bool
}

// TripayConfig holds Tripay credentials. PrivateKey signs/validates callbacks.
type TripayConfig struct {
	APIKey       string
	PrivateKey   string
	MerchantCode string
	IsProduction bool
}

// CORSConfig controls cross-origin access for the SPA.
type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads configuration from the environment. It best-effort loads a .env
// file first (ignored if absent) so local development needs no exported vars.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("config: no .env file found, relying on process environment")
	}

	cfg := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "Billing Hotspot"),
			Env:         getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			BaseURL:     getEnv("APP_BASE_URL", "http://localhost:8080"),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "billing"),
			Password: getEnv("DB_PASSWORD", "billing"),
			Name:     getEnv("DB_NAME", "billing_hotspot"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "change-me-in-production-please-32chars"),
			Issuer:    getEnv("JWT_ISSUER", "billing-hotspot"),
			AccessTTL: getEnvDuration("JWT_ACCESS_TTL", 24*time.Hour),
		},
		Payment: PaymentConfig{
			DefaultProvider: getEnv("PAYMENT_DEFAULT_PROVIDER", "midtrans"),
			Xendit: XenditConfig{
				SecretKey:     getEnv("XENDIT_SECRET_KEY", ""),
				CallbackToken: getEnv("XENDIT_CALLBACK_TOKEN", ""),
			},
			Midtrans: MidtransConfig{
				ServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
				ClientKey:    getEnv("MIDTRANS_CLIENT_KEY", ""),
				IsProduction: getEnvBool("MIDTRANS_IS_PRODUCTION", false),
			},
			Tripay: TripayConfig{
				APIKey:       getEnv("TRIPAY_API_KEY", ""),
				PrivateKey:   getEnv("TRIPAY_PRIVATE_KEY", ""),
				MerchantCode: getEnv("TRIPAY_MERCHANT_CODE", ""),
				IsProduction: getEnvBool("TRIPAY_IS_PRODUCTION", false),
			},
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				out = append(out, trimmed)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return fallback
}

var _ = getEnvInt // reserved for future numeric settings
