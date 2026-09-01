package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// defaultUserOrder: fallback when the filter arrives without OrderBy.
// Columns are table-qualified: every query below joins roles, which has
// created_at and id of its own.
const defaultUserOrder = "users.created_at DESC, users.id ASC"

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository returns an interface, not a concrete struct.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// withRole: users plus the name of the role they point at. The join stays here
// rather than in roleRepository because ORDER BY roles.name and LIMIT have to
// run in the same statement; resolving names in Go would break sorted paging.
// LEFT JOIN so a soft-deleted role hides the name, never the user itself.
func (r *userRepository) withRole(ctx context.Context) *gorm.DB {
	return dbFrom(ctx, r.db).
		Model(&userModel{}).
		Select("users.*, roles.name AS role_name").
		Joins("LEFT JOIN roles ON roles.id = users.role_id")
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	model := userFromDomain(user)
	if err := dbFrom(ctx, r.db).Create(&model).Error; err != nil {
		return wrapErr("create user", err)
	}

	// Columns with database defaults come back via RETURNING.
	user.Status = domain.Status(model.Status)
	user.PasswordChangedAt = model.PasswordChangedAt
	user.CreatedAt = model.CreatedAt
	user.UpdatedAt = model.UpdatedAt

	// RETURNING cannot reach the joined role name, so the row is read back once.
	var created userModel
	if err := r.withRole(ctx).Where("users.id = ?", user.ID).Take(&created).Error; err == nil {
		user.RoleName = created.RoleName
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var model userModel
	if err := r.withRole(ctx).Where("users.id = ?", id).Take(&model).Error; err != nil {
		return nil, wrapErr("get user", err)
	}
	return model.toDomain(), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model userModel
	err := r.withRole(ctx).Where("users.email = ?", strings.ToLower(email)).Take(&model).Error
	if err != nil {
		return nil, wrapErr("get user", err)
	}
	return model.toDomain(), nil
}

func (r *userRepository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	query := dbFrom(ctx, r.db).Model(&userModel{})

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(users.email) LIKE ? OR LOWER(users.full_name) LIKE ?", keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, wrapErr("count users", err)
	}
	if total == 0 {
		return []domain.User{}, 0, nil
	}

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultUserOrder
	}

	var models []userModel
	err := query.
		Select("users.*, roles.name AS role_name").
		Joins("LEFT JOIN roles ON roles.id = users.role_id").
		Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).Find(&models).Error
	if err != nil {
		return nil, 0, wrapErr("list users", err)
	}
	return usersToDomain(models), int(total), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	result := dbFrom(ctx, r.db).
		Model(&userModel{ID: user.ID}).
		Updates(map[string]any{
			"full_name": user.FullName,
			"role_id":   user.RoleID,
		})
	if result.Error != nil {
		return wrapErr("update user", result.Error)
	}
	if result.RowsAffected == 0 {
		return wrapErr("update user", gorm.ErrRecordNotFound)
	}

	// Re-read what the database owns: updated_at, plus the joined role name.
	var refreshed userModel
	if err := r.withRole(ctx).Where("users.id = ?", user.ID).Take(&refreshed).Error; err == nil {
		user.UpdatedAt = refreshed.UpdatedAt
		user.RoleName = refreshed.RoleName
	}
	return nil
}

// Delete: soft delete via the deleted_at column.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := dbFrom(ctx, r.db).Delete(&userModel{}, "id = ?", id)
	if result.Error != nil {
		return wrapErr("delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return wrapErr("delete user", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := dbFrom(ctx, r.db).Model(&userModel{}).
		Where("email = ?", strings.ToLower(email)).
		Limit(1).Count(&count).Error
	if err != nil {
		return false, wrapErr("check email exists", err)
	}
	return count > 0, nil
}

// UpdatePassword is kept separate from Update so the user PATCH endpoint can
// never reach these columns, whatever the client sends.
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	result := dbFrom(ctx, r.db).
		Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_hash":       passwordHash,
			"password_changed_at": time.Now(),
		})
	if result.Error != nil {
		return wrapErr("update password", result.Error)
	}
	if result.RowsAffected == 0 {
		return wrapErr("update password", gorm.ErrRecordNotFound)
	}
	return nil
}

// IncrementFailedLogin counts one rejected sign-in and returns the new total.
// One statement on purpose: read-then-write would let parallel attempts share
// a count and slip past the limit.
func (r *userRepository) IncrementFailedLogin(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `
		UPDATE users SET failed_login_attempts = failed_login_attempts + 1, updated_at = now()
		WHERE id = ? AND deleted_at IS NULL
		RETURNING failed_login_attempts`

	var row struct{ FailedLoginAttempts int }
	if err := dbFrom(ctx, r.db).Raw(q, userID).Scan(&row).Error; err != nil {
		return 0, wrapErr("increment failed login", err)
	}
	return row.FailedLoginAttempts, nil
}

// LockUntil closes the account to sign-ins until the given time.
func (r *userRepository) LockUntil(ctx context.Context, userID uuid.UUID, until time.Time) error {
	err := dbFrom(ctx, r.db).
		Model(&userModel{}).
		Where("id = ?", userID).
		Update("locked_until", until).Error
	if err != nil {
		return wrapErr("lock user", err)
	}
	return nil
}

// ClearFailedLogins wipes the counter after a successful login.
func (r *userRepository) ClearFailedLogins(ctx context.Context, userID uuid.UUID) error {
	err := dbFrom(ctx, r.db).
		Model(&userModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{"failed_login_attempts": 0, "locked_until": nil}).Error
	if err != nil {
		return wrapErr("clear failed logins", err)
	}
	return nil
}
