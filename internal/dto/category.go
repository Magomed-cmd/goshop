package dto

type CreateCategoryRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
}

type CategoryResponse struct {
	ID           int64   `json:"id"`
	UUID         string  `json:"uuid"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	ProductCount int     `json:"product_count"`
}

type CategoriesListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int                `json:"total"`
}
