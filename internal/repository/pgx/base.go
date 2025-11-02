package pgx

import (
	"goshop/internal/domain/repository"
)

type BaseRepository struct {
	conn repository.DBConn
}

func NewBaseRepository(conn repository.DBConn) BaseRepository {
	return BaseRepository{conn: conn}
}

func (b BaseRepository) WithConn(conn repository.DBConn) BaseRepository {
	b.conn = conn
	return b
}

func (b BaseRepository) Conn() repository.DBConn {
	return b.conn
}
