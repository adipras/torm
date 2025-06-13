package torm

import (
	"context"
	"database/sql"

	"github.com/adipras/torm/executor"
	"github.com/adipras/torm/query"

	"github.com/adipras/torm/db"
)

var ErrNoRows = sql.ErrNoRows

type Torm struct {
	DB *db.DB
}

// Open opens a database connection using the given driver and DSN.
func Open(driver string, dsn string) (*Torm, error) {
	conn, err := db.New(driver, dsn)
	if err != nil {
		return nil, err
	}
	return &Torm{DB: conn}, nil
}

// Close closes the underlying SQL database connection.
func (t *Torm) Close() error {
	return t.DB.SQL.Close()
}

// Create inserts one or more rows
// into the database based on the provided schema and data.
// It takes a schema reference (struct type) and the data to insert.
// The data can be a single struct or a slice of structs.
func (t *Torm) Create(schema any, data any) error {
	return executor.Create(t.DB.SQL, schema, data)
}

// Find retrieves rows from the database based on the provided schema.
// It takes a schema reference (struct type) and a destination variable
// where the results will be stored.
func (t *Torm) Find(schema any, dest any) error {
	return executor.Find(t.DB.SQL, schema, dest)
}

// First finds the first matching row based on condition and maps it to dest.
// It takes a schema reference, a destination variable, a WHERE clause, and optional arguments.
// If no rows match, it returns sql.ErrNoRows.
// If multiple rows match, it only returns the first one.
func (t *Torm) First(schema any, dest any, whereClause string, args ...any) error {
	return executor.First(t.DB.SQL, schema, dest, whereClause, args...)
}

// Update updates fields in a table based on a WHERE clause.
// It takes a schema reference, a map of data to update, and a WHERE clause with optional arguments.
func (t *Torm) Update(schema any, data map[string]any, whereClause string, args ...any) error {
	return executor.Update(t.DB.SQL, schema, data, whereClause, args...)
}

// Delete removes rows from the database based on the provided schema and WHERE clause.
// It takes a schema reference and a WHERE clause with optional arguments.
func (t *Torm) Delete(schema any, whereClause string, args ...any) error {
	return executor.Delete(t.DB.SQL, schema, whereClause, args...)
}

// RawSQL executes a raw SQL query with default context
func (t *Torm) RawSQL(query string, args ...any) (*sql.Rows, error) {
	return executor.RawSQL(t.DB.SQL, query, args...)
}

// RawSQLContext executes a raw SQL query with the provided context
func (t *Torm) RawSQLContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return executor.RawSQLContext(t.DB.SQL, ctx, query, args...)
}

// Model initializes a query builder for the given model struct.
// It returns a new query.Builder instance that can be used to build and execute queries.
func (t *Torm) Model(model any) *query.Builder {
	return query.NewBuilder(t.DB, model)
}
