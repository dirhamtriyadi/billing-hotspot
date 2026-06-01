// Package services holds the business logic. Services depend on *gorm.DB and
// the external clients (radius, payment) and return apperror.AppError values so
// the HTTP layer can render them uniformly.
package services

import (
	"errors"

	"github.com/dirhamt/billing-hotspot/backend/internal/apperror"
	"gorm.io/gorm"
)

// mapDBError converts a GORM/driver error into an AppError. Relies on
// gorm.Config.TranslateError to surface ErrRecordNotFound / ErrDuplicatedKey.
func mapDBError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return apperror.NotFound("")
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return apperror.Conflict("A record with the same unique value already exists")
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return apperror.Unprocessable("Operation violates a data relationship constraint")
	default:
		return apperror.Internal("Database error").WithCause(err)
	}
}
