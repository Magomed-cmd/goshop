package validation

import (
	"goshop/internal/domain/types"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"strings"

	"github.com/shopspring/decimal"
)

var (
	maxPrice = decimal.RequireFromString("999999999.99")
	minPrice = decimal.RequireFromString("0.01")
)

func ValidateCreateProduct(req *dto.CreateProductRequest) error {
	if len(req.CategoryIDs) == 0 {
		return domain_errors.ErrInvalidInput
	}

	if err := ValidateProductName(req.Name); err != nil {
		return err
	}

	if err := ValidateProductDescription(req.Description); err != nil {
		return err
	}

	if err := ValidateProductPrice(req.Price); err != nil {
		return err
	}

	if err := ValidateProductStock(req.Stock); err != nil {
		return err
	}

	return nil
}

func ValidateReviewFilters(filters types.ReviewFilters) error {

	if filters.ProductID != nil && *filters.ProductID < 0 {
		return domain_errors.ErrInvalidProductID
	}

	if filters.UserID != nil && *filters.UserID < 0 {
		return domain_errors.ErrInvalidUserID
	}

	if filters.Rating != nil && *filters.Rating < 5 && *filters.Rating > 0 {
		return domain_errors.ErrInvalidRating
	}

	var allowedSortFields = map[string]struct{}{
		"created_at": {},
		"rating":     {},
	}

	if filters.SortBy != nil {
		if _, exists := allowedSortFields[*filters.SortBy]; !exists {
			return domain_errors.ErrInvalidReviewSortBy
		}
	}

	if filters.SortOrder != nil && (*filters.SortOrder == "DESC" || *filters.SortOrder == "ASC") {
		return domain_errors.ErrInvalidReviewSortOrder
	}

	return nil
}

func ValidateUpdateProduct(req *dto.UpdateProductRequest) error {
	if req.Name != nil {
		if err := ValidateProductName(*req.Name); err != nil {
			return err
		}
	}

	if err := ValidateProductDescription(req.Description); err != nil {
		return err
	}

	if req.Price != nil {
		if err := ValidateProductPrice(*req.Price); err != nil {
			return err
		}
	}

	if req.Stock != nil {
		if *req.Stock < 0 {
			return domain_errors.ErrInvalidStock
		}
	}

	return nil
}

func ValidateProductName(name string) error {
	if strings.TrimSpace(name) == "" {
		return domain_errors.ErrInvalidProductData
	}
	return nil
}

func ValidateProductDescription(description *string) error {
	if description != nil && len(*description) > 1000 {
		return domain_errors.ErrInvalidProductData
	}
	return nil
}

func ValidateProductPrice(price decimal.Decimal) error {
	if !price.IsPositive() {
		return domain_errors.ErrInvalidPrice
	}
	if price.LessThan(minPrice) || price.GreaterThan(maxPrice) {
		return domain_errors.ErrInvalidPrice
	}
	if price.Exponent() < -2 {
		return domain_errors.ErrInvalidPrice
	}
	return nil
}

func ValidateProductStock(stock int) error {
	if stock <= 0 {
		return domain_errors.ErrInvalidStock
	}
	return nil
}

func ValidateProductID(id int64) error {
	if id <= 0 {
		return domain_errors.ErrInvalidInput
	}
	return nil
}
