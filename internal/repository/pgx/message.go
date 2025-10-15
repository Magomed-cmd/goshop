package pgx

import (
	"context"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const DefaultMessagesCapacity = 100

type MessageRepository struct {
	db     *pgxpool.Pool
	psql   squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewMessageRepository(db *pgxpool.Pool, logger *zap.Logger) *MessageRepository {
	return &MessageRepository{
		db:     db,
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger: logger,
	}
}

func (r *MessageRepository) CreateMessage(ctx context.Context, message *entities.Message) (int64, error) {

	query := `INSERT INTO messages (uuid, content, user_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.db.QueryRow(
		ctx,
		query,
		message.UUID,
		message.Content,
		message.UserID,
		message.CreatedAt,
	).Scan(&message.ID)
	if err != nil {
		r.logger.Error("Error to resp of query")
		return 0, err
	}

	return message.ID, nil
}

func (r *MessageRepository) GetMessageByID(ctx context.Context, id int64) (*entities.Message, error) {

	query := `SELECT * FROM messages WHERE id = $1`
	message := &entities.Message{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&message.ID,
		&message.UUID,
		&message.UserID,
		&message.Content,
		&message.CreatedAt,
	)
	if err != nil {
		r.logger.Error("Error to resp of query")
		return nil, err
	}

	return message, nil
}

func (r *MessageRepository) GetMessages(ctx context.Context) ([]*entities.Message, error) {

	query := `SELECT * FROM messages`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages, err := r.scanMessages(rows)
	if err != nil {
		r.logger.Error("Failed to scan message row", zap.Error(err))
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetUserMessages(ctx context.Context, userID int64) ([]*entities.Message, error) {

	query := `SELECT * FROM messages WHERE user_id = $1`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages, err := r.scanMessages(rows)
	if err != nil {
		r.logger.Error("Failed to scan message row", zap.Error(err))
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) DeleteMessage(ctx context.Context, id int64) error {

	query := `DELETE FROM messages WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete message", zap.Error(err), zap.Int64("message_id", id))
		return err
	}

	if result.RowsAffected() == 0 {
		r.logger.Warn("Message not found for deletion", zap.Int64("message_id", id))
		return errors.ErrMessageNotFound
	}

	return nil
}

func (r *MessageRepository) UpdateMessage(ctx context.Context, message *entities.Message) error {

	query := `UPDATE messages SET content = $1 WHERE id = $2`

	res, err := r.db.Exec(ctx, query, message.Content, message.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.ErrMessageNotFound
	}

	return nil
}

func (r *MessageRepository) GetMessagesAfterID(ctx context.Context, afterID int64, limit int) ([]*entities.Message, error) {

	query := `
			SELECT * FROM messages 
			WHERE id > $1 
			ORDER BY id
			LIMIT $2
			 `

	rows, err := r.db.Query(ctx, query, afterID, limit)
	if err != nil {
		r.logger.Error("Failed to scan message row", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	messages, err := r.scanMessages(rows)
	if err != nil {
		r.logger.Error("Failed to scan message row", zap.Error(err))
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetRecentMessages(ctx context.Context, limit int) ([]*entities.Message, error) {
	query := `
        SELECT * FROM messages 
        ORDER BY id DESC 
        LIMIT $1
    `

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		r.logger.Error("Failed to get recent messages", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	messages, err := r.scanMessages(rows)
	if err != nil {
		r.logger.Error("Failed to scan messages", zap.Error(err))
		return nil, err
	}

	return messages, nil
}

// --------- helpers ---------

func (r *MessageRepository) scanMessage(row pgx.Rows) (*entities.Message, error) {
	message := &entities.Message{}
	return message, row.Scan(
		&message.ID,
		&message.UUID,
		&message.UserID,
		&message.Content,
		&message.CreatedAt,
	)
}

func (r *MessageRepository) scanMessages(rows pgx.Rows) ([]*entities.Message, error) {
	messages := make([]*entities.Message, 0, DefaultMessagesCapacity)
	for rows.Next() {
		message, err := r.scanMessage(rows)
		if err != nil {
			r.logger.Error("Failed to scan message row", zap.Error(err))
			return nil, err
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		r.logger.Error("Row iteration error", zap.Error(err))
		return nil, err
	}

	return messages, rows.Err()
}
