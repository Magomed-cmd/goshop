package types

import (
	"github.com/shopspring/decimal"
	"time"
)

type ProductFilters struct {
	Page       int              // Номер страницы (1, 2, 3...) для пагинации
	Limit      int              // Сколько товаров показать на одной странице (20, 50, 100)
	CategoryID *int64           // Фильтр по категории (nil = все категории, 5 = только категория с ID=5)
	SortBy     *string          // По какому полю сортировать (nil = дефолт, "price"/"name"/"created_at")
	SortOrder  *string          // Направление сортировки (nil = дефолт, "asc"/"desc")
	MinPrice   *decimal.Decimal // Минимальная цена (nil = без ограничения снизу, 1000 = от 1000₽)
	MaxPrice   *decimal.Decimal // Максимальная цена (nil = без ограничения сверху, 50000 = до 50000₽)
}

type OrderFilters struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`

	// Фильтры по статусу
	Status *string `json:"status,omitempty"` // "pending", "paid", "shipped"

	// Фильтры по времени
	DateFrom *time.Time `json:"date_from,omitempty"` // Заказы с этой даты
	DateTo   *time.Time `json:"date_to,omitempty"`   // Заказы до этой даты

	// Фильтры по сумме
	MinAmount *decimal.Decimal `json:"min_amount,omitempty"` // От 1000₽
	MaxAmount *decimal.Decimal `json:"max_amount,omitempty"` // До 50000₽

	// Сортировка
	SortBy    *string `json:"sort_by,omitempty"`    // "created_at", "total_price", "status"
	SortOrder *string `json:"sort_order,omitempty"` // "asc", "desc"
}
