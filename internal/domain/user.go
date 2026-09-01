package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Built-in role ids, seeded by migration 000006; ids because a name can be renamed.
var (
	RoleIDAdmin = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	RoleIDUser  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

// Status: account state set by an admin, not a measure of user activity.
type Status string

const (
	StatusEnabled  Status = "enabled"
	StatusDisabled Status = "disabled"
	StatusInvited  Status = "invited"
)

func (s Status) Valid() bool {
	switch s {
	case StatusEnabled, StatusDisabled, StatusInvited:
		return true
	default:
		return false
	}
}

// User: a pure entity, with no json/db tags.
type User struct {
	ID                uuid.UUID
	Email             string
	FullName          string
	PasswordHash      string
	RoleID            uuid.UUID
	RoleName          string // from the roles join, never written back
	Status            Status
	PasswordChangedAt time.Time
	// Attempts reset on a successful login; LockedUntil is nil when no lock is active.
	FailedLoginAttempts int
	LockedUntil         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// UserFilter for list + pagination.
type UserFilter struct {
	Search  string
	Limit   int
	Offset  int
	OrderBy string // ORDER BY clause, may only be filled via pagination.Sortable
}

// UserRepository: port to storage.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, filter UserFilter) ([]User, int, error)
	Update(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	IncrementFailedLogin(ctx context.Context, userID uuid.UUID) (int, error)
	LockUntil(ctx context.Context, userID uuid.UUID, until time.Time) error
	ClearFailedLogins(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// Input models owned by the service, separate from the HTTP DTOs.
type CreateUserInput struct {
	Email    string
	FullName string
	Password string
	RoleID   uuid.UUID
}

type UpdateUserInput struct {
	FullName *string
	RoleID   *uuid.UUID
}

// UserService: the business logic contract for users.
type UserService interface {
	Create(ctx context.Context, input CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, filter UserFilter) ([]User, int, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
