package ws

import (
	"context"
	"goshop/internal/dto"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type MessageService interface {
	CreateMessage(ctx context.Context, userID int64, content string) (*dto.MessageResponse, error)
	DeleteMessage(ctx context.Context, messageID int64, userID int64) error
	UpdateMessage(ctx context.Context, messageID int64, userID int64, newContent string) error
	GetMessagesAfterID(ctx context.Context, afterID int64, limit int) ([]*dto.MessageResponse, error)
	GetRecentMessages(ctx context.Context, limit int) ([]*dto.MessageResponse, error)
}

type Hub struct {
	connections    sync.Map // key: userID (string или int), value: *websocket.Conn
	messageService MessageService
	logger         *zap.Logger
}

func (h *Hub) Add(userID int, conn *websocket.Conn) {
	h.connections.Store(userID, conn)
}

func (h *Hub) Remove(userID int) {
	val, ok := h.connections.Load(userID)
	if ok {
		conn := val.(*websocket.Conn)
		conn.Close()
		h.connections.Delete(userID)
	}
}

func (h *Hub) SendToUser(userID int, message *dto.MessageResponse) {
	user, ok := h.connections.Load(userID)
	if !ok {
		h.logger.Info("user not found")

		return
	}

	userConn := user.(*websocket.Conn)

	err := userConn.WriteJSON(message)
	if err != nil {
		h.logger.Error("failed to send message to user", zap.Error(err))
		return
	}

	return
}
