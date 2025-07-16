package models

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
	Categories  []Category      `db:"-" json:"categories,omitempty"`
}

type ProductCategory struct {
	ProductID  int64 `db:"product_id" json:"product_id"`
	CategoryID int64 `db:"category_id" json:"category_id"`
}
