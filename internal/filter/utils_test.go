package filter

import (
	"testing"
)

func TestCleanWord(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello", "hello"},
		{"  HELLO  ", "hello"},
		{"World", "world"},
		{"", ""},
		{"  ", ""},
		{"MiXeD_CaSe", "mixed_case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cleanWord(tt.input)
			if got != tt.want {
				t.Errorf("cleanWord(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple sentence",
			input: "hello world",
			want:  []string{"hello", "world"},
		},
		{
			name:  "with punctuation",
			input: "hello, world!",
			want:  []string{"hello", "world"},
		},
		{
			name:  "multiple spaces",
			input: "hello   world",
			want:  []string{"hello", "world"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "only punctuation",
			input: "...!!!???",
			want:  nil,
		},
		{
			name:  "code-like input",
			input: "func main() { return 42 }",
			want:  []string{"func", "main", "return", "42"},
		},
		{
			name:  "mixed separators",
			input: "hello-world_test.value",
			want:  []string{"hello", "world", "test", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenize(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("tokenize(%q) returned %d tokens, want %d", tt.input, len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func BenchmarkTokenize(b *testing.B) {
	input := "The quick brown fox jumps over the lazy dog. func main() { return 42 }"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tokenize(input)
	}
}
