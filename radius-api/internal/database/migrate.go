package database

import (
	"fmt"

	"github.com/dirhamt/billing-hotspot/radius-api/migrations"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// Migrate applies all pending goose migrations (FreeRADIUS schema).
func Migrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("database: sql.DB for migrate: %w", err)
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("mysql"); err != nil {
		return fmt.Errorf("database: set goose dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("database: goose up: %w", err)
	}
	return nil
}
