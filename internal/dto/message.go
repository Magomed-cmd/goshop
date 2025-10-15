package dto

import "time"

type CreateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

type MessageResponse struct {
	ID        int64     `json:"id"`
	UUID      string    `json:"uuid"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type MessagesListResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int               `json:"total"`
}
