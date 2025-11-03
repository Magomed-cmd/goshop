package dto

type RegisterRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	Name     *string `json:"name"`
	Phone    *string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=2,max=100"`
	Phone *string `json:"phone" binding:"omitempty,min=10,max=20"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

type UserProfile struct {
	UUID  string  `json:"uuid"`
	Email string  `json:"email"`
	Name  *string `json:"name"`
	Phone *string `json:"phone"`
	Role  string  `json:"role"`
}
