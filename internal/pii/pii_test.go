package pii

import "testing"

func TestNewPIIDetector(t *testing.T) {
	d := NewPIIDetector()
	if d == nil {
		t.Fatal("NewPIIDetector() returned nil")
	}
}

func TestDetect_Email(t *testing.T) {
	d := NewPIIDetector()
	findings := d.Detect("Contact user@example.com for help")
	if len(findings) == 0 {
		t.Error("expected email detection")
	}
}

func TestDetect_Phone(t *testing.T) {
	d := NewPIIDetector()
	findings := d.Detect("Call 1-800-555-1234 for support")
	if len(findings) == 0 {
		t.Error("expected phone detection")
	}
}

func TestRedact(t *testing.T) {
	d := NewPIIDetector()
	redacted := d.Redact("My email is user@example.com")
	if redacted == "My email is user@example.com" {
		t.Error("Redact should have changed the output")
	}
}

func TestHasPII(t *testing.T) {
	d := NewPIIDetector()
	if !d.HasPII("ssn: 123-45-6789") {
		t.Error("expected PII detection for SSN")
	}
	if d.HasPII("just normal text here") {
		t.Error("expected no PII in normal text")
	}
}

func TestDetect_NoPII(t *testing.T) {
	d := NewPIIDetector()
	findings := d.Detect("hello world no pii here")
	if len(findings) > 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
}
