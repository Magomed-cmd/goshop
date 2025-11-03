package httpadapter

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"goshop/internal/core/domain/types"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
	"goshop/internal/middleware"
)

type OrderHandler struct {
	service serviceports.OrderService
}

func NewOrderHandler(s serviceports.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {

	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	req := &dto.CreateOrderRequest{}

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request Body"})
		return
	}

	resp, err := h.service.CreateOrder(ctx, userID, req)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {

	ctx := c.Request.Context()
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	filters := types.OrderFilters{}
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	resp, err := h.service.GetUserOrders(ctx, userID, filters)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {

	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orderIDStr := c.Param("id")

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	resp, err := h.service.GetOrderByID(ctx, userID, orderID)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {

	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	err = h.service.CancelOrder(ctx, userID, orderID)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	ctx := c.Request.Context()

	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	err = h.service.UpdateOrderStatus(ctx, orderID, req.Status)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {

	ctx := c.Request.Context()

	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	filters := types.AdminOrderFilters{}
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	resp, err := h.service.GetAllOrders(ctx, filters)
	if err != nil {
		HandleDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
