package httpadapter

import (
	"strconv"

	"github.com/gin-gonic/gin"

	httpErrors "goshop/internal/adapters/input/http/errors"
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

// GetAllCategories godoc
// @Summary     Get categories
// @Description Returns the full list of categories with product counts
// @Tags        categories
// @Produce     json
// @Success     200 {object} dto.CategoriesListResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /categories [get]
func (h *CategoryHandler) GetAllCategories(c *gin.Context) {

	ctx := c.Request.Context()

	resp, err := h.categoryService.GetAllCategories(ctx)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, resp)
}

// GetCategoryByID godoc
// @Summary     Get category by ID
// @Description Returns category details for the provided identifier
// @Tags        categories
// @Produce     json
// @Param       id path int true "Category ID"
// @Success     200 {object} dto.CategoryResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /categories/{id} [get]
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

	c.JSON(200, category)
}

// CreateCategory godoc
// @Summary     Create category
// @Description Creates a new category
// @Tags        admin/categories
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateCategoryRequest true "Category data"
// @Success     201 {object} dto.CategoryResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/categories [post]
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

	c.JSON(201, createdCategory)
}

// UpdateCategory godoc
// @Summary     Update category
// @Description Updates category fields for the provided identifier
// @Tags        admin/categories
// @Accept      json
// @Produce     json
// @Param       id path int true "Category ID"
// @Param       request body dto.UpdateCategoryRequest true "Category data"
// @Success     200 {object} dto.CategoryResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/categories/{id} [put]
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

	updated, err := h.categoryService.UpdateCategory(ctx, id, &req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, updated)
}

// DeleteCategory godoc
// @Summary     Delete category
// @Description Deletes category by identifier
// @Tags        admin/categories
// @Produce     json
// @Param       id path int true "Category ID"
// @Success     204 {string} string "No Content"
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/categories/{id} [delete]
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
