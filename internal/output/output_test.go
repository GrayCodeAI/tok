package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

func TestNewTest_IsNotTTY(t *testing.T) {
	p := NewTest(&bytes.Buffer{}, &bytes.Buffer{})
	if p.IsTTY() {
		t.Error("NewTest printer should not report TTY")
	}
}

func TestPrint_WritesToStdout(t *testing.T) {
	var out, err bytes.Buffer
	p := NewTest(&out, &err)

	p.Print("hello ", "world")
	if got := out.String(); got != "hello world" {
		t.Errorf("stdout = %q, want %q", got, "hello world")
	}
	if err.Len() != 0 {
		t.Errorf("stderr should be empty, got %q", err.String())
	}
}

func TestPrintln_AddsNewline(t *testing.T) {
	var out bytes.Buffer
	p := NewTest(&out, &bytes.Buffer{})

	p.Println("line")
	if got := out.String(); got != "line\n" {
		t.Errorf("stdout = %q, want %q", got, "line\n")
	}
}

func TestPrintf_Formats(t *testing.T) {
	var out bytes.Buffer
	p := NewTest(&out, &bytes.Buffer{})

	p.Printf("n=%d s=%s", 42, "x")
	if got := out.String(); got != "n=42 s=x" {
		t.Errorf("stdout = %q", got)
	}
}

func TestQuietMode_SuppressesStdoutNotStderr(t *testing.T) {
	var out, err bytes.Buffer
	p := NewTest(&out, &err)
	p.SetQuiet(true)

	p.Print("stdout text")
	p.Println("stdout line")
	p.Printf("stdout fmt %d", 1)
	p.Error("stderr text")
	p.Errorf("stderr fmt %d", 2)

	if out.Len() != 0 {
		t.Errorf("stdout should be suppressed in quiet mode, got %q", out.String())
	}
	if !strings.Contains(err.String(), "stderr text") || !strings.Contains(err.String(), "stderr fmt 2") {
		t.Errorf("stderr should not be suppressed, got %q", err.String())
	}
}

func TestJSON_EmitsValidJSON(t *testing.T) {
	var out bytes.Buffer
	p := NewTest(&out, &bytes.Buffer{})

	payload := map[string]any{"a": 1, "b": "two"}
	if err := p.JSON(payload); err != nil {
		t.Fatalf("JSON error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if decoded["b"] != "two" {
		t.Errorf("decoded[b] = %v, want 'two'", decoded["b"])
	}
}

func TestJSON_SuppressedInQuietMode(t *testing.T) {
	var out bytes.Buffer
	p := NewTest(&out, &bytes.Buffer{})
	p.SetQuiet(true)

	if err := p.JSON(map[string]int{"x": 1}); err != nil {
		t.Fatalf("JSON error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("JSON should be suppressed in quiet mode, got %q", out.String())
	}
}

func TestStdoutStderrAccessors(t *testing.T) {
	out := &bytes.Buffer{}
	errW := &bytes.Buffer{}
	p := NewTest(out, errW)

	if p.Stdout() != out {
		t.Error("Stdout() did not return injected writer")
	}
	if p.Stderr() != errW {
		t.Error("Stderr() did not return injected writer")
	}
}

func TestSetGlobal_RoundTrip(t *testing.T) {
	orig := Global()
	defer SetGlobal(orig)

	replacement := NewTest(&bytes.Buffer{}, &bytes.Buffer{})
	prev := SetGlobal(replacement)
	if prev != orig {
		t.Error("SetGlobal did not return the previous global")
	}
	if Global() != replacement {
		t.Error("Global() did not return the replacement")
	}
}

func TestConcurrentWrites_NoRaceNoCorruption(t *testing.T) {
	var out, err bytes.Buffer
	p := NewTest(&out, &err)

	const workers = 16
	const perWorker = 50

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWorker; j++ {
				p.Println("line")
				p.Errorf("e\n")
			}
		}()
	}
	wg.Wait()

	if got := strings.Count(out.String(), "line\n"); got != workers*perWorker {
		t.Errorf("stdout line count = %d, want %d", got, workers*perWorker)
	}
	if got := strings.Count(err.String(), "e\n"); got != workers*perWorker {
		t.Errorf("stderr line count = %d, want %d", got, workers*perWorker)
	}
}
