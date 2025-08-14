package address

import (
	"context"
	"github.com/gin-gonic/gin"
	"goshop/internal/domain/entities"
	"goshop/internal/dto"
	"goshop/internal/middleware"
	"strconv"
	"strings"
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

	response := dto.AddressResponse{
		ID:         result.ID,
		UUID:       result.UUID.String(),
		Address:    result.Address,
		City:       result.City,
		PostalCode: result.PostalCode,
		Country:    result.Country,
		CreatedAt:  result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

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

	var response []dto.AddressResponse
	for _, userAddress := range addresses {
		response = append(response, dto.AddressResponse{
			ID:         userAddress.ID,
			UUID:       userAddress.UUID.String(),
			Address:    userAddress.Address,
			City:       userAddress.City,
			PostalCode: userAddress.PostalCode,
			Country:    userAddress.Country,
			CreatedAt:  userAddress.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
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

	response := dto.AddressResponse{
		ID:         address.ID,
		UUID:       address.UUID.String(),
		Address:    address.Address,
		City:       address.City,
		PostalCode: address.PostalCode,
		Country:    address.Country,
		CreatedAt:  address.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(200, response)
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

	response := dto.AddressResponse{
		ID:         updatedAddress.ID,
		UUID:       updatedAddress.UUID.String(),
		Address:    updatedAddress.Address,
		City:       updatedAddress.City,
		PostalCode: updatedAddress.PostalCode,
		Country:    updatedAddress.Country,
		CreatedAt:  updatedAddress.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(200, response)
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
