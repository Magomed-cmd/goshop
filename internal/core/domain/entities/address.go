package entities

import (
	"time"

	"github.com/google/uuid"
)

type UserAddress struct {
	ID         int64     `json:"id"`
	UUID       uuid.UUID `json:"uuid"`
	UserID     int64     `json:"user_id"`
	Address    string    `json:"address"`
	City       *string   `json:"city"`
	PostalCode *string   `json:"postal_code"`
	Country    *string   `json:"country"`
	CreatedAt  time.Time `json:"created_at"`
}

func (UserAddress) TableName() string {
	return "user_addresses"
}
