package discover

import (
	"testing"
)

func TestRewriteCommand(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		opts     interface{}
		wantCmd  string
		wantBool bool
	}{
		{
			name:     "simple command",
			cmd:      "git status",
			opts:     nil,
			wantCmd:  "git status",
			wantBool: false,
		},
		{
			name:     "command with args",
			cmd:      "docker ps -a",
			opts:     map[string]string{"mode": "verbose"},
			wantCmd:  "docker ps -a",
			wantBool: false,
		},
		{
			name:     "empty command",
			cmd:      "",
			opts:     nil,
			wantCmd:  "",
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotBool := RewriteCommand(tt.cmd, tt.opts)
			if gotCmd != tt.wantCmd {
				t.Errorf("RewriteCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if gotBool != tt.wantBool {
				t.Errorf("RewriteCommand() gotBool = %v, want %v", gotBool, tt.wantBool)
			}
		})
	}
}

func TestDetectCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{
			name: "known command",
			cmd:  "git",
			want: false, // stub returns false
		},
		{
			name: "unknown command",
			cmd:  "unknown-cmd",
			want: false,
		},
		{
			name: "empty command",
			cmd:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectCommand(tt.cmd); got != tt.want {
				t.Errorf("DetectCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKnownCommands(t *testing.T) {
	commands := KnownCommands()

	// Stub returns empty slice
	if commands == nil {
		t.Error("KnownCommands() returned nil")
	}
	if len(commands) != 0 {
		t.Errorf("KnownCommands() returned %d commands, expected 0", len(commands))
	}
}
