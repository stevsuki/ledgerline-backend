package postgres

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// wrapErr translates GORM errors into domain errors.
func wrapErr(op string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("%s: %w", op, domain.ErrNotFound)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return fmt.Errorf("%s: %w", op, domain.ErrConflict)
	// A bad role_id reaches the database as a foreign key violation, not a 500.
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return fmt.Errorf("%s: %w", op, domain.ErrInvalidInput)
	default:
		return fmt.Errorf("%s: %w", op, err)
	}
}
