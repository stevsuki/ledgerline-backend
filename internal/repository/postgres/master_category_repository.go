package postgres

import (
	"context"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
	"gorm.io/gorm"
)

type masterCategoryRepository struct {
	db *gorm.DB
}

func NewMasterCategoryRepository(db *gorm.DB) domain.MasterCategoryRepository {
	return &masterCategoryRepository{db: db}
}

func (r *masterCategoryRepository) List(ctx context.Context) ([]domain.MasterCategory, error) {
	var rows []domain.MasterCategory
	if err := dbFrom(ctx, r.db).Find(&rows).Error; err != nil {
		return nil, masterCategoryErrors.wrap("list master categories", err)
	}
	return rows, nil
}
