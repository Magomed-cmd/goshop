package entities

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        int64     `db:"id" json:"id"`
	UUID      uuid.UUID `db:"uuid" json:"uuid"`
	ProductID int64     `db:"product_id" json:"product_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Rating    int       `db:"rating" json:"rating"`
	Comment   *string   `db:"comment" json:"comment"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	User      *User     `db:"-" json:"user,omitempty"`
	Product   *Product  `db:"-" json:"product,omitempty"`
}
