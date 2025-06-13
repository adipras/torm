package db

import (
	"database/sql"
	"fmt"
)

type DB struct {
	SQL *sql.DB
}

// New creates a new DB wrapper.
func New(driver, dsn string) (*DB, error) {
	sqlDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &DB{SQL: sqlDB}, nil
}

// Ping verifies the database connection.
func (db *DB) Ping() error {
	return db.SQL.Ping()
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.SQL.Close()
}
