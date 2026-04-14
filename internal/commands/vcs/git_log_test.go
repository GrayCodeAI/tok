package vcs

import (
	"strings"
	"testing"
)

func TestFilterLog(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single commit",
			input: "abc1234 fix bug in parser\n",
			want:  []string{"abc1234 fix bug in parser"},
		},
		{
			name:  "multiple commits",
			input: "abc1234 fix bug\ndef5678 add feature\nghi9012 update docs\n",
			want:  []string{"abc1234 fix bug", "def5678 add feature", "ghi9012 update docs"},
		},
		{
			name:  "empty output",
			input: "",
			want:  nil,
		},
		{
			name:  "long commit message",
			input: "abc1234 " + string(make([]byte, 200)) + "\n",
			want:  nil, // just check it doesn't panic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterLog(tt.input)
			if tt.want == nil {
				// Just check no panic
				_ = got
				return
			}
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("filterLog(%q) missing %q, got:\n%s", tt.input, w, got)
				}
			}
		})
	}
}
