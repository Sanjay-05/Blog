package handler

import (
	"blog/helper"
	"blog/request"
	"blog/response"
	"blog/service"
	"blog/utility"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CategoryHandler handles category-related requests
type CategoryHandler interface {
	CreateCategory(c fiber.Ctx) error
	GetCategory(c fiber.Ctx) error
	GetAllCategories(c fiber.Ctx) error
	GetAllCategoriesWithCount(c fiber.Ctx) error
	UpdateCategory(c fiber.Ctx) error
	DeleteCategory(c fiber.Ctx) error
}

type categoryHandler struct {
	categoryService service.CategoryService
	logger          *logrus.Logger
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(cs service.CategoryService, logger *logrus.Logger) CategoryHandler {
	return &categoryHandler{categoryService: cs, logger: logger}
}

// CreateCategory handles category creation
func (h *categoryHandler) CreateCategory(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to CreateCategory")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	var req request.CreateCategoryRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to bind create category request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		h.logger.Error("validation failed for create category", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	category, err := h.categoryService.CreateCategory(userID, req)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to create category", err)
		return utility.SendErrorResponse(c, statusCode, "failed to create category", err)
	}

	h.logger.Info("category created successfully", "category_id", category.ID)
	return utility.SendSuccessResponse(c, http.StatusCreated, "category created successfully", nil)
}

// GetCategory handles getting a single category
func (h *categoryHandler) GetCategory(c fiber.Ctx) error {
	categoryID := c.Params("id")
	id, err := uuid.Parse(categoryID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid category id", err)
	}

	category, err := h.categoryService.GetCategoryByID(id)
	if err != nil {
		h.logger.Error("failed to get category", err)
		return utility.SendErrorResponse(c, http.StatusNotFound, "category not found", err)
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "category fetched successfully", response.CategoryResponse{
		ID:           category.ID,
		CategoryName: category.CategoryName,
		Description:  category.Description,
		CreatedAt:    category.CreatedAt,
	})
}

// GetAllCategories handles getting all categories
func (h *categoryHandler) GetAllCategories(c fiber.Ctx) error {
	limit, offset := utility.GetPagination(c)

	categories, total, err := h.categoryService.GetAllCategories(limit, offset)
	if err != nil {
		h.logger.Error("failed to get categories", err)
		return utility.SendErrorResponse(c, http.StatusInternalServerError, "failed to get categories", err)
	}

	if len(categories) == 0 {
		return utility.SendSuccessResponse(c, http.StatusOK, "no categories found", nil)
	}

	var categoryResponses []response.CategoryResponse
	for _, cat := range categories {
		categoryResponses = append(categoryResponses, response.CategoryResponse{
			ID:           cat.ID,
			CategoryName: cat.CategoryName,
			Description:  cat.Description,
			CreatedAt:    cat.CreatedAt,
		})
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "categories retrieved sucessfully", map[string]any{
		"total":      total,
		"limit":      limit,
		"offset":     offset,
		"categories": categoryResponses,
	})
}

func (h *categoryHandler) GetAllCategoriesWithCount(c fiber.Ctx) error {
	categories, err := h.categoryService.GetAllCategoriesWithCount()
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusInternalServerError, "failed to fetch categories with counts", err)
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "categories fetched successfully", categories)
}

// UpdateCategory handles category update
func (h *categoryHandler) UpdateCategory(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to UpdateCategory")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	var req request.UpdateCategoryRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to bind update category request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "failed to update category", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	role, _ := c.Locals("role").(string)
	isAdmin := role == "admin"

	category, err := h.categoryService.UpdateCategory(userID, req.CategoryID, req, isAdmin)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("failed to update category", err)
		return utility.SendErrorResponse(c, statusCode, "failed to update category", err)
	}

	h.logger.Info("category updated successfully", "category_id", category.ID)
	return utility.SendSuccessResponse(c, http.StatusOK, "Category updated successfully", response.CategoryResponse{
		ID:           category.ID,
		CategoryName: category.CategoryName,
		Description:  category.Description,
		CreatedAt:    category.CreatedAt,
	})
}

// DeleteCategory handles category deletion
func (h *categoryHandler) DeleteCategory(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to DeleteCategory")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	categoryID := c.Params("id")
	id, err := uuid.Parse(categoryID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid category id", err)
	}

	err = h.categoryService.DeleteCategory(userID, id)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to delete category", err)
		return utility.SendErrorResponse(c, statusCode, "failed to delete category", err)
	}

	h.logger.Info("Category deleted successfully", "category_id", id)
	return utility.SendSuccessResponse(c, http.StatusOK, "Category deleted successfully", nil)
}
