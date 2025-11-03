package httpadapter

import (
	"strconv"

	"github.com/gin-gonic/gin"

	httpErrors "goshop/internal/adapters/input/http/errors"
	"goshop/internal/core/mappers"
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

	result, err := h.addressService.CreateAddress(c.Request.Context(), userID, &req)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	addresses, err := h.addressService.GetUserAddresses(ctx, userID)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	address, err := h.addressService.GetAddressByIDForUser(c.Request.Context(), userID, int64(addressID))
	if err != nil {
		httpErrors.HandleError(c, err)
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

	updatedAddress, err := h.addressService.UpdateAddress(c.Request.Context(), userID, int64(addressID), &req)
	if err != nil {
		httpErrors.HandleError(c, err)
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

	err = h.addressService.DeleteAddress(c.Request.Context(), userID, int64(addressID))
	if err != nil {
		httpErrors.HandleError(c, err)
		return
	}

	c.Status(204)
}
