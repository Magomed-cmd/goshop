package ws

import (
	"goshop/internal/service/message"
	"net/http"

	"go.uber.org/zap"
	"goshop/internal/middleware"
	"github.com/gin-gonic/gin"
	"goshop/internal/dto"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin позволяет принимать соединения от любого origin
	// В продакшене следует ограничить разрешенные origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatWSHandler struct {
	hub *Hub
	messageService message.MessageService
	logger         *zap.Logger
}

type incomingMessage struct {
	RecipientID int64  `json:"recipient_id"`
	Content     string `json:"content"`
}


func NewChatWSHandler(hub *Hub, messageService message.MessageService, logger *zap.Logger) *ChatWSHandler {
	return &ChatWSHandler{hub: hub, messageService: messageService, logger: logger}
}


func (h *ChatWSHandler) UpgradeConnection(c *gin.Context) {
	ctx := c.Request.Context()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		h.logger.Error("failed to get user ID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("failed to upgrade connection", zap.Error(err))
		return
	}

	h.hub.Add(userID, conn)
	defer func() {
		h.hub.Remove(userID)
		conn.Close()
	}()

	h.logger.Info("user connected to websocket", zap.Int64("userID", userID))

	for {
		var incoming dto.CreateMessageRequest
		if err := conn.ReadJSON(&incoming); err != nil {
			h.logger.Warn("read failed or connection closed", zap.Error(err))
			break
		}

		msg, err := h.messageService.CreateMessage(ctx, userID, incoming.RecipientID, incoming.Content)
		if err != nil {
			h.logger.Error("failed to create message", zap.Error(err))
			continue
		}

		// отправляем подтверждение себе
		_ = conn.WriteJSON(msg)

		// пытаемся отправить получателю
		h.hub.SendToUser(incoming.RecipientID, msg)
	}
}

