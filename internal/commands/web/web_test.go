package web

import (
	"testing"
)

func TestCurlCmdInitialized(t *testing.T) {
	if curlCmd == nil {
		t.Fatal("curlCmd should be initialized")
	}
	if curlCmd.Use != "curl [flags...] <URL>" {
		t.Errorf("expected Use 'curl [flags...] <URL>', got %q", curlCmd.Use)
	}
}

func TestWgetCmdInitialized(t *testing.T) {
	if wgetCmd == nil {
		t.Fatal("wgetCmd should be initialized")
	}
	if wgetCmd.Use != "wget [flags...] <URL>" {
		t.Errorf("expected Use 'wget [flags...] <URL>', got %q", wgetCmd.Use)
	}
}

func TestParseCurlOutput(t *testing.T) {
	// Body + content-type + status code separated by newlines
	raw := "{\"ok\":true}\napplication/json\n200"
	body, ct, code := parseCurlOutput(raw)
	if body != "{\"ok\":true}" {
		t.Errorf("expected body '{\"ok\":true}', got %q", body)
	}
	if ct != "application/json" {
		t.Errorf("expected content-type 'application/json', got %q", ct)
	}
	if code != "200" {
		t.Errorf("expected status code '200', got %q", code)
	}
}
