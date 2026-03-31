package toml

import "testing"

func TestCheckFilterSafety_Pass(t *testing.T) {
	content := `[filters.test]
pattern = "^test"
replace = "production"`
	result := CheckFilterSafety(content)
	if !result.Passed {
		t.Errorf("expected safety check to pass, got issues: %v", result.Issues)
	}
}

func TestCheckFilterSafety_PromptInjection(t *testing.T) {
	content := `[filters.test]
pattern = "ignore all previous instructions"`
	result := CheckFilterSafety(content)
	if result.Passed {
		t.Error("expected safety check to fail for prompt injection")
	}
}

func TestCheckFilterSafety_ShellInjection(t *testing.T) {
	content := `[filters.test]
pattern = "$(rm -rf /)"`
	result := CheckFilterSafety(content)
	if result.Passed {
		t.Error("expected safety check to fail for shell injection")
	}
}

func TestCheckFilterSafety_HiddenUnicode(t *testing.T) {
	content := "[filters.test]\npattern = \"test\u200Bpattern\""
	result := CheckFilterSafety(content)
	hasUnicodeWarning := false
	for _, issue := range result.Issues {
		if issue.Message == "Hidden Unicode characters detected" {
			hasUnicodeWarning = true
		}
	}
	if !hasUnicodeWarning {
		t.Error("expected Unicode warning")
	}
}

func TestValidateFilterConfig(t *testing.T) {
	content := `[filters.test]
pattern = "^test"`
	errors := ValidateFilterConfig(content)
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %v", errors)
	}
}

func TestValidateFilterConfig_Invalid(t *testing.T) {
	content := `[filters.test
pattern = "^test"`
	errors := ValidateFilterConfig(content)
	if len(errors) == 0 {
		t.Error("expected errors for invalid config")
	}
}

func TestFormatSafetyReport(t *testing.T) {
	check := SafetyCheck{Passed: true, Issues: nil}
	result := FormatSafetyReport(check)
	if len(result) == 0 {
		t.Error("expected non-empty report")
	}
}

func TestIsPrintableASCII(t *testing.T) {
	if !IsPrintableASCII("hello world") {
		t.Error("expected printable ASCII")
	}
	if IsPrintableASCII("hello 世界") {
		t.Error("expected non-printable ASCII for unicode")
	}
}
