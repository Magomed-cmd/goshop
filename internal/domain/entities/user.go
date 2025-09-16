package entities

import (
	"time"

	"github.com/google/uuid"
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

type Message struct {
	ID        int64     `db:"id" json:"id"`
	UUID      uuid.UUID `db:"uuid" json:"uuid"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
