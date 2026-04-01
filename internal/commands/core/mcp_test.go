package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPReadEndpoint(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main\n\nfunc alpha() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	body, _ := json.Marshal(MCPReadRequest{
		Path:         path,
		Mode:         "auto",
		MaxTokens:    50,
		SaveSnapshot: true,
	})
	req := httptest.NewRequest(http.MethodPost, "/read", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newMCPHandler("").ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp MCPReadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if resp.Path != path {
		t.Fatalf("expected path %q, got %q", path, resp.Path)
	}
	if resp.Content == "" {
		t.Fatal("expected non-empty content")
	}
}

func TestMCPBundleEndpoint(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(go.mod) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "pkg", "helper"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(dir, "main.go")
	content := "package main\n\nimport \"example.com/demo/pkg/helper\"\n\nfunc main() { helper.Run() }\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(main.go) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "helper", "helper.go"), []byte("package helper\n\nfunc Run() {}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(helper.go) error = %v", err)
	}

	body, _ := json.Marshal(MCPReadRequest{
		Path:         path,
		Mode:         "graph",
		RelatedFiles: 2,
		MaxTokens:    200,
	})
	req := httptest.NewRequest(http.MethodPost, "/bundle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newMCPHandler("").ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp MCPBundleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if resp.Path != "main.go" {
		t.Fatalf("expected relative path main.go, got %q", resp.Path)
	}
	if len(resp.RelatedFiles) == 0 {
		t.Fatal("expected related files")
	}
	if !strings.Contains(resp.Content, "# Related Files") {
		t.Fatalf("expected bundle content with related section, got %q", resp.Content)
	}
}
