package dto

import "github.com/shopspring/decimal"

type CreateProductRequest struct {
	Name        string          `json:"name" binding:"required,min=2,max=200"`
	Description *string         `json:"description" binding:"omitempty,max=1000"`
	Price       decimal.Decimal `json:"price" binding:"required"`
	Stock       int             `json:"stock" binding:"min=0"`
	CategoryIDs []int64         `json:"category_ids" binding:"required,min=1"`
}

type UpdateProductRequest struct {
	Name        *string          `json:"name" binding:"omitempty,min=2,max=200"`
	Description *string          `json:"description" binding:"omitempty,max=1000"`
	Price       *decimal.Decimal `json:"price" binding:"omitempty"`
	Stock       *int             `json:"stock" binding:"omitempty,min=0"`
	CategoryIDs []int64          `json:"category_ids" binding:"omitempty"`
}

type ProductResponse struct {
	ID          int64              `json:"id"`
	UUID        string             `json:"uuid"`
	Name        string             `json:"name"`
	Description *string            `json:"description"`
	Price       string             `json:"price"`
	Stock       int                `json:"stock"`
	Categories  []CategoryResponse `json:"categories"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

type ProductsListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

type ProductCatalogItem struct {
	ID         int64              `json:"id"`
	UUID       string             `json:"uuid"`
	Name       string             `json:"name"`
	Price      string             `json:"price"`
	Stock      int                `json:"stock"`
	Categories []CategoryResponse `json:"categories"`
}

type ProductCatalogResponse struct {
	Products []ProductCatalogItem `json:"products"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	Limit    int                  `json:"limit"`
}
