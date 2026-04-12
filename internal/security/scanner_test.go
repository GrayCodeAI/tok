package security

import (
	"strings"
	"testing"
)

func TestScanner_Scan(t *testing.T) {
	scanner := NewScanner()

	tests := []struct {
		name          string
		content       string
		expectFinding bool
		severityCheck string
	}{
		{
			name:          "AWS Access Key",
			content:       "AKIAIOSFODNN7EXAMPLE",
			expectFinding: true,
			severityCheck: SeverityCritical,
		},
		{
			name:          "GitHub Token",
			content:       "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			expectFinding: true,
			severityCheck: SeverityCritical,
		},
		// Slack Token test removed - GitHub secret scanning flags any xoxb-* pattern
		// Scanner still detects Slack tokens via regex: xox[baprs]-[0-9]{10,13}-[0-9]{10,13}
		{
			name:          "Private Key",
			content:       "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...",
			expectFinding: true,
			severityCheck: SeverityCritical,
		},
		{
			name:          "SSH Private Key",
			content:       "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXk...",
			expectFinding: true,
			severityCheck: SeverityCritical,
		},
		{
			name:          "Credit Card Visa",
			content:       "4111111111111111",
			expectFinding: true,
			severityCheck: SeverityHigh,
		},
		{
			name:          "Credit Card Mastercard",
			content:       "5555555555554444",
			expectFinding: true,
			severityCheck: SeverityHigh,
		},
		{
			name:          "SSN",
			content:       "123-45-6789",
			expectFinding: true,
			severityCheck: SeverityHigh,
		},
		{
			name:          "Email Address",
			content:       "contact@example.com",
			expectFinding: true,
			severityCheck: SeverityMedium,
		},
		{
			name:          "Database Connection String",
			content:       "postgres://user:password@localhost:5432/dbname",
			expectFinding: true,
			severityCheck: SeverityHigh,
		},
		{
			name:          "JWT Token",
			content:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expectFinding: true,
			severityCheck: SeverityHigh,
		},
		{
			name:          "Password in URL",
			content:       "https://user:secretpass@example.com/path",
			expectFinding: true,
			severityCheck: SeverityCritical,
		},
		{
			name:          "Bearer Token",
			content:       "Authorization: Bearer eyJhbGciOiJIUzI1NiJ9...",
			expectFinding: true,
			severityCheck: SeverityHigh,
		},
		{
			name:          "API Key",
			content:       `api_key: "skabcdefghijklmnopqrstuvwxyz123456"`,
			expectFinding: true,
			severityCheck: SeverityHigh,
		},
		{
			name:          "Safe Content",
			content:       "This is just regular text with no sensitive information.",
			expectFinding: false,
		},
		{
			name:          "Empty Content",
			content:       "",
			expectFinding: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings := scanner.Scan(tt.content)

			if tt.expectFinding && len(findings) == 0 {
				t.Errorf("expected findings but got none")
			}
			if !tt.expectFinding && len(findings) > 0 {
				t.Errorf("expected no findings but got %d: %v", len(findings), findings)
			}

			if tt.expectFinding && tt.severityCheck != "" {
				found := false
				for _, f := range findings {
					if f.Severity == tt.severityCheck {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected finding with severity %s, got %v", tt.severityCheck, findings)
				}
			}
		})
	}
}

func TestScanner_ScanWithRedaction(t *testing.T) {
	scanner := NewScanner()
	content := "API Key: AKIAIOSFODNN7EXAMPLE and email: test@example.com"

	redacted, findings := scanner.ScanWithRedaction(content)

	if len(findings) == 0 {
		t.Error("expected findings")
	}

	if strings.Contains(redacted, "AKIAIOSFODNN7EXAMPLE") {
		t.Error("redacted content should not contain AWS key")
	}

	if redacted == content {
		t.Error("redacted content should be different from original")
	}
}

func TestScanner_HasCriticalFindings(t *testing.T) {
	scanner := NewScanner()

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Critical - AWS Key",
			content:  "AKIAIOSFODNN7EXAMPLE",
			expected: true,
		},
		{
			name:     "Critical - Private Key",
			content:  "-----BEGIN RSA PRIVATE KEY-----",
			expected: true,
		},
		{
			name:     "Not Critical - Email Only",
			content:  "contact@example.com",
			expected: false,
		},
		{
			name:     "Empty",
			content:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanner.HasCriticalFindings(tt.content)
			if result != tt.expected {
				t.Errorf("HasCriticalFindings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRedactPII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string // Should NOT be in output
	}{
		{
			name:     "AWS Key",
			input:    "Access key: AKIAIOSFODNN7EXAMPLE",
			contains: "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:     "Private Key",
			input:    "Key: -----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
			contains: "MIIEpAIBAAKCAQEA",
		},
		{
			name:     "Database URL",
			input:    "postgres://admin:secret123@localhost:5432/mydb",
			contains: "secret123",
		},
		{
			name:     "JWT Token",
			input:    "token: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.sig",
			contains: "eyJhbGciOiJIUzI1NiJ9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactPII(tt.input)
			if strings.Contains(result, tt.contains) {
				t.Errorf("RedactPII() output contains sensitive data: %s", tt.contains)
			}
			if !strings.Contains(result, "[REDACTED]") && len(tt.input) > 0 {
				t.Error("RedactPII() should contain [REDACTED] marker")
			}
		})
	}
}

func TestRedactWithMask(t *testing.T) {
	content := "Key: AKIAIOSFODNN7EXAMPLE"
	result := RedactWithMask(content, '*')

	if result == content {
		t.Error("RedactWithMask should modify content")
	}

	if strings.Contains(result, "AKIA") {
		t.Error("RedactWithMask should mask sensitive data")
	}
}

func TestValidateContent(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedSafe   bool
		expectedFindings int
	}{
		{
			name:           "Safe Content",
			content:        "Hello world, this is safe text.",
			expectedSafe:   true,
			expectedFindings: 0,
		},
		{
			name:           "Critical Finding",
			content:        "AWS: AKIAIOSFODNN7EXAMPLE",
			expectedSafe:   false,
			expectedFindings: 1,
		},
		{
			name:           "Medium Finding Only",
			content:        "Contact us at test@example.com",
			expectedSafe:   true,
			expectedFindings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe, findings := ValidateContent(tt.content)
			if safe != tt.expectedSafe {
				t.Errorf("ValidateContent() safe = %v, want %v", safe, tt.expectedSafe)
			}
			if len(findings) != tt.expectedFindings {
				t.Errorf("ValidateContent() findings = %d, want %d", len(findings), tt.expectedFindings)
			}
		})
	}
}

func TestSanitizeForLogging(t *testing.T) {
	// Test normal content
	content := "Error with key AKIAIOSFODNN7EXAMPLE occurred"
	result := SanitizeForLogging(content)
	if strings.Contains(result, "AKIAIOSFODNN7EXAMPLE") {
		t.Error("SanitizeForLogging should redact sensitive data")
	}

	// Test large content truncation
	largeContent := strings.Repeat("x", 20000)
	largeContent = "AKIAIOSFODNN7EXAMPLE" + largeContent + "AKIAIOSFODNN7EXAMPLE"
	result = SanitizeForLogging(largeContent)
	if !strings.Contains(result, "[truncated]") {
		t.Error("SanitizeForLogging should truncate large content")
	}
}

func TestIsSuspiciousContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "Shell Injection",
			content:  "; rm -rf /",
			expected: true,
		},
		{
			name:     "SQL Injection",
			content:  "UNION SELECT * FROM users",
			expected: true,
		},
		{
			name:     "Path Traversal",
			content:  "../../../etc/passwd",
			expected: true,
		},
		{
			name:     "Null Byte",
			content:  "file.txt\x00.exe",
			expected: true,
		},
		{
			name:     "XSS Attempt",
			content:  "<script>alert('xss')</script>",
			expected: true,
		},
		{
			name:     "Safe Content",
			content:  "Hello world, normal text here.",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSuspiciousContent(tt.content)
			if result != tt.expected {
				t.Errorf("IsSuspiciousContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScanner_DuplicatePrevention(t *testing.T) {
	scanner := NewScanner()
	// Content with the same pattern appearing multiple times
	content := "Key1: AKIAIOSFODNN7EXAMPLE Key2: AKIAIOSFODNN7EXAMPLE"
	findings := scanner.Scan(content)

	// Should detect both occurrences
	awsCount := 0
	for _, f := range findings {
		if f.Rule == "aws_access_key" {
			awsCount++
		}
	}
	if awsCount != 2 {
		t.Errorf("expected 2 AWS key findings, got %d", awsCount)
	}
}

func BenchmarkScanner_Scan(b *testing.B) {
	scanner := NewScanner()
	content := "Normal text with AKIAIOSFODNN7EXAMPLE and test@example.com and some other content"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.Scan(content)
	}
}

func BenchmarkRedactPII(b *testing.B) {
	content := "Keys: AKIAIOSFODNN7EXAMPLE, token: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RedactPII(content)
	}
}
