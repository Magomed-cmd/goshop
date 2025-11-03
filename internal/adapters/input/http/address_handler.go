package httpadapter

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"goshop/internal/core/domain/entities"
	"goshop/internal/core/mappers"
	"goshop/internal/dto"
	"goshop/internal/middleware"
)

type AddressService interface {
	CreateAddress(ctx context.Context, userID int64, req *dto.CreateAddressRequest) (*entities.UserAddress, error)
	GetUserAddresses(ctx context.Context, userID int64) ([]*entities.UserAddress, error)
	GetAddressByID(ctx context.Context, addressID int64) (*entities.UserAddress, error)
	UpdateAddress(ctx context.Context, userID int64, addressID int64, req *dto.UpdateAddressRequest) (*entities.UserAddress, error)
	GetAddressByIDForUser(ctx context.Context, userID, addressID int64) (*entities.UserAddress, error)
	DeleteAddress(ctx context.Context, userID int64, addressID int64) error
}

type AddressHandler struct {
	AddressService AddressService
}

func NewAddressHandler(s AddressService) *AddressHandler {
	return &AddressHandler{
		AddressService: s,
	}
}

func (h *AddressHandler) CreateAddress(c *gin.Context) {
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

	result, err := h.AddressService.CreateAddress(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response := mappers.ToAddressResponse(result)

	c.JSON(201, response)
}

func (h *AddressHandler) GetUserAddresses(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	addresses, err := h.AddressService.GetUserAddresses(ctx, userID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.AddressResponse, 0, len(addresses))
	for _, userAddress := range addresses {
		resp := mappers.ToAddressResponse(userAddress)
		response = append(response, resp)
	}

	c.JSON(200, response)
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

	address, err := h.AddressService.GetAddressByIDForUser(c.Request.Context(), userID, int64(addressID))
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(403, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	resp := mappers.ToAddressResponse(address)

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

	updatedAddress, err := h.AddressService.UpdateAddress(c.Request.Context(), userID, int64(addressID), &req)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	resp := mappers.ToAddressResponse(updatedAddress)

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

	err = h.AddressService.DeleteAddress(c.Request.Context(), userID, int64(addressID))
	if err != nil {
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(403, gin.H{"error": "Access denied"})
			return
		}
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.Status(204)
}
