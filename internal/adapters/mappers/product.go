package mappers

import (
	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
)

func ToProductResponse(product *entities.Product, categories []*entities.Category, images []*entities.ProductImage) *dto.ProductResponse {
	if product == nil {
		return nil
	}

	return &dto.ProductResponse{
		ID:          product.ID,
		UUID:        product.UUID.String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       decimalToString(product.Price),
		Stock:       product.Stock,
		ProductImgs: images,
		Categories:  ToCategoryResponses(categories),
		CreatedAt:   formatTime(product.CreatedAt),
		UpdatedAt:   formatTime(product.UpdatedAt),
	}
}

func ToProductCatalogResponse(products []*entities.Product, total int, page, limit int) *dto.ProductCatalogResponse {
	items := make([]dto.ProductCatalogItem, 0, len(products))
	for _, product := range products {
		if product == nil {
			continue
		}

		items = append(items, dto.ProductCatalogItem{
			ID:         product.ID,
			UUID:       product.UUID.String(),
			Name:       product.Name,
			Price:      decimalToString(product.Price),
			Stock:      product.Stock,
			Categories: nil,
		})
	}

	return &dto.ProductCatalogResponse{
		Products: items,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}
}
