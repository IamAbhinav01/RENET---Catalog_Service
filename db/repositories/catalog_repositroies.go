package repositories

import "gorm.io/gorm"

type CatalogRepository interface {
	GetByID()
	List_Items()
	Search_Items()
	Update_Poster_and_Plot()
}

type CatalogRepositoryImpl struct{
	db *gorm.DB
}

func NewCatalogRepository(_db *gorm.DB) *CatalogRepositoryImpl{
	return &CatalogRepositoryImpl{
		db: _db,
	}
}

