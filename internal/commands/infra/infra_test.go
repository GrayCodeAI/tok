package infra

import (
	"testing"
)

func TestCommandsInitialized(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"ansible", "ansible"},
		{"terraform", "terraform"},
		{"helm", "helm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ansibleCmd == nil && tt.name == "ansible" {
				t.Fatal("ansibleCmd should be initialized")
			}
			if terraformCmd == nil && tt.name == "terraform" {
				t.Fatal("terraformCmd should be initialized")
			}
			if helmCmd == nil && tt.name == "helm" {
				t.Fatal("helmCmd should be initialized")
			}
		})
	}
}

func TestAnsibleCmd(t *testing.T) {
	if ansibleCmd == nil {
		t.Fatal("ansibleCmd should be initialized")
	}
	if ansibleCmd.Use != "ansible [subcommand] [args...]" {
		t.Errorf("expected Use 'ansible [subcommand] [args...]', got %q", ansibleCmd.Use)
	}
}

func TestTerraformCmd(t *testing.T) {
	if terraformCmd == nil {
		t.Fatal("terraformCmd should be initialized")
	}
	if terraformCmd.Use != "terraform [subcommand] [args...]" {
		t.Errorf("expected Use 'terraform [subcommand] [args...]', got %q", terraformCmd.Use)
	}
}

func TestHelmCmd(t *testing.T) {
	if helmCmd == nil {
		t.Fatal("helmCmd should be initialized")
	}
	if helmCmd.Use != "helm [subcommand] [args...]" {
		t.Errorf("expected Use 'helm [subcommand] [args...]', got %q", helmCmd.Use)
	}
}

func TestAtoi(t *testing.T) {
	if got := atoi("42"); got != 42 {
		t.Errorf("atoi(\"42\") = %d, want 42", got)
	}
	if got := atoi("invalid"); got != 0 {
		t.Errorf("atoi(\"invalid\") = %d, want 0", got)
	}
}
