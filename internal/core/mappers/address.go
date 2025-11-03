package mappers

import (
    "goshop/internal/core/domain/entities"
    "goshop/internal/dto"
)

// ToAddressResponse converts a user address entity into a DTO.
func ToAddressResponse(address *entities.UserAddress) dto.AddressResponse {
    if address == nil {
        return dto.AddressResponse{}
    }

    return dto.AddressResponse{
        ID:         address.ID,
        UUID:       address.UUID.String(),
        Address:    address.Address,
        City:       address.City,
        PostalCode: address.PostalCode,
        Country:    address.Country,
        CreatedAt:  formatTime(address.CreatedAt),
    }
}
