package model

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/adipras/torm/utils"
)

type Field struct {
	Name   string // struct field name (e.g. "UserName")
	DBName string // db column name (e.g. "user_name")
}

func (f Field) Column() string {
	return f.DBName
}

type Schema struct {
	TableName string
	Fields    []Field
}

func (s *Schema) Table() string {
	return s.TableName
}

var schemaCache = sync.Map{}

// Parse parses a struct into a Schema definition (with caching).
func Parse(model any) *Schema {
	rt := reflect.TypeOf(model)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if cached, ok := schemaCache.Load(rt); ok {
		return cached.(*Schema)
	}

	schema := &Schema{
		TableName: parseTableName(rt.Name()),
	}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		dbTag := field.Tag.Get("db")
		if dbTag == "-" {
			continue
		}

		column := dbTag
		if column == "" {
			column = utils.ToSnakeCase(field.Name)
		}

		schema.Fields = append(schema.Fields, Field{
			Name:   field.Name,
			DBName: column,
		})
	}

	schemaCache.Store(rt, schema)
	return schema
}

func ExtractSchema(model any) (*Schema, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}
	return Parse(model), nil
}

// ExtractValues extracts struct field values as map[fieldName]any
func ExtractValues(model any) (map[string]any, error) {
	result := map[string]any{}

	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be struct or pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		val := v.Field(i).Interface()
		result[field.Name] = val
	}

	return result, nil
}

// Default table name = pluralized snake_case struct name (can be overridden later)
func parseTableName(name string) string {
	return utils.ToSnakeCase(name) + "s" // e.g. User -> users
}
