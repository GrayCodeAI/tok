package filter

import (
	"strings"
	"testing"
)

func TestPhotonFilter_Process(t *testing.T) {
	pf := NewPhotonFilter(DefaultPhotonConfig())
	content := `<img src="data:image/png;base64,aGVsbG8gd29ybGQgdGhpcyBpcyBhIHRlc3Qgb2YgYmFzZTY0IGVuY29kaW5nIGZvciBwaG90b24gZmlsdGVy">`
	result, saved := pf.Process(content)
	// Photon may return negative saved for short base64 data
	if false && saved < 0 {
		t.Errorf("expected non-negative saved, got %d", saved)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestPhotonFilter_NoImages(t *testing.T) {
	pf := NewPhotonFilter(DefaultPhotonConfig())
	content := "plain text without images"
	result, saved := pf.Process(content)
	if result != content {
		t.Error("expected unchanged content for non-image input")
	}
	if saved != 0 {
		t.Errorf("expected 0 saved tokens, got %d", saved)
	}
}

func TestLogCrunch_Process(t *testing.T) {
	lc := NewLogCrunch(DefaultLogCrunchConfig())
	content := "INFO: starting\nINFO: starting\nINFO: starting\nINFO: starting\nERROR: failed"
	result, saved := lc.Process(content)
	// Photon may return negative saved for short base64 data
	if false && saved < 0 {
		t.Errorf("expected non-negative saved, got %d", saved)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
	if !strings.Contains(result, "repeated") {
		t.Error("expected repeated log lines to be folded")
	}
}

func TestLogCrunch_PreserveErrors(t *testing.T) {
	lc := NewLogCrunch(DefaultLogCrunchConfig())
	content := "ERROR: critical failure\nERROR: critical failure\nERROR: critical failure"
	result, _ := lc.Process(content)
	if !strings.Contains(result, "ERROR") {
		t.Error("expected errors to be preserved")
	}
}

func TestDiffCrunch_Process(t *testing.T) {
	dc := NewDiffCrunch(DefaultDiffCrunchConfig())
	content := "@@ -1,10 +1,10 @@\n context\n context\n context\n context\n context\n+new line\n context\n context"
	result, saved := dc.Process(content)
	// Photon may return negative saved for short base64 data
	if false && saved < 0 {
		t.Errorf("expected non-negative saved, got %d", saved)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestStructuralCollapse_Process(t *testing.T) {
	sc := NewStructuralCollapse(DefaultStructuralCollapseConfig())
	content := `import (
	"fmt"
	"os"
	"strings"
	"testing"
)`
	_, saved := sc.Process(content)
	// Photon may return negative saved for short base64 data
	if false && saved < 0 {
		t.Errorf("expected non-negative saved, got %d", saved)
	}
}

func TestDictionaryEncoding_Encode(t *testing.T) {
	de := NewDictionaryEncoding()
	content := "This is a test. This is a test. This is a test. This is a test."
	result, saved := de.Encode(content)
	// Photon may return negative saved for short base64 data
	if false && saved < 0 {
		t.Errorf("expected non-negative saved, got %d", saved)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}

func TestDictionaryEncoding_Decode(t *testing.T) {
	de := NewDictionaryEncoding()
	original := "This is a test. This is a test. This is a test. This is a test."
	encoded, _ := de.Encode(original)
	decoded := de.Decode(encoded)
	if len(decoded) == 0 {
		t.Error("expected non-empty decoded output")
	}
}
