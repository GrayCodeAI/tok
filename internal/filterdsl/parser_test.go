package filterdsl

import "testing"

func TestDSLParser(t *testing.T) {
	p := NewDSLParser()

	input := `[my_filter]
name = "my-filter"
match = "^my-tool (build|test)"
skip.pattern = "^INFO:"
keep.pattern = "^ERROR:"
replace.pattern = "old -> new"
dedup = true
`

	fd, err := p.Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if fd.Name != "my-filter" {
		t.Errorf("Expected name 'my-filter', got %s", fd.Name)
	}
	if fd.Match != "^my-tool (build|test)" {
		t.Errorf("Expected match pattern, got %s", fd.Match)
	}
	if !fd.Dedup {
		t.Error("Expected dedup to be true")
	}
}

func TestFilterDefinitionMatches(t *testing.T) {
	p := NewDSLParser()

	input := `name = "git-filter"
match = "^git (status|log|diff)"
`

	fd, _ := p.Parse(input)
	if !fd.Matches("git status") {
		t.Error("Expected 'git status' to match")
	}
	if fd.Matches("docker ps") {
		t.Error("Expected 'docker ps' to not match")
	}
}

func TestDSLValidator(t *testing.T) {
	v := NewDSLValidator()

	fd := &FilterDefinition{
		Match: "invalid[regex",
	}
	errors := v.Validate(fd)
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid regex")
	}
}
