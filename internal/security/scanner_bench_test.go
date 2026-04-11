package security

import (
	"testing"
)

// BenchmarkScanner_ScanSmall benchmarks scanning small content
func BenchmarkScanner_ScanSmall(b *testing.B) {
	scanner := NewScanner()
	content := "API Key: AKIAIOSFODNN7EXAMPLE"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		scanner.Scan(content)
	}
}

// BenchmarkScanner_ScanMedium benchmarks scanning medium content
func BenchmarkScanner_ScanMedium(b *testing.B) {
	scanner := NewScanner()
	content := "Normal text with AKIAIOSFODNN7EXAMPLE and test@example.com and some other content here"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		scanner.Scan(content)
	}
}

// BenchmarkScanner_ScanLarge benchmarks scanning large content
func BenchmarkScanner_ScanLarge(b *testing.B) {
	scanner := NewScanner()
	// 10KB of text with sensitive data
	base := "Normal text content here. "
	content := ""
	for i := 0; i < 400; i++ {
		content += base
	}
	content += "AKIAIOSFODNN7EXAMPLE test@example.com"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		scanner.Scan(content)
	}
}

// BenchmarkRedactPII benchmarks PII redaction
func BenchmarkRedactPII_Small(b *testing.B) {
	content := "Key: AKIAIOSFODNN7EXAMPLE, token: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		RedactPII(content)
	}
}

func BenchmarkRedactPII_Large(b *testing.B) {
	// 10KB of text
	base := "Normal text content here with email contact@example.com and key AKIAIOSFODNN7EXAMPLE. "
	content := ""
	for i := 0; i < 200; i++ {
		content += base
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		RedactPII(content)
	}
}

// BenchmarkIsSuspiciousContent benchmarks suspicious content detection
func BenchmarkIsSuspiciousContent(b *testing.B) {
	content := "Normal text without any suspicious patterns"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		IsSuspiciousContent(content)
	}
}

// BenchmarkValidateContent benchmarks content validation
func BenchmarkValidateContent(b *testing.B) {
	content := "Safe content without any sensitive information"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ValidateContent(content)
	}
}

// Benchmark parallel scanning
func BenchmarkScanner_ScanParallel(b *testing.B) {
	scanner := NewScanner()
	content := "API Key: AKIAIOSFODNN7EXAMPLE and email: test@example.com"

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			scanner.Scan(content)
		}
	})
}
