package datavalidation

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateConsistency(data map[string]interface{}) []string {
	var errors []string
	if tokens, ok := data["input_tokens"].(int); ok {
		if tokens < 0 {
			errors = append(errors, "input_tokens cannot be negative")
		}
	}
	if tokens, ok := data["output_tokens"].(int); ok {
		if tokens < 0 {
			errors = append(errors, "output_tokens cannot be negative")
		}
	}
	if cost, ok := data["cost"].(float64); ok {
		if cost < 0 {
			errors = append(errors, "cost cannot be negative")
		}
	}
	return errors
}
