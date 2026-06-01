// Package config loads radius-api configuration from the environment.
package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config is the root configuration for the radius-api.
type Config struct {
	App    AppConfig
	DB     DBConfig
	Auth   AuthConfig
	CoA    CoAConfig
	Reload ReloadConfig
	CORS   CORSConfig
}

// AppConfig holds general settings.
type AppConfig struct {
	Name string
	Env  string
	Port string
}

// IsProduction reports whether the service runs in production.
func (a AppConfig) IsProduction() bool { return strings.EqualFold(a.Env, "production") }

// DBConfig holds MariaDB/MySQL connection settings (FreeRADIUS database).
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	Params   string
}

// DSN renders a go-sql-driver/mysql DSN (used by GORM and goose).
func (d DBConfig) DSN() string {
	params := d.Params
	if params == "" {
		params = "charset=utf8mb4&parseTime=True&loc=Local"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		d.User, d.Password, d.Host, d.Port, d.Name, params)
}

// AuthConfig holds the shared API key the billing backend must present.
type AuthConfig struct {
	APIKey string
}

// CoAConfig controls how disconnect (PoD) packets are sent to NAS devices.
type CoAConfig struct {
	Port    int           // NAS CoA listen port (Mikrotik default 3799)
	Timeout time.Duration // per-NAS exchange timeout
}

// ReloadConfig controls FreeRADIUS auto-reload after a NAS change. When a NAS is
// added/changed/removed, radius-api restarts the FreeRADIUS container (via the
// Docker socket) so the new SQL client list takes effect immediately. Empty
// Container disables the feature (local `go run`, or manual-reload setups).
type ReloadConfig struct {
	Container    string // FreeRADIUS container name to restart
	DockerSocket string // path to the Docker Engine API unix socket
}

// CORSConfig controls cross-origin access (handy for the radius-api Swagger UI).
type CORSConfig struct {
	AllowedOrigins []string
}

// Load reads configuration from the environment.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("config: no .env file found, relying on process environment")
	}

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "Radius API"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8081"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "radius"),
			Password: getEnv("DB_PASSWORD", "radius"),
			Name:     getEnv("DB_NAME", "radius"),
			Params:   getEnv("DB_PARAMS", "charset=utf8mb4&parseTime=True&loc=Local"),
		},
		Auth: AuthConfig{
			APIKey: getEnv("API_KEY", "radius-shared-secret"),
		},
		CoA: CoAConfig{
			Port:    getEnvInt("COA_PORT", 3799),
			Timeout: getEnvDuration("COA_TIMEOUT", 5*time.Second),
		},
		Reload: ReloadConfig{
			Container:    getEnv("FREERADIUS_CONTAINER", ""),
			DockerSocket: getEnv("DOCKER_SOCKET", "/var/run/docker.sock"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		},
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
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
			if t := strings.TrimSpace(p); t != "" {
				out = append(out, t)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return fallback
}
