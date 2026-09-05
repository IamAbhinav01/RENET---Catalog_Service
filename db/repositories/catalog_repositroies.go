package repositories

import (
	"renet-catalog/models"

	"gorm.io/gorm"
)


type CatalogRepository interface {
	GetByID(id int) (*models.Item, error)
	ListItems(page, limit int) ([]models.Item, int64, error)
	SearchItems(query string, limit int) ([]models.Item, error)
	UpdatePosterAndPlot(id int, posterURL, plot string) error
	CreateInteraction(*models.Interaction) error
	GetByUserIDInteraction(userID, limit int) ([]models.Interaction, error)
}

type CatalogRepositoryImpl struct {
	db *gorm.DB
}

func NewCatalogRepository(db *gorm.DB) CatalogRepository {
	return &CatalogRepositoryImpl{
		db: db,
	}
}

// 1. Get item by ID
func (repo *CatalogRepositoryImpl) GetByID(id int) (*models.Item, error) {
	var item models.Item
	err := repo.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
// 2. List items with pagination
func (repo *CatalogRepositoryImpl) ListItems(page, limit int) ([]models.Item, int64, error) {
	var items []models.Item
	var total int64

	if err := repo.db.Model(&models.Item{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := repo.db.Order("id ASC").Offset(offset).Limit(limit).Find(&items).Error

	return items, total, err
}
// 3. Search items by title (Case-insensitive)
func (repo *CatalogRepositoryImpl) SearchItems(query string, limit int) ([]models.Item, error) {
	var items []models.Item
	err := repo.db.Where("LOWER(title) LIKE LOWER(?)", "%"+query+"%").
		Order("id ASC").
		Limit(limit).
		Find(&items).Error

	return items, err
}
// 4. Update poster URL and plot (e.g., from OMDb enrichment)
func (repo *CatalogRepositoryImpl) UpdatePosterAndPlot(id int, posterURL, plot string) error {
	return repo.db.Model(&models.Item{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"poster_url": posterURL,
			"plot":       plot,
		}).Error
}
//5. create interaction logic 
func(repo *CatalogRepositoryImpl)CreateInteraction(interactions *models.Interaction) error{
	return repo.db.Create(interactions).Error
}
//6. detals of users based on interactions
func(repo *CatalogRepositoryImpl)GetByUserIDInteraction(userID, limit int) ([]models.Interaction, error){
	var list []models.Interaction
	err := repo.db.Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Find(&list).Error
	return list, err
}
