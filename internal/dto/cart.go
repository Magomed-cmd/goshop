package dto

type AddToCartRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

type CartResponse struct {
	ID         int64              `json:"id"`
	Items      []CartItemResponse `json:"items"`
	TotalPrice string             `json:"total_price"`
	TotalItems int                `json:"total_items"`
}

type CartItemResponse struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Price       string `json:"price"`
	Subtotal    string `json:"subtotal"`
}
