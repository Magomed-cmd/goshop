package mappers

import (
	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
)

func ToCategoryResponse(category *entities.Category) dto.CategoryResponse {
	if category == nil {
		return dto.CategoryResponse{}
	}

	return dto.CategoryResponse{
		ID:           category.ID,
		UUID:         category.UUID.String(),
		Name:         category.Name,
		Description:  category.Description,
		CreatedAt:    category.CreatedAt,
		UpdatedAt:    category.UpdatedAt,
		ProductCount: 0,
	}
}

func ToCategoryResponses(categories []*entities.Category) []dto.CategoryResponse {
	if len(categories) == 0 {
		return []dto.CategoryResponse{}
	}

	responses := make([]dto.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		resp := ToCategoryResponse(category)
		responses = append(responses, resp)
	}

	return responses
}

func ToCategoriesListResponse(categories []*entities.CategoryWithCount) *dto.CategoriesListResponse {
	resp := &dto.CategoriesListResponse{
		Categories: make([]dto.CategoryResponse, 0, len(categories)),
		Total:      len(categories),
	}

	for _, categoryWithCount := range categories {
		category := categoryWithCount
		response := ToCategoryResponse(&category.Category)
		response.ProductCount = int(category.ProductCount)
		resp.Categories = append(resp.Categories, response)
	}

	return resp
}
