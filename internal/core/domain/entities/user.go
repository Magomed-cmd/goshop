package entities

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainerrors "goshop/internal/core/domain/errors"
	"goshop/internal/core/domain/vo"
)

type Role struct {
	ID   int64     `db:"id" json:"id"`
	UUID uuid.UUID `db:"uuid" json:"uuid"`
	Name string    `db:"name" json:"name"`
}

type User struct {
	ID           int64     `db:"id" json:"id"`
	UUID         uuid.UUID `db:"uuid" json:"uuid"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         *string   `db:"name" json:"name"`
	Phone        *string   `db:"phone" json:"phone"`
	RoleID       *int64    `db:"role_id" json:"role_id"`
	Role         *Role     `db:"-" json:"role,omitempty"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type UserAvatar struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	ImageURL  string    `db:"image_url"`
	UUID      string    `db:"uuid"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewUserForRegistration(email vo.Email, passwordHash string, roleID int64, name, phone *string, createdAt time.Time) (*User, error) {
	if passwordHash == "" || roleID < 1 {
		return nil, domainerrors.ErrInvalidInput
	}

	user := &User{
		UUID:         uuid.New(),
		Email:        email.String(),
		PasswordHash: passwordHash,
		RoleID:       &roleID,
		CreatedAt:    createdAt,
	}
	user.Name = normalizeOptional(name)
	user.Phone = normalizeOptional(phone)

	return user, nil
}

func NewOAuthUser(email vo.Email, roleID int64, name *string) (*User, error) {
	if roleID < 1 {
		return nil, domainerrors.ErrInvalidInput
	}

	normalizedName := normalizeOptional(name)

	return &User{
		UUID:         uuid.New(),
		Email:        email.String(),
		PasswordHash: "",
		Name:         normalizedName,
		Phone:        nil,
		RoleID:       &roleID,
	}, nil
}

func (u *User) AttachRole(role *Role) {
	u.Role = role
}

func (u *User) VerifyPassword(password vo.RawPassword) error {
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password.String())); err != nil {
		return domainerrors.ErrInvalidPassword
	}
	return nil
}

func (u *User) ApplyProfilePatch(name, phone *string) error {
	if name == nil && phone == nil {
		return domainerrors.ErrInvalidInput
	}

	if name != nil {
		normalizedName := strings.TrimSpace(*name)
		if normalizedName == "" {
			return domainerrors.ErrInvalidInput
		}
		u.Name = &normalizedName
	}

	if phone != nil {
		normalizedPhone := strings.TrimSpace(*phone)
		if normalizedPhone == "" {
			return domainerrors.ErrInvalidInput
		}
		u.Phone = &normalizedPhone
	}

	return nil
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}
