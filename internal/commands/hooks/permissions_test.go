package hooks

import "testing"

func TestCheckCommandWithRules(t *testing.T) {
	tests := []struct {
		name  string
		cmd   string
		deny  []string
		ask   []string
		allow []string
		want  PermissionVerdict
	}{
		{
			name:  "deny wins",
			cmd:   "rm -rf dist",
			deny:  []string{"rm *"},
			ask:   []string{"rm -rf *"},
			allow: []string{"*"},
			want:  PermissionDeny,
		},
		{
			name:  "ask rule",
			cmd:   "git push origin main",
			ask:   []string{"git push*"},
			allow: []string{"git *"},
			want:  PermissionAsk,
		},
		{
			name:  "allow when every segment allowed",
			cmd:   "git status && git log -5",
			allow: []string{"git *"},
			want:  PermissionAllow,
		},
		{
			name:  "default when one segment not allowed",
			cmd:   "git status && echo done",
			allow: []string{"git *"},
			want:  PermissionDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkCommandWithRules(tt.cmd, tt.deny, tt.ask, tt.allow)
			if got != tt.want {
				t.Fatalf("verdict = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildCopilotVSCodeResponse(t *testing.T) {
	origCheck := checkCommandPermissions
	defer func() { checkCommandPermissions = origCheck }()

	tests := []struct {
		name      string
		verdict   PermissionVerdict
		want      string
		wantInput bool
	}{
		{name: "allow", verdict: PermissionAllow, want: "allow", wantInput: true},
		{name: "ask", verdict: PermissionAsk, want: "ask", wantInput: true},
		{name: "default", verdict: PermissionDefault, want: "ask", wantInput: true},
		{name: "deny", verdict: PermissionDeny, want: "deny", wantInput: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkCommandPermissions = func(string) PermissionVerdict { return tt.verdict }
			out := buildCopilotVSCodeResponse("git status", "tokman git status")
			hookOutput := out["hookSpecificOutput"].(map[string]any)
			if hookOutput["permissionDecision"] != tt.want {
				t.Fatalf("permissionDecision = %v, want %s", hookOutput["permissionDecision"], tt.want)
			}
			_, hasInput := hookOutput["updatedInput"]
			if hasInput != tt.wantInput {
				t.Fatalf("updatedInput present = %v, want %v", hasInput, tt.wantInput)
			}
		})
	}
}
