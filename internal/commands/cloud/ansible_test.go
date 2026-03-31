package cloud

import (
	"testing"
)

func TestFilterAnsibleOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful playbook",
			input: `PLAY [Deploy Application] *********************************************************

TASK [Gathering Facts] ***********************************************************
ok: [web-server]

TASK [Install nginx] *************************************************************
changed: [web-server]

TASK [Start nginx] ***************************************************************
ok: [web-server]

PLAY RECAP ***********************************************************************
web-server                 : ok=3    changed=1    unreachable=0    failed=0`,
		},
		{
			name: "playbook with failure",
			input: `PLAY [Deploy Application] *********************************************************

TASK [Gathering Facts] ***********************************************************
ok: [web-server]

TASK [Install package] ***********************************************************
fatal: [web-server]: FAILED! => {"changed": false, "msg": "No package matching 'nonexistent' found"}

PLAY RECAP ***********************************************************************
web-server                 : ok=1    changed=0    unreachable=0    failed=1`,
		},
		{
			name:  "empty output",
			input: "",
		},
		{
			name: "skipped tasks",
			input: `PLAY [Deploy] ********************************************************************

TASK [Install nginx] *************************************************************
skipping: [web-server]

PLAY RECAP ***********************************************************************
web-server                 : ok=1    changed=0    unreachable=0    failed=0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterAnsibleOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterAnsibleOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestExtractBracketContent(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PLAY [Deploy Application]", "Deploy Application"},
		{"TASK [Install nginx]", "Install nginx"},
		{"no brackets here", "no brackets here"},
		{"[single]", "single"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractBracketContent(tt.input)
			if got != tt.want {
				t.Errorf("extractBracketContent(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterAnsibleOutput_MultipleHosts(t *testing.T) {
	input := `PLAY [Configure Servers] *********************************************************

TASK [Gathering Facts] ***********************************************************
ok: [server1]
ok: [server2]
ok: [server3]

TASK [Install nginx] *************************************************************
changed: [server1]
changed: [server2]
ok: [server3]

PLAY RECAP ***********************************************************************
server1                    : ok=2    changed=1    unreachable=0    failed=0
server2                    : ok=2    changed=1    unreachable=0    failed=0
server3                    : ok=2    changed=0    unreachable=0    failed=0`

	result := filterAnsibleOutput(input)
	if result == "" {
		t.Error("filterAnsibleOutput() returned empty string for multi-host input")
	}
}
