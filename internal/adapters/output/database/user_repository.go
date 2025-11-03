package database

import (
    "context"
    "errors"
    "fmt"

    "github.com/Masterminds/squirrel"
    "github.com/jackc/pgx/v5"
    "go.uber.org/zap"

    "goshop/internal/core/domain/entities"
    errors2 "goshop/internal/core/domain/errors"
    portrepo "goshop/internal/core/ports/repositories"
)

const DefaultAvatarURL = "avatars/default-avatar.png"

type UserRepository struct {
	base   BaseRepository
	psql   squirrel.StatementBuilderType
	logger *zap.Logger
}

func NewUserRepository(conn portrepo.DBConn, logger *zap.Logger) *UserRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &UserRepository{
		base:   NewBaseRepository(conn),
		psql:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger: logger,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entities.User) error {
	r.logger.Debug("Creating user in database",
		zap.String("email", user.Email),
		zap.String("name", *user.Name),
		zap.Int64("role_id", *user.RoleID))

	query := `
       INSERT INTO users (uuid, email, password_hash, name, phone, role_id, created_at) 
       VALUES ($1, $2, $3, $4, $5, $6, $7) 
       RETURNING id`

	err := r.base.Conn().QueryRow(ctx, query,
		user.UUID, user.Email, user.PasswordHash,
		user.Name, user.Phone, user.RoleID, user.CreatedAt,
	).Scan(&user.ID)

	if err != nil {
		r.logger.Error("Failed to create user in database",
			zap.Error(err),
			zap.String("email", user.Email),
			zap.String("name", *user.Name),
			zap.Int64("role_id", *user.RoleID))
		return err
	}

	r.logger.Info("User created successfully in database",
		zap.Int64("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("name", *user.Name))

	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.logger.Debug("Getting user by email from database", zap.String("email", email))

	query := "SELECT id, uuid, email, password_hash, name, phone, role_id, created_at FROM users WHERE email = $1"

	var userStruct entities.User
	err := r.base.Conn().QueryRow(ctx, query, email).Scan(
		&userStruct.ID, &userStruct.UUID, &userStruct.Email, &userStruct.PasswordHash,
		&userStruct.Name, &userStruct.Phone, &userStruct.RoleID, &userStruct.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("User not found in database", zap.String("email", email))
			return nil, errors2.ErrUserNotFound
		}
		r.logger.Error("Failed to get user by email from database", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	r.logger.Debug("User retrieved successfully from database",
		zap.Int64("user_id", userStruct.ID),
		zap.String("email", email))

	return &userStruct, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*entities.User, error) {
	r.logger.Debug("Getting user by ID from database", zap.Int64("user_id", id))

	query := "SELECT id, uuid, email, password_hash, name, phone, role_id, created_at FROM users WHERE id = $1"

	var userStruct entities.User
	err := r.base.Conn().QueryRow(ctx, query, id).Scan(
		&userStruct.ID, &userStruct.UUID, &userStruct.Email, &userStruct.PasswordHash,
		&userStruct.Name, &userStruct.Phone, &userStruct.RoleID, &userStruct.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("User not found in database", zap.Int64("user_id", id))
			return nil, errors2.ErrUserNotFound
		}
		r.logger.Error("Failed to get user by ID from database", zap.Error(err), zap.Int64("user_id", id))
		return nil, err
	}

	r.logger.Debug("User retrieved successfully from database",
		zap.Int64("user_id", id),
		zap.String("email", userStruct.Email))

	return &userStruct, nil
}

func (r *UserRepository) UpdateUserProfile(ctx context.Context, userID int64, name *string, phone *string) error {
	r.logger.Debug("Updating user profile in database",
		zap.Int64("user_id", userID),
		zap.Any("name", name),
		zap.Any("phone", phone))

	query := r.psql.Update("users")
	paramsCnt := 0

	if name != nil {
		r.logger.Debug("Setting new name", zap.Int64("user_id", userID), zap.String("new_name", *name))
		query = query.Set("name", *name)
		paramsCnt++
	}

	if phone != nil {
		r.logger.Debug("Setting new phone", zap.Int64("user_id", userID), zap.String("new_phone", *phone))
		query = query.Set("phone", *phone)
		paramsCnt++
	}

	if paramsCnt == 0 {
		r.logger.Warn("No fields to update", zap.Int64("user_id", userID))
		return fmt.Errorf("no fields to update")
	}

	query = query.Where(squirrel.Eq{"id": userID})

	sql, args, err := query.ToSql()
	if err != nil {
		r.logger.Error("Failed to build update query", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	r.logger.Debug("Executing update query", zap.Int64("user_id", userID), zap.String("query", sql))

	result, err := r.base.Conn().Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("Failed to execute update query", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	if result.RowsAffected() == 0 {
		r.logger.Warn("User not found for update", zap.Int64("user_id", userID))
		return fmt.Errorf("user with ID %d not found", userID)
	}

	r.logger.Info("User profile updated successfully",
		zap.Int64("user_id", userID),
		zap.Int("params_updated", paramsCnt))

	return nil
}

func (r *UserRepository) SaveAvatar(ctx context.Context, userAvatarInfo *entities.UserAvatar) (int, error) {

	query := `INSERT INTO user_avatars (user_id, image_url, created_at, updated_at, uuid) 
				VALUES ($1, $2, $3, $4, $5) 
				ON CONFLICT (user_id) 
				DO UPDATE 
				SET image_url = EXCLUDED.image_url, 
				    updated_at = EXCLUDED.updated_at, 
				    uuid = EXCLUDED.uuid 
				RETURNING id`

	var id int

	if err := r.base.Conn().QueryRow(
		ctx,
		query,
		userAvatarInfo.UserID,
		userAvatarInfo.ImageURL,
		userAvatarInfo.CreatedAt,
		userAvatarInfo.UpdatedAt,
		userAvatarInfo.UUID).
		Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *UserRepository) GetAvatar(ctx context.Context, userID int) (*entities.UserAvatar, error) {

	query := `SELECT id, user_id, image_url, created_at, updated_at, uuid FROM user_avatars WHERE user_id = $1`

	userAvatarInfo := &entities.UserAvatar{}

	if err := r.base.Conn().QueryRow(ctx, query, userID).Scan(
		&userAvatarInfo.ID,
		&userAvatarInfo.UserID,
		&userAvatarInfo.ImageURL,
		&userAvatarInfo.CreatedAt,
		&userAvatarInfo.UpdatedAt,
		&userAvatarInfo.UUID,
	); err != nil {
		return nil, err
	}

	return userAvatarInfo, nil
}

func (r *UserRepository) DeleteAvatar(ctx context.Context, userID int) error {

	query := `DELETE FROM user_avatars WHERE user_id = $1`

	result, err := r.base.Conn().Exec(ctx, query, userID)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors2.ErrAvatarNotFound
	}

	return nil
}
