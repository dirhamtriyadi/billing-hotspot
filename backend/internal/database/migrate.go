package database

import (
	"fmt"

	"github.com/dirhamt/billing-hotspot/backend/migrations"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// Migrate applies all pending goose migrations embedded in the binary.
func Migrate(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("database: sql.DB for migrate: %w", err)
	}

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("database: set goose dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		return fmt.Errorf("database: goose up: %w", err)
	}
	return nil
}

// MigrateDown rolls back a single migration. Intended for development use.
func MigrateDown(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Down(sqlDB, ".")
}
