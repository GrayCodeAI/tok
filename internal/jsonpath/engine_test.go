package jsonpath

import "testing"

func TestJSONPathEngine(t *testing.T) {
	engine := NewJSONPathEngine()

	json := `{"name": "Alice", "age": 30, "city": "NYC"}`
	result, err := engine.Query(json, "$.name")
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if result.Value != "Alice" {
		t.Errorf("Expected 'Alice', got %v", result.Value)
	}
}

func TestJSONPathExtractFields(t *testing.T) {
	engine := NewJSONPathEngine()

	json := `{"name": "Bob", "age": 25, "city": "LA"}`
	fields, err := engine.ExtractFields(json, []string{"name", "age"})
	if err != nil {
		t.Fatalf("ExtractFields error: %v", err)
	}
	if fields["name"] != "Bob" {
		t.Errorf("Expected 'Bob', got %v", fields["name"])
	}
	if fields["age"].(float64) != 25 {
		t.Errorf("Expected 25, got %v", fields["age"])
	}
}

func TestJSONPathFilterByKeys(t *testing.T) {
	engine := NewJSONPathEngine()

	json := `{"name": "Charlie", "age": 35, "city": "SF"}`
	filtered, err := engine.FilterByKeys(json, []string{"name", "city"})
	if err != nil {
		t.Fatalf("FilterByKeys error: %v", err)
	}
	if filtered == "" {
		t.Error("Expected non-empty filtered output")
	}
}
