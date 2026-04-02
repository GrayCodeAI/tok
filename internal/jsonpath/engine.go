package jsonpath

import (
	"encoding/json"
	"strings"
)

type JSONPathResult struct {
	Value    interface{} `json:"value"`
	Original string      `json:"original"`
	Filtered int         `json:"filtered"`
}

type JSONPathEngine struct{}

func NewJSONPathEngine() *JSONPathEngine {
	return &JSONPathEngine{}
}

func (e *JSONPathEngine) Query(jsonStr, path string) (*JSONPathResult, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	result := e.evaluate(data, path, 0)
	return &JSONPathResult{
		Value:    result,
		Original: path,
	}, nil
}

func (e *JSONPathEngine) evaluate(data interface{}, path string, depth int) interface{} {
	if depth > 20 {
		return nil
	}

	path = strings.TrimPrefix(path, "$")

	if data == nil || path == "" {
		return data
	}

	if strings.HasPrefix(path, ".") {
		path = strings.TrimPrefix(path, ".")
	}

	field, rest := e.splitPath(path)
	if field == "" {
		return data
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if field == "*" {
			var results []interface{}
			for _, val := range v {
				if rest != "" {
					results = append(results, e.evaluate(val, rest, depth+1))
				} else {
					results = append(results, val)
				}
			}
			return results
		}
		if val, ok := v[field]; ok {
			if rest != "" {
				return e.evaluate(val, rest, depth+1)
			}
			return val
		}
		return nil

	case []interface{}:
		if field == "*" || field == "[" {
			var results []interface{}
			for _, item := range v {
				trimmed := strings.TrimPrefix(rest, "]")
				if trimmed != "" {
					results = append(results, e.evaluate(item, trimmed, depth+1))
				} else {
					results = append(results, item)
				}
			}
			return results
		}

		if strings.HasPrefix(field, "[") && strings.HasSuffix(field, "]") {
			idx := strings.Trim(field, "[]")
			if idx == "*" {
				var results []interface{}
				for _, item := range v {
					results = append(results, e.evaluate(item, rest, depth+1))
				}
				return results
			}

			if idx == "?" {
				filter := strings.TrimPrefix(rest, "(")
				filter = strings.TrimSuffix(filter, ")")
				var results []interface{}
				for _, item := range v {
					if e.matchesFilter(item, filter) {
						results = append(results, item)
					}
				}
				return results
			}

			if strings.Contains(idx, ":") {
				parts := strings.SplitN(idx, ":", 2)
				start := 0
				end := len(v)
				if parts[0] != "" {
					start = e.parseIndex(parts[0])
				}
				if parts[1] != "" {
					end = e.parseIndex(parts[1])
				}
				if start < 0 {
					start = len(v) + start
				}
				if end < 0 {
					end = len(v) + end
				}
				if start >= len(v) {
					start = len(v)
				}
				if end > len(v) {
					end = len(v)
				}
				if start < 0 {
					start = 0
				}
				return v[start:end]
			}

			i := e.parseIndex(idx)
			if i < 0 {
				i = len(v) + i
			}
			if i >= 0 && i < len(v) {
				return e.evaluate(v[i], rest, depth+1)
			}
		}

		var results []interface{}
		for _, item := range v {
			if val := e.evaluate(item, "."+field+"."+rest, depth+1); val != nil {
				results = append(results, val)
			}
		}
		if len(results) > 0 {
			return results
		}
	}

	return nil
}

func (e *JSONPathEngine) splitPath(path string) (field, rest string) {
	if path == "" {
		return "", ""
	}

	if path[0] == '[' {
		end := strings.Index(path, "]")
		if end >= 0 {
			return path[:end+1], path[end+1:]
		}
		return path, ""
	}

	parts := strings.SplitN(path, ".", 2)
	field = parts[0]
	if len(parts) > 1 {
		rest = parts[1]
	}
	return field, rest
}

func (e *JSONPathEngine) parseIndex(s string) int {
	n := 0
	negative := false
	for i, c := range s {
		if i == 0 && c == '-' {
			negative = true
			continue
		}
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	if negative {
		return -n
	}
	return n
}

func (e *JSONPathEngine) matchesFilter(item interface{}, filter string) bool {
	switch v := item.(type) {
	case map[string]interface{}:
		for k, val := range v {
			if strings.Contains(filter, k) && strings.Contains(filter, "==") {
				expected := strings.TrimSpace(strings.Split(filter, "==")[1])
				if valStr, ok := val.(string); ok && valStr == expected {
					return true
				}
			}
		}
	case string:
		if strings.Contains(filter, v) {
			return true
		}
	}
	return false
}

func (e *JSONPathEngine) ExtractFields(jsonStr string, fields []string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, field := range fields {
		if val, ok := data[field]; ok {
			result[field] = val
		}
	}
	return result, nil
}

func (e *JSONPathEngine) FilterByKeys(jsonStr string, keepKeys []string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	filtered := make(map[string]interface{})
	for _, key := range keepKeys {
		if val, ok := data[key]; ok {
			filtered[key] = val
		}
	}

	result, err := json.MarshalIndent(filtered, "", "  ")
	return string(result), err
}
