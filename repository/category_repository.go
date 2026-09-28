package repository

import (
	"blog/models"
	"blog/response"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryRepository defines category data operations
type CategoryRepository interface {
	CreateCategory(category *models.Category) error
	GetCategoryByID(id uuid.UUID) (*models.Category, error)
	GetCategoryByName(name string) (*models.Category, error)
	GetAllCategories(limit, offset int) ([]models.Category, int64, error)
	GetAllCategoriesWithCount() ([]response.CategoryWithCountResponse, error)
	UpdateCategory(id uuid.UUID, updates map[string]interface{}) error
	DeleteCategory(id uuid.UUID) error
}

type categoryRepo struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepo{db: db}
}

// CreateCategory creates a new category
func (r *categoryRepo) CreateCategory(category *models.Category) error {
	return r.db.Create(category).Error
}

// GetCategoryByID gets category by ID
func (r *categoryRepo) GetCategoryByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category
	if err := r.db.Where("id = ?", id).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// GetCategoryByName gets category by name
func (r *categoryRepo) GetCategoryByName(name string) (*models.Category, error) {
	var category models.Category
	if err := r.db.Where("category_name = ?", name).First(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// GetAllCategories gets all categories with pagination
func (r *categoryRepo) GetAllCategories(limit, offset int) ([]models.Category, int64, error) {
	var categories []models.Category
	var total int64

	if err := r.db.Model(&models.Category{}).Where("status = ?", "active").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Where("status = ?", "active").Limit(limit).Offset(offset).Find(&categories).Error; err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *categoryRepo) GetAllCategoriesWithCount() ([]response.CategoryWithCountResponse, error) {
	var results []response.CategoryWithCountResponse

	err := r.db.
		Table("categories").
		Select(`
		categories.id,
		categories.category_name,
		categories.description,
		categories.created_at,
		COUNT(posts.id) AS count
	`).
		Joins(`
		LEFT JOIN posts 
		ON posts.category_id = categories.id 
		AND posts.status = 'published'
	`).
		Where("categories.status = ?", "active").
		Group("categories.id, categories.category_name, categories.description, categories.created_at").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// UpdateCategory updates category information
func (r *categoryRepo) UpdateCategory(id uuid.UUID, updates map[string]interface{}) error {
	result := r.db.Model(&models.Category{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return result.Error
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to update category")
	}
	return nil
}

// DeleteCategory deletes a category
func (r *categoryRepo) DeleteCategory(id uuid.UUID) error {
	now := time.Now()
	result := r.db.Model(&models.Category{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "deleted",
			"deleted_at": &now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to delete category")
	}
	return nil
}
