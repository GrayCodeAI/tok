package filtervariants

import "testing"

func TestNewVariantSelector(t *testing.T) {
	vs := NewVariantSelector()
	if vs == nil {
		t.Fatal("Expected non-nil selector")
	}

	variants := vs.GetVariants()
	if len(variants) < 5 {
		t.Errorf("Expected at least 5 built-in variants, got %d", len(variants))
	}
}

func TestSelectVariant(t *testing.T) {
	vs := NewVariantSelector()

	variant := vs.SelectVariant("go test", "PASS", "/tmp")
	if variant == nil {
		t.Error("Expected a variant to be selected")
	}
}

func TestAddVariant(t *testing.T) {
	vs := NewVariantSelector()
	initial := len(vs.GetVariants())

	vs.AddVariant(Variant{
		Name:       "custom",
		Type:       VariantOutputPattern,
		Priority:   5,
		FilterName: "custom_filter",
	})

	if len(vs.GetVariants()) != initial+1 {
		t.Error("Expected variant to be added")
	}
}
