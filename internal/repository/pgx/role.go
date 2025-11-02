package pgx

import (
	"context"
	"errors"
	"goshop/internal/domain/entities"
	"goshop/internal/domain/repository"

	"github.com/jackc/pgx/v5"
)

type RoleRepository struct {
	base BaseRepository
}

func NewRoleRepository(conn repository.DBConn) *RoleRepository {
	return &RoleRepository{base: NewBaseRepository(conn)}
}

func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*entities.Role, error) {

	query := "SELECT * FROM roles WHERE id = $1"
	var role entities.Role

	err := r.base.Conn().QueryRow(ctx, query, id).Scan(&role.ID, &role.UUID, &role.Name)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &role, nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*entities.Role, error) {
	query := "SELECT * FROM roles WHERE name = $1"
	var role entities.Role
	err := r.base.Conn().QueryRow(ctx, query, name).Scan(&role.ID, &role.UUID, &role.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &role, nil
}
