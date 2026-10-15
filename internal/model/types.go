package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONBArray JSONB 数组类型（Postgres JSONB / SQLite TEXT 均走 Valuer）
type JSONBArray []string

func (j *JSONBArray) Scan(value any) error {
	if value == nil {
		*j = JSONBArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		if s, ok := value.(string); ok {
			bytes = []byte(s)
		} else {
			return fmt.Errorf("failed to scan JSONBArray: not a byte slice")
		}
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBArray) Value() (driver.Value, error) {
	if j == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(j)
}

// JSONBMap JSONB 对象类型
type JSONBMap map[string]any

func (j *JSONBMap) Scan(value any) error {
	if value == nil {
		*j = JSONBMap{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		if s, ok := value.(string); ok {
			bytes = []byte(s)
		} else {
			return fmt.Errorf("failed to scan JSONBMap: not a byte slice")
		}
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j)
}
