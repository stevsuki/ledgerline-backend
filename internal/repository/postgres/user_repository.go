package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

// defaultUserOrder: fallback without OrderBy, qualified because every query joins roles.
const defaultUserOrder = "users.created_at DESC, users.id ASC"

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository returns an interface, not a concrete struct.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// withRole: users plus their role name, joined here so ORDER BY roles.name can page.
func (r *userRepository) withRole(ctx context.Context) *gorm.DB {
	return dbFrom(ctx, r.db).
		Model(&model.UserModel{}).
		Select("users.*, roles.name AS role_name").
		Joins("LEFT JOIN roles ON roles.id = users.role_id")
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	// Nil on self-registration: nobody was signed in to be the author.
	actor := domain.ActorFrom(ctx)
	user.CreatedBy, user.UpdatedBy = actor, actor

	row := model.UserFromDomain(user)
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return userErrors.wrap("create user", err)
	}

	// Columns with database defaults come back via RETURNING.
	user.Status = domain.Status(row.Status)
	user.PasswordChangedAt = row.PasswordChangedAt
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt

	// RETURNING cannot reach the joined role name, so the row is read back once.
	var created model.UserModel
	if err := r.withRole(ctx).Where("users.id = ?", user.ID).Take(&created).Error; err == nil {
		user.RoleName = created.RoleName
	}
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var row model.UserModel
	if err := r.withRole(ctx).Where("users.id = ?", id).Take(&row).Error; err != nil {
		return nil, userErrors.wrap("get user", err)
	}
	return row.ToDomain(), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var row model.UserModel
	err := r.withRole(ctx).Where("users.email = ?", strings.ToLower(email)).Take(&row).Error
	if err != nil {
		return nil, userErrors.wrap("get user", err)
	}
	return row.ToDomain(), nil
}

func (r *userRepository) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, int, error) {
	query := dbFrom(ctx, r.db).Model(&model.UserModel{})

	if filter.Search != "" {
		keyword := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(users.email) LIKE ? OR LOWER(users.full_name) LIKE ?", keyword, keyword)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, userErrors.wrap("count users", err)
	}
	if total == 0 {
		return []domain.User{}, 0, nil
	}

	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = defaultUserOrder
	}

	var rows []model.UserModel
	err := query.
		Select("users.*, roles.name AS role_name").
		Joins("LEFT JOIN roles ON roles.id = users.role_id").
		Order(orderBy).Limit(filter.Limit).Offset(filter.Offset).Find(&rows).Error
	if err != nil {
		return nil, 0, userErrors.wrap("list users", err)
	}
	return model.UsersToDomain(rows), int(total), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedBy = domain.ActorFrom(ctx)

	result := dbFrom(ctx, r.db).
		Model(&model.UserModel{ID: user.ID}).
		Updates(map[string]any{
			"full_name":  user.FullName,
			"role_id":    user.RoleID,
			"updated_by": user.UpdatedBy,
		})
	if result.Error != nil {
		return userErrors.wrap("update user", result.Error)
	}
	if result.RowsAffected == 0 {
		return userErrors.wrap("update user", gorm.ErrRecordNotFound)
	}

	// Re-read what the database owns: updated_at, plus the joined role name.
	var refreshed model.UserModel
	if err := r.withRole(ctx).Where("users.id = ?", user.ID).Take(&refreshed).Error; err == nil {
		user.UpdatedAt = refreshed.UpdatedAt
		user.RoleName = refreshed.RoleName
	}
	return nil
}

// Delete: soft delete via the deleted_at column, stamped with who did it.
// One statement rather than Delete plus an update, so the two can never diverge;
// UpdateColumns keeps updated_at pointing at the last real edit. GORM still adds
// "deleted_at IS NULL", so deleting twice reports not found as before.
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := dbFrom(ctx, r.db).
		Model(&model.UserModel{}).
		Where("id = ?", id).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"deleted_by": domain.ActorFrom(ctx),
		})
	if result.Error != nil {
		return userErrors.wrap("delete user", result.Error)
	}
	if result.RowsAffected == 0 {
		return userErrors.wrap("delete user", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := dbFrom(ctx, r.db).Model(&model.UserModel{}).
		Where("email = ?", strings.ToLower(email)).
		Limit(1).Count(&count).Error
	if err != nil {
		return false, userErrors.wrap("check email exists", err)
	}
	return count > 0, nil
}

// UpdatePassword is separate from Update so the PATCH endpoint cannot reach these columns.
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	result := dbFrom(ctx, r.db).
		Model(&model.UserModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"password_hash":       passwordHash,
			"password_changed_at": time.Now(),
		})
	if result.Error != nil {
		return userErrors.wrap("update password", result.Error)
	}
	if result.RowsAffected == 0 {
		return userErrors.wrap("update password", gorm.ErrRecordNotFound)
	}
	return nil
}

// IncrementFailedLogin returns the new total; one statement so parallel attempts cannot share it.
func (r *userRepository) IncrementFailedLogin(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `
		UPDATE users SET failed_login_attempts = failed_login_attempts + 1, updated_at = now()
		WHERE id = ? AND deleted_at IS NULL
		RETURNING failed_login_attempts`

	var row struct{ FailedLoginAttempts int }
	if err := dbFrom(ctx, r.db).Raw(q, userID).Scan(&row).Error; err != nil {
		return 0, userErrors.wrap("increment failed login", err)
	}
	return row.FailedLoginAttempts, nil
}

// LockUntil closes the account to sign-ins until the given time.
func (r *userRepository) LockUntil(ctx context.Context, userID uuid.UUID, until time.Time) error {
	err := dbFrom(ctx, r.db).
		Model(&model.UserModel{}).
		Where("id = ?", userID).
		Update("locked_until", until).Error
	if err != nil {
		return userErrors.wrap("lock user", err)
	}
	return nil
}

// ClearFailedLogins wipes the counter after a successful login.
func (r *userRepository) ClearFailedLogins(ctx context.Context, userID uuid.UUID) error {
	err := dbFrom(ctx, r.db).
		Model(&model.UserModel{}).
		Where("id = ?", userID).
		Updates(map[string]any{"failed_login_attempts": 0, "locked_until": nil}).Error
	if err != nil {
		return userErrors.wrap("clear failed logins", err)
	}
	return nil
}
