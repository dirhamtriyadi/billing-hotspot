// Package database manages the PostgreSQL connection (via GORM) and applies
// goose migrations embedded into the binary.
package database

import (
	"fmt"
	"time"

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a pooled GORM connection to PostgreSQL.
func Connect(cfg config.DBConfig, isProduction bool) (*gorm.DB, error) {
	logLevel := logger.Info
	if isProduction {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
		TranslateError:         true, // normalise driver errors to gorm.Err* sentinels
	})
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("database: underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return db, nil
}
