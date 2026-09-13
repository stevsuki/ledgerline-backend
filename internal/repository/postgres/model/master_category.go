package model

import "github.com/google/uuid"

type MasterCategoryModel struct {
	ID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name  string    `gorm:"size:100;not null;uniqueIndex"`
	Type  string    `gorm:"size:20;not null;default:expense"`
	Icon  string    `gorm:"size:50;not null;default:''"`
	Color string    `gorm:"size:10;not null;default:''"`
}

func (MasterCategoryModel) TableName() string {
	return "master_categories"
}
