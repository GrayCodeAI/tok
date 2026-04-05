package datavalidation

import "testing"

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
}

func TestValidateConsistency_Empty(t *testing.T) {
	v := NewValidator()
	issues := v.ValidateConsistency(map[string]interface{}{})
	if len(issues) != 0 {
		t.Errorf("ValidateConsistency(empty) = %d issues, want 0", len(issues))
	}
}

func TestValidateConsistency_ValidData(t *testing.T) {
	v := NewValidator()
	data := map[string]interface{}{
		"tokens_saved":    float64(1000),
		"original_tokens": float64(5000),
		"filtered_tokens": float64(4000),
	}
	issues := v.ValidateConsistency(data)
	_ = issues // Should not panic
}

func TestValidateConsistency_NonFloat64(t *testing.T) {
	v := NewValidator()
	data := map[string]interface{}{
		"tokens_saved":    "1000",
		"original_tokens": int64(5000),
		"filtered_tokens": 4000,
	}
	// Should not panic even with wrong types
	issues := v.ValidateConsistency(data)
	_ = issues
}

func TestValidateConsistency_NilData(t *testing.T) {
	v := NewValidator()
	issues := v.ValidateConsistency(nil)
	if len(issues) != 0 {
		t.Errorf("ValidateConsistency(nil) = %d issues, want 0", len(issues))
	}
}
