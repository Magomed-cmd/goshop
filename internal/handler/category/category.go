package category

import (
	"context"
	"errors"
	"goshop/internal/domain/entities"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"strconv"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) ([]*entities.CategoryWithCount, error)
	GetCategoryByID(ctx context.Context, id int64) (*dto.CategoryResponse, error)
	CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*entities.Category, error)
	UpdateCategory(ctx context.Context, category *entities.Category) error
	DeleteCategory(ctx context.Context, id int64) error
}

type CategoryHandler struct {
	CategoryService CategoryService
}

func NewCategoryHandler(s CategoryService) *CategoryHandler {
	return &CategoryHandler{
		CategoryService: s,
	}
}

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.CategoryService.GetAllCategories(ctx)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch categories"})
		return
	}

	var response []dto.CategoryResponse
	for _, category := range categories {
		response = append(response, dto.CategoryResponse{
			ID:           category.Category.ID,
			UUID:         category.Category.UUID.String(),
			Name:         category.Category.Name,
			Description:  category.Category.Description,
			ProductCount: int(category.ProductCount),
		})
	}

	c.JSON(200, response)
}

func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "Invalid category ID"})
		return
	}

	category, err := h.CategoryService.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain_errors.ErrCategoryNotFound) {
			c.JSON(404, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(500, gin.H{"error": "Failed to fetch category"})
		return
	}

	response := dto.CategoryResponse{
		ID:           category.ID,
		UUID:         category.UUID,
		Name:         category.Name,
		Description:  category.Description,
		ProductCount: int(category.ProductCount),
	}

	c.JSON(200, response)
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	createdCategory, err := h.CategoryService.CreateCategory(ctx, &req)
	if err != nil {
		if errors.Is(err, domain_errors.ErrInvalidInput) {
			c.JSON(400, gin.H{"error": "Invalid category data"})
			return
		}
		c.JSON(500, gin.H{"error": "Failed to create category"})
		return
	}

	response := dto.CategoryResponse{
		ID:           createdCategory.ID,
		UUID:         createdCategory.UUID.String(),
		Name:         createdCategory.Name,
		Description:  createdCategory.Description,
		ProductCount: 0,
	}

	c.JSON(201, response)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "Invalid category ID"})
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	category, err := h.CategoryService.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain_errors.ErrCategoryNotFound) {
			c.JSON(404, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(500, gin.H{"error": "Failed to fetch category"})
		return
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = req.Description
	}
	
	entityCategory := &entities.Category{
		ID:          category.ID,
		UUID:        uuid.MustParse(category.UUID),
		Name:        category.Name,
		Description: category.Description,
	}

	err = h.CategoryService.UpdateCategory(ctx, entityCategory)
	if err != nil {
		if errors.Is(err, domain_errors.ErrCategoryNotFound) {
			c.JSON(404, gin.H{"error": "Category not found"})
			return
		}
		if errors.Is(err, domain_errors.ErrInvalidInput) {
			c.JSON(400, gin.H{"error": "Invalid category data"})
			return
		}
		if errors.Is(err, domain_errors.ErrInvalidCategoryData) {
			c.JSON(400, gin.H{"error": "Invalid category data"})
			return
		}
		c.JSON(500, gin.H{"error": "Failed to update category"})
		return
	}

	updatedCategory, err := h.CategoryService.GetCategoryByID(ctx, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch updated category"})
		return
	}

	response := dto.CategoryResponse{
		ID:           updatedCategory.ID,
		UUID:         updatedCategory.UUID,
		Name:         updatedCategory.Name,
		Description:  updatedCategory.Description,
		ProductCount: int(updatedCategory.ProductCount),
	}

	c.JSON(200, response)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "Invalid category ID"})
		return
	}

	err = h.CategoryService.DeleteCategory(ctx, id)
	if err != nil {
		if errors.Is(err, domain_errors.ErrCategoryNotFound) {
			c.JSON(404, gin.H{"error": "Category not found"})
			return
		}
		if errors.Is(err, domain_errors.ErrInvalidInput) {
			c.JSON(400, gin.H{"error": "Invalid category ID"})
			return
		}
		c.JSON(500, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(204, nil)
}
