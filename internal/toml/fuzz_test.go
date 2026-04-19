package toml

import (
	"testing"
)

func FuzzTOMLFilterParse(f *testing.F) {
	seeds := []string{
		`name = "test"
[pattern]
match = "ERROR"`,
		`name = ""
[pattern]
match = ""`,
		`[pattern]
match = ".*"
replace = ""`,
		`name = "test"`,
		``,
		`[pattern]`,
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		p := NewParser()
		_, _ = p.ParseContent([]byte(input), "fuzz")
	})
}
