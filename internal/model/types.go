package model

import (
	"encoding/json"
	"fmt"
)

// JSONBArray JSONB 数组类型
type JSONBArray []string

func (j *JSONBArray) Scan(value any) error {
	if value == nil {
		*j = JSONBArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONBArray: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBArray) Value() (any, error) {
	if j == nil {
		return "[]", nil
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
		return fmt.Errorf("failed to scan JSONBMap: not a byte slice")
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONBMap) Value() (any, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}
