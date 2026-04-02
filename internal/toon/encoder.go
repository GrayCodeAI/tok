package toon

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type TOONEncoder struct{}

func NewTOONEncoder() *TOONEncoder {
	return &TOONEncoder{}
}

func (e *TOONEncoder) IsHomogeneous(data interface{}) bool {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice {
		return false
	}
	if rv.Len() < 3 {
		return false
	}

	firstType := rv.Index(0).Type()
	for i := 1; i < rv.Len(); i++ {
		if rv.Index(i).Type() != firstType {
			return false
		}
	}
	return true
}

func (e *TOONEncoder) EncodeJSON(jsonInput string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonInput), &data); err != nil {
		return "", err
	}

	arr, ok := data.([]interface{})
	if !ok || len(arr) < 3 {
		return "", fmt.Errorf("not a homogeneous array")
	}

	firstObj, ok := arr[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("array elements are not objects")
	}

	var columns []string
	for k := range firstObj {
		columns = append(columns, k)
	}

	var sb strings.Builder
	sb.WriteString("TOON\n")
	sb.WriteString(strings.Join(columns, "\t"))
	sb.WriteString("\n")

	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var values []string
		for _, col := range columns {
			val := obj[col]
			values = append(values, formatValue(val))
		}
		sb.WriteString(strings.Join(values, "\t"))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (e *TOONEncoder) Decode(toonInput string) (string, error) {
	lines := strings.Split(strings.TrimSpace(toonInput), "\n")
	if len(lines) < 3 || lines[0] != "TOON" {
		return "", fmt.Errorf("invalid TOON format")
	}

	columns := strings.Split(lines[1], "\t")
	var result []map[string]interface{}

	for _, line := range lines[2:] {
		values := strings.Split(line, "\t")
		obj := make(map[string]interface{})
		for i, col := range columns {
			if i < len(values) {
				obj[col] = parseValue(values[i])
			}
		}
		result = append(result, obj)
	}

	jsonBytes, err := json.Marshal(result)
	return string(jsonBytes), err
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int(val)) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", val)
	}
}

func parseValue(s string) interface{} {
	if s == "null" {
		return nil
	}
	if s == "1" || s == "0" {
		return s
	}
	var n float64
	if _, err := fmt.Sscanf(s, "%g", &n); err == nil {
		return n
	}
	return s
}

func (e *TOONEncoder) CompressionRatio(original string, encoded string) float64 {
	if len(original) == 0 {
		return 0
	}
	return float64(len(encoded)) / float64(len(original)) * 100
}
