package httpadapter

import (
	"strconv"

	"github.com/gin-gonic/gin"

	httpErrors "goshop/internal/adapters/input/http/errors"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
)

type CartHandler struct {
	cartService serviceports.CartService
}

func NewCartHandler(cartService serviceports.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	cart, err := h.cartService.GetCart(ctx, userID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, cart)
}

func (h *CartHandler) AddItem(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	var req dto.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.cartService.AddItem(ctx, userID, &req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(201, gin.H{"message": "Item added to cart"})
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(400, gin.H{"error": "Invalid product ID"})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.cartService.UpdateItem(ctx, userID, productID, req.Quantity)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{"message": "Item updated in cart"})
}

func (h *CartHandler) RemoveItem(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(400, gin.H{"error": "Invalid product ID"})
		return
	}

	err = h.cartService.RemoveItem(ctx, userID, productID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{"message": "Item removed from cart"})
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	err := h.cartService.ClearCart(ctx, userID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{"message": "Cart cleared"})
}

