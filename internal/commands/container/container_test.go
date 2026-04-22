package container

import (
	"strings"
	"testing"

	"github.com/GrayCodeAI/tok/internal/commands/shared"
)

func TestCompactPorts(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "-"},
		{"single", "8080->80/tcp", "8080"},
		{"with host", "0.0.0.0:8080->80/tcp", "8080"},
		{"multiple", "8080->80/tcp, 8443->443/tcp", "8080, 8443"},
		{"many", "80->80, 443->443, 8080->8080, 9090->9090", "80, 443, ... +2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compactPorts(tt.input)
			if !strings.Contains(got, strings.Split(tt.want, ",")[0]) {
				t.Errorf("compactPorts(%q) = %q, expected to contain %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFilterLogOutput(t *testing.T) {
	input := "line1\nline1\nline1\nline2\nline2\nline3"
	got := filterLogOutput(input)
	if !strings.Contains(got, "repeated") {
		t.Errorf("expected repeated count in output, got %q", got)
	}
}

func TestFilterDockerBuild(t *testing.T) {
	input := "Step 1/3 : FROM ubuntu\n ---> abc123\nStep 2/3 : RUN apt-get update\n ---> def456\nSuccessfully built xyz789"
	got := filterDockerBuild(input)
	if !strings.Contains(got, "Build:") {
		t.Errorf("expected Build: in output, got %q", got)
	}
}

func TestFilterDockerStats(t *testing.T) {
	input := "CONTAINER ID   NAME   CPU %   MEM USAGE   NET I/O   BLOCK I/O   PIDS\nabc123   web   0.50%   100MiB   1kB / 2kB   0B / 0B   5"
	got := filterDockerStats(input)
	if !strings.Contains(got, "cpu=") {
		t.Errorf("expected cpu= in output, got %q", got)
	}
}

func TestFilterDockerTop(t *testing.T) {
	input := "UID   PID   PPID   C   STIME   TTY   TIME   CMD\nroot   1   0   0   00:00   ?   00:00:00   /bin/bash"
	got := filterDockerTop(input)
	if got == "" {
		t.Error("expected non-empty output")
	}
}

func TestFilterDockerSystemDf(t *testing.T) {
	input := "TYPE            TOTAL     ACTIVE    SIZE      RECLAIMABLE\nImages          5         2         1.5GB     500MB (33%)"
	got := filterDockerSystemDf(input)
	if !strings.Contains(got, "Images") {
		t.Errorf("expected Images in output, got %q", got)
	}
}

func TestFilterDockerInfo(t *testing.T) {
	input := "Server Version: 24.0.0\nStorage Driver: overlay2\nOperating System: Ubuntu 22.04\nArchitecture: x86_64\nCPUs: 4\nTotal Memory: 8GiB"
	got := filterDockerInfo(input)
	if !strings.Contains(got, "Server Version") {
		t.Errorf("expected Server Version in output, got %q", got)
	}
}

func TestFilterDockerVersion(t *testing.T) {
	input := "Client:\n Version: 24.0.0\n API version: 1.43\nServer:\n Version: 24.0.0\n API version: 1.43"
	got := filterDockerVersion(input)
	if !strings.Contains(got, "Client:") {
		t.Errorf("expected Client: in output, got %q", got)
	}
}

func TestFilterDockerLifecycle(t *testing.T) {
	orig := shared.UltraCompact
	defer func() { shared.UltraCompact = orig }()
	shared.UltraCompact = true

	got := filterDockerLifecycle("", []string{"stop"})
	if !strings.Contains(got, "ok") {
		t.Errorf("expected ok in output, got %q", got)
	}
}

func TestFilterDockerRunUltraCompact(t *testing.T) {
	got := filterDockerRunUltraCompact("line1\nline2\nline3\n")
	if !strings.Contains(got, "3 lines") {
		t.Errorf("expected '3 lines' in output, got %q", got)
	}
}

func TestFilterComposeUp(t *testing.T) {
	input := "Creating web_1 ... done\nStarting api_1 ... done\nError: failed to start db_1"
	got := filterComposeUp(input)
	if !strings.Contains(got, "Creating") {
		t.Errorf("expected Creating in output, got %q", got)
	}
}

func TestFilterComposeDown(t *testing.T) {
	input := "Removing web_1 ... done\nRemoving network default\n"
	got := filterComposeDown(input)
	if !strings.Contains(got, "Down:") {
		t.Errorf("expected Down: in output, got %q", got)
	}
}

func TestFilterComposeBuild(t *testing.T) {
	input := "Building web\nStep 1/2 : FROM node:18\n[+] Building 2.3s (5/5) FINISHED"
	got := filterComposeBuild(input)
	if !strings.Contains(got, "Build:") {
		t.Errorf("expected Build: in output, got %q", got)
	}
}

func TestCompactDockerInspect(t *testing.T) {
	input := `[{"Name":"/web","State":{"Status":"running"},"Config":{"Image":"nginx:latest"}}]`
	got := compactDockerInspect(input)
	if !strings.Contains(got, "web") {
		t.Errorf("expected web in output, got %q", got)
	}
}

func TestCompactDockerInspect_InvalidJSON(t *testing.T) {
	input := "not json"
	got := compactDockerInspect(input)
	if got != input {
		t.Errorf("expected unchanged output for invalid JSON, got %q", got)
	}
}

func TestFormatAsJSON(t *testing.T) {
	got := formatAsJSON("hello")
	if !strings.Contains(got, "hello") {
		t.Errorf("expected hello in JSON, got %q", got)
	}
}

func TestCompactJSON(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"age":  30,
	}
	var sb strings.Builder
	compactJSON(data, "", &sb, 2)
	got := sb.String()
	if !strings.Contains(got, "name:") {
		t.Errorf("expected name: in output, got %q", got)
	}
}

func TestFilterDockerRunOutput(t *testing.T) {
	input := "Starting container...\nError: port already allocated"
	got := filterDockerRunOutput(input)
	if !strings.Contains(got, "Error") {
		t.Errorf("expected Error in output, got %q", got)
	}
}
