package validation

import (
	"github.com/shopspring/decimal"
	"goshop/internal/domain_errors"
	"goshop/internal/dto"
	"strings"
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
