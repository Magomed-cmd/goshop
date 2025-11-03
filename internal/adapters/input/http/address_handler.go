package httpadapter

import (
	"strconv"

	"github.com/gin-gonic/gin"

	httpErrors "goshop/internal/adapters/input/http/errors"
	serviceports "goshop/internal/core/ports/services"
	"goshop/internal/dto"
	"goshop/internal/middleware"
)

type AddressHandler struct {
	addressService serviceports.AddressService
}

func NewAddressHandler(s serviceports.AddressService) *AddressHandler {
	return &AddressHandler{
		addressService: s,
	}
}

func (h *AddressHandler) CreateAddress(c *gin.Context) {

	ctx := c.Request.Context()
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.addressService.CreateAddress(ctx, userID, &req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(201, resp)
}

func (h *AddressHandler) GetUserAddresses(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	resp, err := h.addressService.GetUserAddresses(ctx, userID)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, resp)
}

func (h *AddressHandler) GetAddressByID(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	addressID, err := strconv.Atoi(c.Param("addressID"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid address ID"})
		return
	}

	resp, err := h.addressService.GetAddressByIDForUser(c.Request.Context(), userID, int64(addressID))
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, resp)
}

func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	addressID, err := strconv.Atoi(c.Param("addressID"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid address ID"})
		return
	}

	resp, err := h.addressService.UpdateAddress(c.Request.Context(), userID, int64(addressID), &req)
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.JSON(200, resp)
}

func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	addressID, err := strconv.Atoi(c.Param("addressID"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid address ID"})
		return
	}

	err = h.addressService.DeleteAddress(c.Request.Context(), userID, int64(addressID))
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.Status(204)
}
