package linter

import (
	"strings"
	"testing"
)

func TestFilterRubocopJSON_NoOffenses(t *testing.T) {
	input := `{"files":[{"path":"app/models/user.rb","offenses":[]}],"summary":{"offense_count":0,"inspected_file_count":15,"correctable_offense_count":0}}`
	result := filterRubocopJSON(input)
	if !strings.Contains(result, "ok ✓ rubocop") {
		t.Errorf("expected 'ok ✓ rubocop', got: %s", result)
	}
	if !strings.Contains(result, "15 files") {
		t.Errorf("expected file count, got: %s", result)
	}
}

func TestFilterRubocopJSON_Empty(t *testing.T) {
	result := filterRubocopJSON("")
	if !strings.Contains(result, "No output") {
		t.Errorf("expected 'No output', got: %s", result)
	}
}

func TestFilterRubocopJSON_WithOffenses(t *testing.T) {
	input := `{"files":[
		{"path":"/home/user/project/app/models/user.rb","offenses":[
			{"cop_name":"Style/StringLiterals","severity":"convention","message":"Prefer single-quoted strings","correctable":true,"location":{"start_line":10}},
			{"cop_name":"Lint/UselessAssignment","severity":"warning","message":"Useless assignment to variable x","correctable":false,"location":{"start_line":25}}
		]}
	],"summary":{"offense_count":2,"inspected_file_count":1,"correctable_offense_count":1}}`
	result := filterRubocopJSON(input)
	if !strings.Contains(result, "2 offenses") {
		t.Errorf("expected offense count, got: %s", result)
	}
	if !strings.Contains(result, "Lint/UselessAssignment") {
		t.Errorf("expected cop name, got: %s", result)
	}
	if !strings.Contains(result, "correctable") {
		t.Errorf("expected correctable hint, got: %s", result)
	}
}

func TestFilterRubocopJSON_SeveritySorting(t *testing.T) {
	input := `{"files":[
		{"path":"app/models/user.rb","offenses":[
			{"cop_name":"Style/StringLiterals","severity":"convention","message":"Prefer single-quoted strings","correctable":true,"location":{"start_line":10}}
		]},
		{"path":"app/controllers/users_controller.rb","offenses":[
			{"cop_name":"Lint/UselessAssignment","severity":"error","message":"Useless assignment","correctable":false,"location":{"start_line":5}}
		]}
	],"summary":{"offense_count":2,"inspected_file_count":2,"correctable_offense_count":1}}`
	result := filterRubocopJSON(input)
	// Error severity file should come first
	errorIdx := strings.Index(result, "users_controller.rb")
	conventionIdx := strings.Index(result, "user.rb")
	if errorIdx > conventionIdx {
		t.Errorf("error severity file should come before convention file, got:\n%s", result)
	}
}

func TestFilterRubocopJSON_Overflow(t *testing.T) {
	var files []string
	for i := 0; i < 12; i++ {
		files = append(files, `{"path":"file`+string(rune('a'+i))+`.rb","offenses":[{"cop_name":"Style/Test","severity":"convention","message":"test","correctable":false,"location":{"start_line":1}}]}`)
	}
	input := `{"files":[` + strings.Join(files, ",") + `],"summary":{"offense_count":12,"inspected_file_count":12,"correctable_offense_count":0}}`
	result := filterRubocopJSON(input)
	if !strings.Contains(result, "... +2 more files") {
		t.Errorf("expected file overflow, got: %s", result)
	}
}

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		severity string
		want     int
	}{
		{"fatal", 0},
		{"error", 0},
		{"warning", 1},
		{"convention", 2},
		{"refactor", 2},
		{"info", 2},
		{"unknown", 3},
	}
	for _, tt := range tests {
		got := severityRank(tt.severity)
		if got != tt.want {
			t.Errorf("severityRank(%q) = %d, want %d", tt.severity, got, tt.want)
		}
	}
}

func TestFilterRubocopText_Autocorrect(t *testing.T) {
	input := `Inspecting 15 files
.C.C....C.C.....

15 files inspected, 5 offenses detected, 5 offenses autocorrectable
`
	result := filterRubocopText(input)
	if !strings.Contains(result, "15 files inspected") {
		t.Errorf("expected summary line, got: %s", result)
	}
	if !strings.Contains(result, "offenses autocorrectable") {
		t.Errorf("expected autocorrect count, got: %s", result)
	}
}

func TestFilterRubocopText_BundlerError(t *testing.T) {
	input := `Could not find gem 'rubocop (~> 1.60)' in any of the relevant sources
Bundler::GemNotFound
`
	result := filterRubocopText(input)
	if !strings.Contains(result, "Could not find") {
		t.Errorf("expected error message, got: %s", result)
	}
}

func TestCompactRubyPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/project/app/models/user.rb", "app/models/user.rb"},
		{"/home/user/project/lib/my_gem.rb", "lib/my_gem.rb"},
		{"/home/user/project/spec/models/user_spec.rb", "spec/models/user_spec.rb"},
		{"app/models/user.rb", "app/models/user.rb"},
	}
	for _, tt := range tests {
		got := compactRubyPath(tt.input)
		if got != tt.want {
			t.Errorf("compactRubyPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
