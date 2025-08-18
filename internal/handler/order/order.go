package order

import (
	"context"
	"goshop/internal/domain/errors"
	"goshop/internal/domain/types"
	"goshop/internal/dto"
	"goshop/internal/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID int64, req *dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetUserOrders(ctx context.Context, userID int64, filters types.OrderFilters) (*dto.OrdersListResponse, error)
	GetOrderByID(ctx context.Context, userID int64, orderID int64) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, userID int64, orderID int64) error
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	GetAllOrders(ctx context.Context, filters types.AdminOrderFilters) (*dto.OrdersListResponse, error)
}

type OrderHandler struct {
	OrderService OrderService
}

func NewOrderHandler(s OrderService) *OrderHandler {
	return &OrderHandler{OrderService: s}
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

	resp, err := h.OrderService.CreateOrder(ctx, userID, req)
	if err != nil {
		errors.HandleError(c, err)
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

	resp, err := h.OrderService.GetUserOrders(ctx, userID, filters)
	if err != nil {
		errors.HandleError(c, err)
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

	resp, err := h.OrderService.GetOrderByID(ctx, userID, orderID)
	if err != nil {
		errors.HandleError(c, err)
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

	err = h.OrderService.CancelOrder(ctx, userID, orderID)
	if err != nil {
		errors.HandleError(c, err)
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

	err = h.OrderService.UpdateOrderStatus(ctx, orderID, req.Status)
	if err != nil {
		errors.HandleError(c, err)
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

	resp, err := h.OrderService.GetAllOrders(ctx, filters)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
