package query

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/adipras/torm/db"
	"github.com/adipras/torm/executor"
	"github.com/adipras/torm/model"
	"github.com/adipras/torm/utils"
)

type Builder struct {
	db        *db.DB
	modelRef  any
	schema    *model.Schema
	whereStmt []string
	args      []any
}

// NewBuilder creates a new query builder for the given model.
func NewBuilder(d *db.DB, modelStruct any) *Builder {
	schema := model.Parse(modelStruct)
	return &Builder{
		db:       d,
		modelRef: modelStruct,
		schema:   schema,
	}
}

// Where adds a WHERE clause to the query.
func (b *Builder) Where(condition string, args ...any) *Builder {
	b.whereStmt = append(b.whereStmt, condition)
	b.args = append(b.args, args...)
	return b
}

// Find executes SELECT * FROM table WHERE ... and fills result.
func (b *Builder) Find(dest any) error {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(b.schema.TableName)

	if len(b.whereStmt) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.whereStmt, " AND "))
	}

	query := sb.String()
	rows, err := b.db.SQL.Query(query, b.args...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	return utils.ScanRows(rows, dest)
}

func (b *Builder) Create(value any) error {
	return executor.Create(b.db.SQL, b.modelRef, value)
}

// First executes SELECT * FROM table WHERE ... LIMIT 1 and fills single struct.
func (b *Builder) First(dest any) error {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	sb.WriteString(b.schema.TableName)

	if len(b.whereStmt) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(b.whereStmt, " AND "))
	}
	sb.WriteString(" LIMIT 1")

	query := sb.String()
	row := b.db.SQL.QueryRow(query, b.args...)

	schema := b.schema
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be pointer to struct")
	}
	elem := val.Elem()

	// Pemetaan kolom ke field
	colToField := map[string]reflect.Value{}
	for _, field := range schema.Fields {
		fieldVal := elem.FieldByName(field.Name)
		if fieldVal.IsValid() && fieldVal.CanAddr() {
			colToField[field.Column()] = fieldVal
		}
	}

	// Siapkan target scan
	columns := make([]string, len(schema.Fields))
	values := make([]any, len(columns))
	for i, f := range schema.Fields {
		columns[i] = f.Column()
		if fv, ok := colToField[f.Column()]; ok {
			values[i] = fv.Addr().Interface()
		} else {
			var dummy any
			values[i] = &dummy
		}
	}

	if err := row.Scan(values...); err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	return nil
}
