package entities

import "time"

type UserAvatar struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	ImageURL  string    `db:"image_url"`
	UUID      string    `db:"uuid"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
