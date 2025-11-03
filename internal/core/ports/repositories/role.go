package repositories

import (
    "context"

    "goshop/internal/core/domain/entities"
)

type RoleRepository interface {
    GetByID(ctx context.Context, id int64) (*entities.Role, error)
    GetByName(ctx context.Context, name string) (*entities.Role, error)
}

