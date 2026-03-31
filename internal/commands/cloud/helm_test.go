package cloud

import (
	"testing"
)

func TestFilterHelmListOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "with releases",
			input: `NAME           	NAMESPACE	REVISION	UPDATED                                	STATUS  	CHART                	APP VERSION
my-app         	default  	1       	2026-03-01 10:00:00.000000 +0000 UTC	deployed	my-app-1.0.0         	1.0.0
my-service     	default  	3       	2026-03-02 12:00:00.000000 +0000 UTC	deployed	my-service-2.0.0     	2.0.0`,
		},
		{
			name:  "no releases",
			input: "",
		},
		{
			name:  "single release",
			input: `my-app	default	1	2026-03-01 10:00:00 +0000 UTC	deployed	my-app-1.0.0	1.0.0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHelmListOutput(tt.input)
			if result == "" {
				t.Error("filterHelmListOutput() returned empty string")
			}
		})
	}
}

func TestFilterHelmStatusOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "deployed status",
			input: `NAME: my-app
LAST DEPLOYED: Sat Mar  1 10:00:00 2026
NAMESPACE: default
STATUS: deployed
REVISION: 1`,
		},
		{
			name: "failed status",
			input: `NAME: my-app
LAST DEPLOYED: Sat Mar  1 10:00:00 2026
NAMESPACE: default
STATUS: failed
REVISION: 3`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHelmStatusOutput(tt.input)
			if result == "" {
				t.Error("filterHelmStatusOutput() returned empty string")
			}
		})
	}
}

func TestFilterHelmUpgradeOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful upgrade",
			input: `Release "my-app" has been upgraded. Happy Helming!
NAME: my-app
LAST DEPLOYED: Sun Mar  1 10:00:00 2026
NAMESPACE: default
STATUS: deployed
REVISION: 2`,
		},
		{
			name:  "upgrade error",
			input: "Error: UPGRADE FAILED: timed out waiting for the condition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHelmUpgradeOutput(tt.input)
			if result == "" {
				t.Error("filterHelmUpgradeOutput() returned empty string")
			}
		})
	}
}

func TestFilterHelmInstallOutput(t *testing.T) {
	input := `NAME: my-app
LAST DEPLOYED: Sun Mar  1 10:00:00 2026
NAMESPACE: default
STATUS: deployed
REVISION: 1

NOTES:
1. Get the application URL by running:
  export POD_NAME=$(kubectl get pods -l "app=my-app" -o jsonpath="{.items[0].metadata.name}")`
	result := filterHelmInstallOutput(input)
	if result == "" {
		t.Error("filterHelmInstallOutput() returned empty string")
	}
}

func TestFilterHelmRollbackOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "successful rollback",
			input: "Rollback was a success! Happy Helming!",
		},
		{
			name:  "rollback error",
			input: "Error: rollback failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHelmRollbackOutput(tt.input)
			if result == "" {
				t.Error("filterHelmRollbackOutput() returned empty string")
			}
		})
	}
}

func TestFilterHelmRepoUpdateOutput(t *testing.T) {
	input := `Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "stable" chart repository
...Successfully got an update from the "bitnami" chart repository
Update Complete. Happy Helming!`
	result := filterHelmRepoUpdateOutput(input)
	if result == "" {
		t.Error("filterHelmRepoUpdateOutput() returned empty string")
	}
}

func TestFilterHelmOutput(t *testing.T) {
	input := "line1\nline2\nline3"
	result := filterHelmOutput(input)
	if result == "" {
		t.Error("filterHelmOutput() returned empty string")
	}
}

func TestFilterHelmOutputTruncation(t *testing.T) {
	input := ""
	for i := 0; i < 40; i++ {
		input += "line\n"
	}
	result := filterHelmOutput(input)
	if len(result) == 0 {
		t.Error("filterHelmOutput() returned empty string")
	}
}
