package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Category struct {
	ID          int64     `db:"id" json:"id"`
	UUID        uuid.UUID `db:"uuid" json:"uuid"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type CategoryWithCount struct {
	Category
	ProductCount int64
}

type Product struct {
	ID          int64           `db:"id" json:"id"`
	UUID        uuid.UUID       `db:"uuid" json:"uuid"`
	Name        string          `db:"name" json:"name"`
	Description *string         `db:"description" json:"description"`
	Price       decimal.Decimal `db:"price" json:"price"`
	Stock       int             `db:"stock" json:"stock"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}

type ProductCategory struct {
	ProductID  int64 `db:"product_id" json:"product_id"`
	CategoryID int64 `db:"category_id" json:"category_id"`
}

type ProductImage struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	ImageURL  string    `json:"image_url"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UUID      uuid.UUID `json:"uuid"`
}
