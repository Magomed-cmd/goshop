package types

import (
	"time"

	"github.com/shopspring/decimal"
)

type ProductFilters struct {
	Page       int              `form:"page,default=1"`   // Номер страницы для пагинации
	Limit      int              `form:"limit,default=20"` // Количество товаров на странице
	CategoryID *int64           `form:"category_id"`      // Фильтр по категории (опционально)
	SortBy     *string          `form:"sort_by"`          // Поле сортировки (price/name/created_at)
	SortOrder  *string          `form:"sort_order"`       // Направление сортировки (asc/desc)
	MinPrice   *decimal.Decimal `form:"min_price"`        // Минимальная цена
	MaxPrice   *decimal.Decimal `form:"max_price"`        // Максимальная цена
}

type OrderFilters struct {
	Page  int `form:"page,default=1"`   // Номер страницы
	Limit int `form:"limit,default=10"` // Количество заказов на странице

	Status *string `form:"status"` // "pending", "paid", "shipped", "delivered", "cancelled"

	DateFrom *time.Time `form:"date_from" time_format:"2006-01-02"`
	DateTo   *time.Time `form:"date_to" time_format:"2006-01-02"`

	// Фильтры по сумме
	MinAmount *decimal.Decimal `form:"min_amount"`
	MaxAmount *decimal.Decimal `form:"max_amount"`

	SortBy    *string `form:"sort_by"`
	SortOrder *string `form:"sort_order"`
}

type AdminOrderFilters struct {
	OrderFilters
	UserID *int64 `form:"user_id"`
}

type ReviewFilters struct {
	Page      int     `form:"page,default=1"`
	Limit     int     `form:"limit,default=20"`
	ProductID *int64  `form:"product_id"` // Отзывы для конкретного продукта
	UserID    *int64  `form:"user_id"`    // Отзывы от конкретного пользователя
	Rating    *int    `form:"rating"`     // Фильтр по оценке (1-5)
	SortBy    *string `form:"sort_by"`    // created_at, rating
	SortOrder *string `form:"sort_order"` // asc, desc
}
