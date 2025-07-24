package types

import "github.com/shopspring/decimal"

type ProductFilters struct {
	Page       int              // Номер страницы (1, 2, 3...) для пагинации
	Limit      int              // Сколько товаров показать на одной странице (20, 50, 100)
	CategoryID *int64           // Фильтр по категории (nil = все категории, 5 = только категория с ID=5)
	SortBy     *string          // По какому полю сортировать (nil = дефолт, "price"/"name"/"created_at")
	SortOrder  *string          // Направление сортировки (nil = дефолт, "asc"/"desc")
	MinPrice   *decimal.Decimal // Минимальная цена (nil = без ограничения снизу, 1000 = от 1000₽)
	MaxPrice   *decimal.Decimal // Максимальная цена (nil = без ограничения сверху, 50000 = до 50000₽)
}
