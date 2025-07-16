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
