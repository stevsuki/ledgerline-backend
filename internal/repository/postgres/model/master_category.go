package model

type MasterCategoryModel struct {
	ID   int    `gorm:"primaryKey"`
	Name string `gorm:"size:100;not null;uniqueIndex"`
}

func (MasterCategoryModel) TableName() string {
	return "master_categories"
}
