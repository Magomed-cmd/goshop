package message

import (
	"context"
	"goshop/internal/domain/errors"
	"goshop/internal/dto"
	"goshop/internal/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MessageService interface {
	CreateMessage(ctx context.Context, userID int64, content string) (*dto.MessageResponse, error)
	DeleteMessage(ctx context.Context, messageID int64, userID int64) error
	UpdateMessage(ctx context.Context, messageID int64, userID int64, newContent string) error
	GetMessagesAfterID(ctx context.Context, afterID int64, limit int) ([]*dto.MessageResponse, error)
	GetRecentMessages(ctx context.Context, limit int) ([]*dto.MessageResponse, error)
}

type MessageHandler struct {
	messageService MessageService
	logger         *zap.Logger
}

func NewMessageHandler(messageService MessageService, logger *zap.Logger) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		logger:         logger,
	}
}

func (m *MessageHandler) CreateMessage(c *gin.Context) {

	req := &dto.CreateMessageRequest{}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	ctx := c.Request.Context()
	resp, err := m.messageService.CreateMessage(ctx, userID, req.Content)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(200, resp)
}

func (m *MessageHandler) DeleteMessage(c *gin.Context) {

	messageIDParam := c.Param("id")

	if messageIDParam == "" {
		c.JSON(400, gin.H{"error": "Invalid message ID"})
		return
	}

	messageID, err := strconv.Atoi(messageIDParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid message ID"})
		return
	}

	ctx := c.Request.Context()
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	err = m.messageService.DeleteMessage(ctx, int64(messageID), userID)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "Message deleted successfully",
	})
}

func (m *MessageHandler) UpdateMessage(c *gin.Context) {

	messageIDParam := c.Param("id")

	if messageIDParam == "" {
		c.JSON(400, gin.H{"error": "Invalid message ID"})
		return
	}

	messageID, err := strconv.Atoi(messageIDParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid message ID"})
		return
	}

	req := &dto.UpdateMessageRequest{}
	if err = c.ShouldBindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := c.Request.Context()
	err = m.messageService.UpdateMessage(ctx, int64(messageID), userID, req.Content)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"message": "Message updated successfully",
	})
}

func (m *MessageHandler) GetMessagesAfterID(c *gin.Context) {

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	messageIDParam := c.Param("id")
	if messageIDParam == "" {
		c.JSON(400, gin.H{"error": "Invalid message ID"})
		return
	}
	messageID, err := strconv.Atoi(messageIDParam)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid message ID"})
		return
	}

	ctx := c.Request.Context()

	resp, err := m.messageService.GetMessagesAfterID(ctx, int64(messageID), limit)
	if err != nil {
		errors.HandleError(c, err)
		return
	}
	c.JSON(200, resp)
	return

}

func (m *MessageHandler) GetRecentMessages(c *gin.Context) {

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	messages, err := m.messageService.GetRecentMessages(c.Request.Context(), limit)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}
