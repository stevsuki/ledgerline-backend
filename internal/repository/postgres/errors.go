// Package postgres: GORM repositories; the GORM-tagged models live in the model subpackage.
package postgres

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// resourceErrors: what one repository reports for the three failures the database can raise
// on its own. Templates only; wrap copies them before attaching a cause.
type resourceErrors struct {
	notFound *domain.Error
	conflict *domain.Error
	invalid  *domain.Error // a foreign key pointing at nothing
}

// wrap translates a GORM error into this resource's domain error; op stays in the log only.
func (r resourceErrors) wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	cause := fmt.Errorf("%s: %w", op, err)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return pick(r.notFound, domain.NotFound(domain.CodeNotFound, "resource not found")).WithCause(cause)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return pick(r.conflict, domain.Conflict(domain.CodeConflict, "resource already exists")).WithCause(cause)
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return pick(r.invalid, domain.InvalidInput(domain.CodeInvalidInput, "a referenced record does not exist")).WithCause(cause)
	default:
		return cause
	}
}

func pick(e, fallback *domain.Error) *domain.Error {
	if e == nil {
		return fallback
	}
	return e
}

var (
	userErrors = resourceErrors{
		notFound: domain.NotFound(domain.CodeUserNotFound, "user not found"),
		conflict: domain.Conflict(domain.CodeUserEmailTaken, "email is already registered"),
		invalid:  domain.InvalidInput(domain.CodeUserInvalidRole, "role_id does not refer to an existing role").WithField("role_id"),
	}

	categoryErrors = resourceErrors{
		notFound: domain.NotFound(domain.CodeCategoryNotFound, "category not found"),
		conflict: domain.Conflict(domain.CodeCategoryNameTaken, "a category with that name already exists").WithField("name"),
	}

	roleErrors = resourceErrors{
		notFound: domain.NotFound(domain.CodeRoleNotFound, "role not found"),
		conflict: domain.Conflict(domain.CodeRoleNameTaken, "a role with that name already exists").WithField("name"),
		invalid:  domain.InvalidInput(domain.CodeRoleInvalidMenu, "one of the menu ids does not exist").WithField("permissions"),
	}

	walletErrors = resourceErrors{
		notFound: domain.NotFound(domain.CodeWalletNotFound, "wallet not found"),
		conflict: domain.Conflict(domain.CodeWalletNameTaken, "a wallet with that name already exists").WithField("name"),
	}

	menuErrors = resourceErrors{
		notFound: domain.NotFound(domain.CodeMenuNotFound, "menu not found"),
	}

	auditLogErrors = resourceErrors{
		notFound: domain.NotFound(domain.CodeAuditLogNotFound, "audit log not found"),
	}

	// Never addressed directly by a client; the service decides what a missing token means.
	passwordResetTokenErrors = resourceErrors{
		notFound: domain.NotFound(domain.CodeNotFound, "password reset token not found"),
	}
)
