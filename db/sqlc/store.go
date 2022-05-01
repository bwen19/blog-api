package db

import "database/sql"

// Store provides all functions to execute db queries and transactions
type Store interface {
	Querier
}

// SQLStore provides all functions to execute db queries and transactions
type SQLStore struct {
	db *sql.DB
	*Queries
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}
