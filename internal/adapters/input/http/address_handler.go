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

// CreateAddress godoc
// @Summary     Create address
// @Description Creates a new address for the authenticated user
// @Tags        addresses
// @Accept      json
// @Produce     json
// @Param       request body dto.CreateAddressRequest true "Address data"
// @Success     201 {object} dto.AddressResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/addresses [post]
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

// GetUserAddresses godoc
// @Summary     List addresses
// @Description Returns all addresses for the authenticated user
// @Tags        addresses
// @Produce     json
// @Success     200 {array} dto.AddressResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/addresses [get]
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

// GetAddressByID godoc
// @Summary     Get address
// @Description Returns address details for the authenticated user
// @Tags        addresses
// @Produce     json
// @Param       id path int true "Address ID"
// @Success     200 {object} dto.AddressResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/addresses/{id} [get]
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

// UpdateAddress godoc
// @Summary     Update address
// @Description Updates an existing address belonging to the authenticated user
// @Tags        addresses
// @Accept      json
// @Produce     json
// @Param       id path int true "Address ID"
// @Param       request body dto.UpdateAddressRequest true "Address data"
// @Success     200 {object} dto.AddressResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/addresses/{id} [put]
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

// DeleteAddress godoc
// @Summary     Delete address
// @Description Deletes an address belonging to the authenticated user
// @Tags        addresses
// @Produce     json
// @Param       id path int true "Address ID"
// @Success     204 {string} string "No Content"
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /api/v1/addresses/{id} [delete]
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
