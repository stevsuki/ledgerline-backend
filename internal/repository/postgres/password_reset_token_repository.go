package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

type passwordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) domain.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	model := passwordResetTokenFromDomain(token)
	if err := dbFrom(ctx, r.db).Create(&model).Error; err != nil {
		return wrapErr("create password reset token", err)
	}

	// Columns with database defaults come back via RETURNING.
	token.Attempts = model.Attempts
	token.CreatedAt = model.CreatedAt
	return nil
}

func (r *passwordResetTokenRepository) Update(ctx context.Context, payload *domain.PasswordResetToken) error {
	result := dbFrom(ctx, r.db).Model(&passwordResetTokenModel{}).Where("id = ?", payload.ID).Updates(map[string]any{
		"attempts": payload.Attempts,
		"used_at":  payload.UsedAt,
	})
	if result.Error != nil {
		return wrapErr("update password reset token", result.Error)
	}
	if result.RowsAffected == 0 {
		return wrapErr("update password reset token", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *passwordResetTokenRepository) ExistsByUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	var count int64
	if err := dbFrom(ctx, r.db).Model(&passwordResetTokenModel{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return false, wrapErr("check password reset token existence", err)
	}
	return count > 0, nil
}

// DeleteActiveByUserID removes tokens when forgot password
func (r *passwordResetTokenRepository) DeleteActiveByUserID(ctx context.Context, userID uuid.UUID) error {
	// The model has no DeletedAt field, so this is a hard delete.
	err := dbFrom(ctx, r.db).
		Where("user_id = ?", userID).
		Delete(&passwordResetTokenModel{}).Error
	if err != nil {
		return wrapErr("delete password reset token", err)
	}

	return nil
}

func (r *passwordResetTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PasswordResetToken, error) {
	var model passwordResetTokenModel
	err := dbFrom(ctx, r.db).Where("user_id = ?", userID).First(&model).Error
	if err != nil {
		return nil, wrapErr("get password reset token by user ID", err)
	}

	return model.toDomain(), nil
}
