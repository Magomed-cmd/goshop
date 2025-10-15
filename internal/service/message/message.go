package message

import (
	"context"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/errors"
	"goshop/internal/dto"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

/*
type MessageService interface {
	CreateMessage(ctx context.Context, userID int64, content string) (*dto.MessageResponse, error)
	DeleteMessage(ctx context.Context, messageID int64, userID int64) error
	UpdateMessage(ctx context.Context, messageID int64, userID int64, newContent string) error
	GetMessagesAfterID(ctx context.Context, afterID int64, limit int) ([]*dto.MessageResponse, error)
	GetRecentMessages(ctx context.Context, limit int) ([]*dto.MessageResponse, error)
}*/

type MessageRepository interface {
	CreateMessage(ctx context.Context, message *entities.Message) (int64, error)
	GetMessageByID(ctx context.Context, id int64) (*entities.Message, error)
	GetMessages(ctx context.Context) ([]*entities.Message, error)
	GetUserMessages(ctx context.Context, userID int64) ([]*entities.Message, error)
	DeleteMessage(ctx context.Context, id int64) error
	UpdateMessage(ctx context.Context, message *entities.Message) error
	GetMessagesAfterID(ctx context.Context, afterID int64, limit int) ([]*entities.Message, error)
	GetRecentMessages(ctx context.Context, limit int) ([]*entities.Message, error)
}

type MessageService struct {
	messageRepo MessageRepository
	logger      *zap.Logger
}

func NewMessageService(messageRepo MessageRepository, logger *zap.Logger) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		logger:      logger,
	}
}

func (m *MessageService) CreateMessage(ctx context.Context, userID int64, content string) (*dto.MessageResponse, error) {

	if userID < 1 {
		m.logger.Error("invalid user ID", zap.Int64("userID", userID))
		return nil, errors.ErrInvalidUserID
	}

	if content == "" {
		m.logger.Error("message content is empty")
		return nil, errors.ErrMessageContentEmpty
	}

	if len(content) > 5000 {
		m.logger.Error("message content is too long", zap.Int("length", len(content)), zap.String("content", content))
		return nil, errors.ErrMessageTooLong
	}

	message := &entities.Message{
		UUID:      uuid.New(),
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	id, err := m.messageRepo.CreateMessage(ctx, message)
	if err != nil {
		return nil, err
	}
	message.ID = id

	return &dto.MessageResponse{
		ID:        message.ID,
		UUID:      message.UUID.String(),
		UserID:    message.UserID,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}, nil
}

func (m *MessageService) DeleteMessage(ctx context.Context, messageID int64, userID int64) error {

	if userID < 1 {
		m.logger.Error("invalid user ID", zap.Int64("userID", userID))
		return errors.ErrInvalidUserID
	}

	if messageID < 1 {
		m.logger.Error("invalid message ID", zap.Int64("messageID", messageID))
		return errors.ErrInvalidMessageID
	}

	message, err := m.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		m.logger.Error("failed to get message by ID", zap.Error(err), zap.Int64("messageID", messageID))
		return err
	}

	if message.UserID != userID {
		m.logger.Error("user does not own message", zap.Int64("userID", userID), zap.Int64("messageID", messageID))
		return errors.ErrMessageNotOwnedByUser
	}

	if err = m.messageRepo.DeleteMessage(ctx, messageID); err != nil {
		m.logger.Error("failed to delete message", zap.Error(err), zap.Int64("messageID", messageID))
		return err
	}

	return nil
}

func (m *MessageService) UpdateMessage(ctx context.Context, messageID int64, userID int64, newContent string) error {

	if messageID < 1 {
		m.logger.Error("invalid message ID", zap.Int64("messageID", messageID))
		return errors.ErrInvalidMessageID
	}
	if userID < 1 {
		m.logger.Error("invalid user ID", zap.Int64("userID", userID))
		return errors.ErrInvalidUserID
	}
	if newContent == "" {
		m.logger.Error("message content is empty")
		return errors.ErrMessageContentEmpty
	}
	if len(newContent) > 5000 {
		m.logger.Error("message content is too long", zap.Int("length", len(newContent)), zap.String("content", newContent))
		return errors.ErrMessageTooLong
	}

	message, err := m.messageRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		m.logger.Error("failed to get message by ID", zap.Error(err), zap.Int64("messageID", messageID))
		return err
	}

	if message.UserID != userID {
		m.logger.Error("user does not own message", zap.Int64("userID", userID), zap.Int64("messageID", messageID))
		return errors.ErrMessageNotOwnedByUser
	}

	messageEntity := &entities.Message{
		ID:        messageID,
		UserID:    message.UserID,
		Content:   newContent,
		CreatedAt: message.CreatedAt,
	}

	err = m.messageRepo.UpdateMessage(ctx, messageEntity)
	if err != nil {
		m.logger.Error("failed to update message", zap.Error(err), zap.Int64("messageID", messageID))
		return err
	}

	return nil
}

func (m *MessageService) GetMessagesAfterID(ctx context.Context, afterID int64, limit int) ([]*dto.MessageResponse, error) {

	if afterID < 0 {
		m.logger.Error("After ID should be 0 or more", zap.Int64("afterID", afterID))
		return nil, errors.ErrInvalidMessageID
	}

	if limit < 1 {
		m.logger.Error("Limit should be 1 or more", zap.Int("limit", limit))
		return nil, errors.ErrInvalidLimit
	}

	messagesEntity, err := m.messageRepo.GetMessagesAfterID(ctx, afterID, limit)
	if err != nil {
		m.logger.Error("failed to get messages after ID", zap.Error(err), zap.Int64("afterID", afterID))
		return nil, err
	}

	messagesDTO := make([]*dto.MessageResponse, 0, len(messagesEntity))
	for _, message := range messagesEntity {
		messagesDTO = append(messagesDTO, &dto.MessageResponse{
			ID:        message.ID,
			UUID:      message.UUID.String(),
			UserID:    message.UserID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		})
		m.logger.Debug("message", zap.Any("message", message))
	}

	return messagesDTO, nil
}

func (m *MessageService) GetRecentMessages(ctx context.Context, limit int) ([]*dto.MessageResponse, error) {
	if limit < 1 {
		m.logger.Error("Limit should be 1 or more", zap.Int("limit", limit))
		return nil, errors.ErrInvalidLimit
	}

	messagesEntity, err := m.messageRepo.GetRecentMessages(ctx, limit)
	if err != nil {
		m.logger.Error("failed to get recent messages", zap.Error(err), zap.Int("limit", limit))
		return nil, err
	}

	messagesDTO := make([]*dto.MessageResponse, 0, len(messagesEntity))
	for _, message := range messagesEntity {
		messagesDTO = append(messagesDTO, &dto.MessageResponse{
			ID:        message.ID,
			UUID:      message.UUID.String(),
			UserID:    message.UserID,
			Content:   message.Content,
			CreatedAt: message.CreatedAt,
		})
	}

	return messagesDTO, nil
}
