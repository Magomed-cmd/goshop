package httpadapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	httpErrors "goshop/internal/adapters/input/http/errors"
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

// CreateOrder godoc
// @Summary     Create order
// @Description Creates a new order for the authenticated user
// @Tags        orders
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateOrderRequest true "Order payload"
// @Success     201 {object} dto.OrderResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {

	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	req := &dto.CreateOrderRequest{}

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request Body"})
		return
	}

	resp, err := h.service.CreateOrder(ctx, userID, req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetUserOrders godoc
// @Summary     List user orders
// @Description Returns paginated list of orders for the authenticated user
// @Tags        orders
// @Produce     json
// @Param       page       query int    false "Page number"
// @Param       limit      query int    false "Page size"
// @Param       status     query string false "Order status filter"
// @Param       date_from  query string false "Filter from date (YYYY-MM-DD)"
// @Param       date_to    query string false "Filter to date (YYYY-MM-DD)"
// @Param       min_amount query number false "Minimum order amount"
// @Param       max_amount query number false "Maximum order amount"
// @Param       sort_by    query string false "Sort field"
// @Param       sort_order query string false "Sort order"
// @Success     200 {object} dto.OrdersListResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/orders [get]
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
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetOrderByID godoc
// @Summary     Get order
// @Description Returns details of an order belonging to the authenticated user
// @Tags        orders
// @Produce     json
// @Param       id path int true "Order ID"
// @Success     200 {object} dto.OrderResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/orders/{id} [get]
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
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CancelOrder godoc
// @Summary     Cancel order
// @Description Cancels an order belonging to the authenticated user
// @Tags        orders
// @Produce     json
// @Param       id path int true "Order ID"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/orders/{id}/cancel [post]
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
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}

// UpdateOrderStatus godoc
// @Summary     Update order status
// @Description Updates the status of an order (admin only)
// @Tags        admin/orders
// @Accept      json
// @Produce     json
// @Param       id      path int                         true  "Order ID"
// @Param       request body dto.UpdateOrderStatusRequest true "Status payload"
// @Success     200 {object} map[string]string
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/orders/{id}/status [put]
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
	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	err = h.service.UpdateOrderStatus(ctx, orderID, req.Status)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

// GetAllOrders godoc
// @Summary     List orders (admin)
// @Description Returns paginated list of all orders for administrators
// @Tags        admin/orders
// @Produce     json
// @Param       page       query int    false "Page number"
// @Param       limit      query int    false "Page size"
// @Param       user_id    query int    false "Filter by user ID"
// @Param       status     query string false "Order status filter"
// @Param       date_from  query string false "Filter from date (YYYY-MM-DD)"
// @Param       date_to    query string false "Filter to date (YYYY-MM-DD)"
// @Param       min_amount query number false "Minimum order amount"
// @Param       max_amount query number false "Maximum order amount"
// @Param       sort_by    query string false "Sort field"
// @Param       sort_order query string false "Sort order"
// @Success     200 {object} dto.OrdersListResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /admin/orders [get]
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
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
