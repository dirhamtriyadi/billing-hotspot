// Package database manages the MariaDB/MySQL connection (FreeRADIUS database)
// and applies goose migrations.
package database

import (
	"fmt"
	"time"

	"github.com/dirhamt/billing-hotspot/radius-api/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a pooled GORM connection to MariaDB/MySQL.
func Connect(cfg config.DBConfig, isProduction bool) (*gorm.DB, error) {
	logLevel := logger.Info
	if isProduction {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
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
