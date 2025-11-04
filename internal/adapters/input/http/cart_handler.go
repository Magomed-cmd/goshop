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

// GetCart godoc
// @Summary     Get cart
// @Description Returns the current cart for the authenticated user
// @Tags        cart
// @Produce     json
// @Success     200 {object} dto.CartResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/cart [get]
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

// AddItem godoc
// @Summary     Add item to cart
// @Description Adds a product to the authenticated user's cart
// @Tags        cart
// @Accept      json
// @Produce     json
// @Param       request body dto.AddToCartRequest true "Cart item payload"
// @Success     201 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/cart/items [post]
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

// UpdateItem godoc
// @Summary     Update cart item
// @Description Updates the quantity of a product in the authenticated user's cart
// @Tags        cart
// @Accept      json
// @Produce     json
// @Param       product_id path int true "Product ID"
// @Param       request body dto.UpdateCartItemRequest true "Quantity payload"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/cart/items/{product_id} [put]
func (h *CartHandler) UpdateItem(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(400, gin.H{"error": "Invalid product ID"})
		return
	}

	var req dto.UpdateCartItemRequest
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

// RemoveItem godoc
// @Summary     Remove cart item
// @Description Removes a product from the authenticated user's cart
// @Tags        cart
// @Produce     json
// @Param       product_id path int true "Product ID"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/cart/items/{product_id} [delete]
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

// ClearCart godoc
// @Summary     Clear cart
// @Description Removes all products from the authenticated user's cart
// @Tags        cart
// @Produce     json
// @Success     200 {object} map[string]string
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/cart [delete]
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
