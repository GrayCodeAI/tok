// Package security provides content scanning and PII redaction for tok.
//
// The security package detects sensitive information in command output such as
// API keys, tokens, credentials, and personally identifiable information (PII).
//
// # Scanner
//
// Use the Scanner to detect sensitive content:
//
//	scanner := security.NewScanner()
//	findings := scanner.Scan(output)
//	for _, f := range findings {
//	    fmt.Printf("Found %s: %s\n", f.Rule, f.Match)
//	}
//
// # Redaction
//
// RedactPII removes detected sensitive information:
//
//	safe := security.RedactPII(output)
//
// # Validation
//
// The Validator type provides input validation for API requests:
//
//	v := security.NewValidator()
//	if err := v.ValidateBudget(budget); err != nil { ... }
package security
