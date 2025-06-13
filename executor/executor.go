package executor

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/adipras/torm/model"
	"github.com/adipras/torm/utils"
)

// Create inserts a single record into the database
func Create(db *sql.DB, modelRef any, data any) error {
	schema, err := model.ExtractSchema(modelRef)
	if err != nil {
		return err
	}

	vmap, err := model.ExtractValues(data)
	if err != nil {
		return err
	}

	fieldNames := []string{}
	placeholders := []string{}
	values := []any{}

	for _, f := range schema.Fields {
		if val, ok := vmap[f.Name]; ok {
			fieldNames = append(fieldNames, f.Column())
			placeholders = append(placeholders, "?")
			values = append(values, val)
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		schema.Table(),
		strings.Join(fieldNames, ", "),
		strings.Join(placeholders, ", "),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}

	// Optional: set auto-increment ID ke struct
	id, err := res.LastInsertId()
	if err == nil {
		// coba set field ID jika ada
		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			if idField := rv.FieldByName("ID"); idField.IsValid() && idField.CanSet() && idField.Kind() == reflect.Int {
				idField.SetInt(id)
			}
		}
	}

	return nil
}

// Find retrieves all rows for the given schema and maps to dest
func Find(db *sql.DB, schema any, dest any) error {
	// Extract table name
	s := model.Parse(schema)

	query := fmt.Sprintf("SELECT * FROM %s", s.Table())
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	return utils.ScanRows(rows, dest)
}

// First retrieves the first matching row for the given schema and maps to dest.
func First(db *sql.DB, schema any, dest any, whereClause string, args ...any) error {
	s := model.Parse(schema)

	query := fmt.Sprintf("SELECT * FROM %s %s LIMIT 1", s.Table(), whereClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Scan ke slice sementara, ambil index 0
	tmp := reflect.New(reflect.SliceOf(reflect.TypeOf(dest).Elem())).Interface()

	if err := utils.ScanRows(rows, tmp); err != nil {
		return err
	}

	tmpVal := reflect.ValueOf(tmp).Elem()
	if tmpVal.Len() == 0 {
		return sql.ErrNoRows
	}

	reflect.ValueOf(dest).Elem().Set(tmpVal.Index(0))
	return nil
}

// Update updates fields in a table based on a WHERE clause.
func Update(db *sql.DB, schemaRef any, data map[string]any, whereClause string, args ...any) error {
	schema, err := model.ExtractSchema(schemaRef)
	if err != nil {
		return err
	}

	setClauses := []string{}
	values := []any{}

	for key, val := range data {
		colName := key
		// Optional: if key is struct field name, convert to snake_case
		colName = utils.ToSnakeCase(colName)
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", colName))
		values = append(values, val)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s %s",
		schema.Table(),
		strings.Join(setClauses, ", "),
		whereClause,
	)

	values = append(values, args...) // add WHERE args

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, values...)
	return err
}

// Delete removes rows from a table based on a WHERE clause.
func Delete(db *sql.DB, schemaRef any, whereClause string, args ...any) error {
	schema, err := model.ExtractSchema(schemaRef)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s %s", schema.Table(), whereClause)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, query, args...)
	return err
}

// RawSQL runs a raw SQL query with default context (no timeout)
func RawSQL(db *sql.DB, query string, args ...any) (*sql.Rows, error) {
	return db.Query(query, args...)
}

// RawSQLContext runs a raw SQL query with a provided context
func RawSQLContext(db *sql.DB, ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}
