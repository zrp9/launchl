// Package jsonb provides a generic type to parse postgres/sqldb jsonb types as structs
package jsonb

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONB[T any] struct{ V T }

func (j JSONB[T]) Value() (driver.Value, error) { return json.Marshal(j.V) }

func (j *JSONB[T]) Scan(src any) error {
	if src == nil {
		var zero T
		j.V = zero
		return nil
	}

	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, &j)
	case string:
		return json.Unmarshal([]byte(v), &j.V)
	default:
		return fmt.Errorf("unsupported type %T", src)
	}
}

type NullJSONB[T any] struct {
	V     T
	Valid bool // Valid is true if V is non-NULL
}

func (n NullJSONB[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return json.Marshal(n.V)
}

func (n *NullJSONB[T]) Scan(src any) error {
	if src == nil {
		n.Valid = false
		var zero T
		n.V = zero
		return nil
	}

	n.Valid = true
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, &n.V)
	case string:
		return json.Unmarshal([]byte(v), &n.V)
	default:
		return fmt.Errorf("unsupported type %T", src)
	}
}
