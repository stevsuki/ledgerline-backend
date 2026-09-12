package model

import "github.com/google/uuid"

type MasterCategoryModel struct {
	// The column is UUID, as the seed's pinned ids are; it was declared int here,
	// which only went unnoticed because List reads into the domain struct.
	ID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name  string    `gorm:"size:100;not null;uniqueIndex"`
	Type  string    `gorm:"size:20;not null;default:expense"`
	Icon  string    `gorm:"size:50;not null;default:''"`
	Color string    `gorm:"size:10;not null;default:''"`
}

func (MasterCategoryModel) TableName() string {
	return "master_categories"
}
