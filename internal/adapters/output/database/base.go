package database

import (
	"goshop/internal/core/ports/repositories"
)

type BaseRepository struct {
	conn repositories.DBConn
}

func NewBaseRepository(conn repositories.DBConn) BaseRepository {
	return BaseRepository{conn: conn}
}

func (b BaseRepository) WithConn(conn repositories.DBConn) BaseRepository {
	b.conn = conn
	return b
}

func (b BaseRepository) Conn() repositories.DBConn {
	return b.conn
}
