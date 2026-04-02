package scanning

import (
	"strings"
	"sync"
	"time"
)

type ScanResult struct {
	Allowed    bool      `json:"allowed"`
	Findings   []string  `json:"findings,omitempty"`
	Severity   int       `json:"severity"`
	Redacted   string    `json:"redacted,omitempty"`
	ScannedAt  time.Time `json:"scanned_at"`
	ScanTimeNs int64     `json:"scan_time_ns"`
}

type StreamingScanner struct {
	overlapSize int
	buffer      string
	mu          sync.Mutex
	findings    []string
}

func NewStreamingScanner(overlapSize int) *StreamingScanner {
	if overlapSize == 0 {
		overlapSize = 200
	}
	return &StreamingScanner{overlapSize: overlapSize}
}

func (s *StreamingScanner) ScanChunk(chunk string) *ScanResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	start := time.Now()
	combined := s.buffer + chunk
	s.buffer = chunk[max(0, len(chunk)-s.overlapSize):]

	var findings []string
	if strings.Contains(combined, "api_key") {
		findings = append(findings, "secret_exposure")
	}
	if strings.Contains(strings.ToLower(combined), "ignore previous") {
		findings = append(findings, "prompt_injection")
	}
	if strings.Contains(combined, "@") && strings.Contains(combined, ".com") {
		findings = append(findings, "pii_detected")
	}

	s.findings = append(s.findings, findings...)

	return &ScanResult{
		Allowed:    len(findings) == 0,
		Findings:   findings,
		Severity:   len(findings) * 3,
		ScannedAt:  time.Now(),
		ScanTimeNs: time.Since(start).Nanoseconds(),
	}
}

func (s *StreamingScanner) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffer = ""
	s.findings = nil
}

func (s *StreamingScanner) GetFindings() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string{}, s.findings...)
}

type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Scanner   string    `json:"scanner"`
	RuleID    string    `json:"rule_id"`
	Allowed   bool      `json:"allowed"`
	Details   string    `json:"details"`
	Forensic  string    `json:"forensic,omitempty"`
}

type AuditLog struct {
	entries []AuditEntry
	mu      sync.Mutex
}

func NewAuditLog() *AuditLog {
	return &AuditLog{}
}

func (a *AuditLog) Record(entry AuditEntry) {
	entry.Timestamp = time.Now()
	a.mu.Lock()
	a.entries = append(a.entries, entry)
	a.mu.Unlock()
}

func (a *AuditLog) GetLast(n int) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	if n > len(a.entries) {
		n = len(a.entries)
	}
	return a.entries[len(a.entries)-n:]
}

func (a *AuditLog) GetBlocked() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	var blocked []AuditEntry
	for _, e := range a.entries {
		if !e.Allowed {
			blocked = append(blocked, e)
		}
	}
	return blocked
}

func (a *AuditLog) GetByScanner(scanner string) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	var result []AuditEntry
	for _, e := range a.entries {
		if e.Scanner == scanner {
			result = append(result, e)
		}
	}
	return result
}

func (a *AuditLog) GetByRule(ruleID string) []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	var result []AuditEntry
	for _, e := range a.entries {
		if e.RuleID == ruleID {
			result = append(result, e)
		}
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
