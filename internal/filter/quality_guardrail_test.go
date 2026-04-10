package filter

import "testing"

func TestQualityGuardrailValidate_Pass(t *testing.T) {
	g := NewQualityGuardrail()
	before := "ERROR: build failed at main.go:12"
	after := "ERROR: build failed\nmain.go:12"
	res := g.Validate(before, after)
	if !res.Passed {
		t.Fatalf("expected guardrail pass, got reason=%s", res.Reason)
	}
}

func TestQualityGuardrailValidate_FailOnDroppedErrors(t *testing.T) {
	g := NewQualityGuardrail()
	before := "panic: nil pointer dereference"
	after := "summary only"
	res := g.Validate(before, after)
	if res.Passed {
		t.Fatal("expected guardrail failure")
	}
}
