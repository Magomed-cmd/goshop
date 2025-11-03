package httpadapter

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	httpErrors "goshop/internal/adapters/input/http/errors"
	"goshop/internal/core/domain/entities"
	"goshop/internal/core/mappers"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
)

type CategoryHandler struct {
	categoryService serviceports.CategoryService
}

func NewCategoryHandler(s serviceports.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: s,
	}
}

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {

	ctx := c.Request.Context()

	resp, err := h.categoryService.GetAllCategories(ctx)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	category, err := h.categoryService.GetCategoryByID(ctx, id)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	createdCategory, err := h.categoryService.CreateCategory(ctx, &req)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	cur, err := h.categoryService.GetCategoryByID(ctx, id)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	updated, err := h.categoryService.UpdateCategory(ctx, entity)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	resp := mappers.ToCategoryResponse(updated)
	resp.ProductCount = cur.ProductCount

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

	err = h.categoryService.DeleteCategory(ctx, id)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(204, nil)
}
