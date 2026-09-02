package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"github.com/stevensuki/ledgerline-backend/internal/repository/postgres/model"
)

type passwordResetTokenRepository struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) domain.PasswordResetTokenRepository {
	return &passwordResetTokenRepository{db: db}
}

func (r *passwordResetTokenRepository) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	row := model.PasswordResetTokenFromDomain(token)
	if err := dbFrom(ctx, r.db).Create(&row).Error; err != nil {
		return passwordResetTokenErrors.wrap("create password reset token", err)
	}

	// Columns with database defaults come back via RETURNING.
	token.Attempts = row.Attempts
	token.CreatedAt = row.CreatedAt
	return nil
}

func (r *passwordResetTokenRepository) Update(ctx context.Context, payload *domain.PasswordResetToken) error {
	result := dbFrom(ctx, r.db).Model(&model.PasswordResetTokenModel{}).Where("id = ?", payload.ID).Updates(map[string]any{
		"attempts": payload.Attempts,
		"used_at":  payload.UsedAt,
	})
	if result.Error != nil {
		return passwordResetTokenErrors.wrap("update password reset token", result.Error)
	}
	if result.RowsAffected == 0 {
		return passwordResetTokenErrors.wrap("update password reset token", gorm.ErrRecordNotFound)
	}
	return nil
}

// DeleteActiveByUserID removes tokens when forgot password
func (r *passwordResetTokenRepository) DeleteActiveByUserID(ctx context.Context, userID uuid.UUID) error {
	// The model has no DeletedAt field, so this is a hard delete.
	err := dbFrom(ctx, r.db).
		Where("user_id = ?", userID).
		Delete(&model.PasswordResetTokenModel{}).Error
	if err != nil {
		return passwordResetTokenErrors.wrap("delete password reset token", err)
	}

	return nil
}

func (r *passwordResetTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.PasswordResetToken, error) {
	var row model.PasswordResetTokenModel
	err := dbFrom(ctx, r.db).Where("user_id = ?", userID).First(&row).Error
	if err != nil {
		return nil, passwordResetTokenErrors.wrap("get password reset token by user ID", err)
	}

	return row.ToDomain(), nil
}
