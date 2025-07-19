package models

import (
	"github.com/google/uuid"
	"time"
)

type UserAddress struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	UUID       uuid.UUID `gorm:"type:uuid;unique;not null" json:"uuid"`
	UserID     int64     `gorm:"not null;index" json:"user_id"`
	Address    string    `gorm:"not null" json:"address"`
	City       *string   `json:"city"`
	PostalCode *string   `json:"postal_code"`
	Country    *string   `json:"country"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserAddress) TableName() string {
	return "user_addresses"
}
