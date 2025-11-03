package mappers

import (
	"goshop/internal/core/domain/entities"
	"goshop/internal/dto"
)

func ToUserProfile(user *entities.User, roleName string) dto.UserProfile {
	if user == nil {
		return dto.UserProfile{}
	}

	return dto.UserProfile{
		UUID:  user.UUID.String(),
		Email: user.Email,
		Name:  user.Name,
		Phone: user.Phone,
		Role:  roleName,
	}
}
