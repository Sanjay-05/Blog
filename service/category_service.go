package service

import (
	"blog/models"
	"blog/repository"
	"blog/request"
	"blog/response"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CategoryService defines category business logic
type CategoryService interface {
	CreateCategory(userID uuid.UUID, req request.CreateCategoryRequest) (*models.Category, error)
	GetCategoryByID(id uuid.UUID) (*models.Category, error)
	GetAllCategories(limit, offset int) ([]models.Category, int64, error)
	GetAllCategoriesWithCount() ([]response.CategoryWithCountResponse, error)
	UpdateCategory(userID uuid.UUID, id uuid.UUID, req request.UpdateCategoryRequest, isAdmin bool) (*models.Category, error)
	DeleteCategory(userID uuid.UUID, id uuid.UUID) error
}

type categoryService struct {
	repo   repository.CategoryRepository
	logger *logrus.Logger
}

// NewCategoryService creates a new category service
func NewCategoryService(repo repository.CategoryRepository, logger *logrus.Logger) CategoryService {
	return &categoryService{repo: repo, logger: logger}
}

// CreateCategory creates a new category
func (s *categoryService) CreateCategory(userID uuid.UUID, req request.CreateCategoryRequest) (*models.Category, error) {
	category, err := s.repo.GetCategoryByName(req.CategoryName)
	if err == nil {
		return nil, ErrUniqueViolation
	}

	category = &models.Category{
		CategoryName: req.CategoryName,
		Description:  req.Description,
		UserID:       userID,
		Status:       "active",
	}

	if err := s.repo.CreateCategory(category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategoryByID gets category by ID
func (s *categoryService) GetCategoryByID(id uuid.UUID) (*models.Category, error) {
	return s.repo.GetCategoryByID(id)
}

// GetAllCategories gets all categories
func (s *categoryService) GetAllCategories(limit, offset int) ([]models.Category, int64, error) {
	return s.repo.GetAllCategories(limit, offset)
}

func (s *categoryService) GetAllCategoriesWithCount() ([]response.CategoryWithCountResponse, error) {
	return s.repo.GetAllCategoriesWithCount()
}

// UpdateCategory updates a category
func (s *categoryService) UpdateCategory(userID uuid.UUID, id uuid.UUID, req request.UpdateCategoryRequest, isAdmin bool) (*models.Category, error) {
	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		s.logger.Warn("Failed to update category - category not found", "category_id", id)
		return nil, ErrNotFound
	}

	// Check ownership unless user is admin
	if !isAdmin && category.UserID != userID {
		s.logger.Warn("Failed to update category - unauthorized user", "category_id", id, "user_id", userID)
		return nil, ErrOwnershipViolation
	}

	updates := make(map[string]interface{})
	if req.CategoryName != nil {
		existingCat, err := s.repo.GetCategoryByName(*req.CategoryName)
		if err == nil && existingCat.ID != id {
			if existingCat.Status == "active" {
				s.logger.Warn("Failed to update category - name already exists", "category_name", *req.CategoryName)
				return nil, ErrUniqueViolation
			}
		}
		updates["category_name"] = *req.CategoryName
	}

	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if err := s.repo.UpdateCategory(id, updates); err != nil {
		s.logger.Error("Failed to update category", err)
		return nil, err
	}

	s.logger.Info("Category updated successfully", "category_id", id)
	return s.repo.GetCategoryByID(id)
}

// DeleteCategory deletes a category
func (s *categoryService) DeleteCategory(userID uuid.UUID, id uuid.UUID) error {
	category, err := s.repo.GetCategoryByID(id)
	if err != nil {
		s.logger.Warn("Failed to delete category - category not found", "category_id", id)
		return ErrNotFound
	}

	if category.UserID != userID {
		s.logger.Warn("Failed to delete category - unauthorized user", "category_id", id, "user_id", userID)
		return ErrOwnershipViolation
	}

	if err := s.repo.DeleteCategory(id); err != nil {
		s.logger.Error("Failed to delete category", err)
		return err
	}

	s.logger.Info("Category deleted successfully", "category_id", id)
	return nil
}
