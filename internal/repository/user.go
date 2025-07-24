package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"goshop/internal/domain/entities"
	"goshop/internal/service/user"
)

type UserRepository struct {
	db   *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewUserRepository(conn *pgxpool.Pool) user.UserRepositoryInterface {
	return &UserRepository{
		db:   conn,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *entities.User) error {
	query := `
       INSERT INTO users (uuid, email, password_hash, name, phone, role_id, created_at) 
       VALUES ($1, $2, $3, $4, $5, $6, $7) 
       RETURNING id`

	err := r.db.QueryRow(ctx, query,
		user.UUID, user.Email, user.PasswordHash,
		user.Name, user.Phone, user.RoleID, user.CreatedAt,
	).Scan(&user.ID)

	return err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := "SELECT id, uuid, email, password_hash, name, phone, role_id, created_at FROM users WHERE email = $1"

	var userStruct entities.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&userStruct.ID, &userStruct.UUID, &userStruct.Email, &userStruct.PasswordHash,
		&userStruct.Name, &userStruct.Phone, &userStruct.RoleID, &userStruct.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &userStruct, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*entities.User, error) {
	query := "SELECT id, uuid, email, password_hash, name, phone, role_id, created_at FROM users WHERE id = $1"

	var userStruct entities.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&userStruct.ID, &userStruct.UUID, &userStruct.Email, &userStruct.PasswordHash,
		&userStruct.Name, &userStruct.Phone, &userStruct.RoleID, &userStruct.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}

	return &userStruct, nil
}

func (r *UserRepository) UpdateUserProfile(ctx context.Context, userID int64, name *string, phone *string) error {

	query := r.psql.Update("users")
	paramsCnt := 0

	if name != nil {
		query = query.Set("name", *name)
		paramsCnt++
	}

	if phone != nil {
		query = query.Set("phone", *phone)
		paramsCnt++
	}

	if paramsCnt == 0 {
		return fmt.Errorf("no fields to update")
	}

	query = query.Where(squirrel.Eq{"id": userID})

	sql, args, err := query.ToSql()

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with ID %d not found", userID)
	}

	return nil
}
