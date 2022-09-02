package db

import (
	"blog/server/db/sqlc"

	"github.com/jackc/pgx/v4"
)

// Store provides all functions to execute db queries and transactions
type Store interface {
	sqlc.Querier
}

// PgxStore provides all functions to execute db queries and transactions
type PgxStore struct {
	db *pgx.Conn
	*sqlc.Queries
}

func NewStore(db *pgx.Conn) Store {
	return &PgxStore{
		db:      db,
		Queries: sqlc.New(db),
	}
}
