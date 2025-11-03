package httpadapter

import (
	"context"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"goshop/internal/core/domain/entities"
	errors2 "goshop/internal/core/domain/errors"
	"goshop/internal/dto"
)

type CategoryService interface {
	GetAllCategories(ctx context.Context) (*dto.CategoriesListResponse, error)
	GetCategoryByID(ctx context.Context, id int64) (*dto.CategoryResponse, error)
	CreateCategory(ctx context.Context, req *dto.CreateCategoryRequest) (*entities.Category, error)
	UpdateCategory(ctx context.Context, category *entities.Category) (*entities.Category, error)
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

	resp, err := h.CategoryService.GetAllCategories(ctx)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(200, resp)
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
		if errors.Is(err, errors2.ErrCategoryNotFound) {
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
		if errors.Is(err, errors2.ErrInvalidInput) {
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

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "Invalid category ID"})
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	cur, err := h.CategoryService.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, errors2.ErrCategoryNotFound) {
			c.JSON(404, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(500, gin.H{"error": "Failed to fetch category"})
		return
	}

	name := cur.Name
	if req.Name != nil {
		name = *req.Name
	}
	desc := cur.Description
	if req.Description != nil {
		desc = req.Description
	}

	entity := &entities.Category{
		ID:          cur.ID,
		UUID:        uuid.MustParse(cur.UUID),
		Name:        name,
		Description: desc,
		CreatedAt:   cur.CreatedAt,
		UpdatedAt:   cur.UpdatedAt,
	}

	updated, err := h.CategoryService.UpdateCategory(ctx, entity)
	if err != nil {
		switch {
		case errors.Is(err, errors2.ErrCategoryNotFound):
			c.JSON(404, gin.H{"error": "Category not found"})
		case errors.Is(err, errors2.ErrInvalidInput),
			errors.Is(err, errors2.ErrInvalidCategoryData):
			c.JSON(400, gin.H{"error": "Invalid category data"})
		default:
			c.JSON(500, gin.H{"error": "Failed to update category"})
		}
		return
	}

	resp := dto.CategoryResponse{
		ID:           updated.ID,
		UUID:         updated.UUID.String(),
		Name:         updated.Name,
		Description:  updated.Description,
		CreatedAt:    updated.CreatedAt,
		UpdatedAt:    updated.UpdatedAt,
		ProductCount: cur.ProductCount,
	}
	c.JSON(200, resp)
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
		if errors.Is(err, errors2.ErrCategoryNotFound) {
			c.JSON(404, gin.H{"error": "Category not found"})
			return
		}
		if errors.Is(err, errors2.ErrInvalidInput) {
			c.JSON(400, gin.H{"error": "Invalid category ID"})
			return
		}
		c.JSON(500, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(204, nil)
}
