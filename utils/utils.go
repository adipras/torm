package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
)

// ToSnakeCase converts CamelCase or PascalCase to snake_case
// Reuse from model package if shared
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return stringLower(result)
}

func stringLower(rs []rune) string {
	for i := range rs {
		if rs[i] >= 'A' && rs[i] <= 'Z' {
			rs[i] += 'a' - 'A'
		}
	}
	return string(rs)
}

// ScanRows maps rows from DB to a slice of structs
func ScanRows(rows *sql.Rows, dest any) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Slice {
		return errors.New("dest must be a pointer to slice")
	}

	sliceVal := destVal.Elem()
	elemType := sliceVal.Type().Elem() // struct type

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	for rows.Next() {
		elemPtr := reflect.New(elemType) // *T
		elem := elemPtr.Elem()           // T

		fieldPtrs := make([]any, len(columns))
		colToField := map[string]reflect.Value{}

		// Map column name -> struct field pointer
		for i := 0; i < elem.NumField(); i++ {
			field := elem.Type().Field(i)
			if !field.IsExported() {
				continue
			}

			col := field.Tag.Get("db")
			if col == "" {
				col = ToSnakeCase(field.Name)
			}

			colToField[col] = elem.Field(i)
		}

		for i, colName := range columns {
			if f, ok := colToField[colName]; ok && f.CanAddr() {
				fieldPtrs[i] = f.Addr().Interface()
			} else {
				var dummy any
				fieldPtrs[i] = &dummy // ignore column
			}
		}

		if err := rows.Scan(fieldPtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		sliceVal.Set(reflect.Append(sliceVal, elem))
	}

	return rows.Err()
}
