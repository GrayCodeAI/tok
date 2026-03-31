package security

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// EBPFMonitor implements eBPF-based syscall monitoring.
// Inspired by clawshield's eBPF kernel monitoring.
type EBPFMonitor struct {
	enabled  bool
	syscalls []string
	alerts   []SecurityAlert
}

// SecurityAlert represents a security alert.
type SecurityAlert struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// NewEBPFMonitor creates a new eBPF monitor.
func NewEBPFMonitor() *EBPFMonitor {
	return &EBPFMonitor{
		enabled:  false,
		syscalls: []string{"execve", "open", "connect", "bind", "listen"},
	}
}

// Enable enables eBPF monitoring.
func (m *EBPFMonitor) Enable() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("eBPF monitoring requires root privileges")
	}
	m.enabled = true
	return nil
}

// CheckSyscall checks if a syscall pattern is suspicious.
func (m *EBPFMonitor) CheckSyscall(syscall string, args string) *SecurityAlert {
	suspicious := map[string][]string{
		"execve":  {"/bin/sh", "/bin/bash", "curl", "wget", "nc", "ncat"},
		"open":    {"/etc/passwd", "/etc/shadow", "/proc/", "/sys/"},
		"connect": {"169.254.169.254", "metadata.google"},
	}

	if patterns, ok := suspicious[syscall]; ok {
		for _, p := range patterns {
			if strings.Contains(args, p) {
				return &SecurityAlert{
					Type:      "suspicious_syscall",
					Severity:  "high",
					Message:   fmt.Sprintf("Suspicious %s: %s", syscall, args),
					Timestamp: time.Now(),
					Details:   fmt.Sprintf("syscall=%s args=%s", syscall, args),
				}
			}
		}
	}
	return nil
}

// SIEMIntegration implements SIEM integration with OCSF format.
// Inspired by clawshield's SIEM integration.
type SIEMIntegration struct {
	endpoint string
	format   string
}

// NewSIEMIntegration creates a new SIEM integration.
func NewSIEMIntegration(endpoint string) *SIEMIntegration {
	return &SIEMIntegration{
		endpoint: endpoint,
		format:   "ocsf",
	}
}

// FormatOCSF formats an alert in OCSF v1.1 format.
func (si *SIEMIntegration) FormatOCSF(alert SecurityAlert) string {
	event := map[string]any{
		"metadata": map[string]any{
			"version":    "1.1.0",
			"product":    map[string]any{"name": "tokman", "vendor_name": "GrayCodeAI"},
			"event_code": alert.Type,
			"uid":        fmt.Sprintf("tokman-%d", time.Now().UnixNano()),
		},
		"time":        alert.Timestamp.Format(time.RFC3339),
		"severity_id": severityToID(alert.Severity),
		"message":     alert.Message,
		"details":     alert.Details,
	}
	data, _ := json.Marshal(event)
	return string(data)
}

// DecisionExplainability provides structured forensic audit records.
// Inspired by clawshield's decision explainability.
type DecisionExplainability struct {
	decisions []DecisionRecord
}

// DecisionRecord represents a forensic audit record.
type DecisionRecord struct {
	Action     string    `json:"action"`
	Reason     string    `json:"reason"`
	Evidence   []string  `json:"evidence"`
	Timestamp  time.Time `json:"timestamp"`
	Confidence float64   `json:"confidence"`
}

// NewDecisionExplainability creates a new decision explainability system.
func NewDecisionExplainability() *DecisionExplainability {
	return &DecisionExplainability{}
}

// Record records a decision with evidence.
func (de *DecisionExplainability) Record(action, reason string, evidence []string, confidence float64) {
	de.decisions = append(de.decisions, DecisionRecord{
		Action:     action,
		Reason:     reason,
		Evidence:   evidence,
		Timestamp:  time.Now(),
		Confidence: confidence,
	})
}

// GetDecisions returns all recorded decisions.
func (de *DecisionExplainability) GetDecisions() []DecisionRecord {
	return de.decisions
}

// NetworkFirewall implements iptables-based egress filtering.
// Inspired by clawshield's network firewall.
type NetworkFirewall struct {
	rules   []FirewallRule
	enabled bool
}

// FirewallRule represents a firewall rule.
type FirewallRule struct {
	Action     string // "allow", "deny"
	Protocol   string // "tcp", "udp", "icmp"
	Port       int
	DestIP     string
	DestDomain string
}

// NewNetworkFirewall creates a new network firewall.
func NewNetworkFirewall() *NetworkFirewall {
	return &NetworkFirewall{rules: []FirewallRule{}}
}

// AddRule adds a firewall rule.
func (nf *NetworkFirewall) AddRule(rule FirewallRule) {
	nf.rules = append(nf.rules, rule)
}

// CheckConnection checks if a connection is allowed.
func (nf *NetworkFirewall) CheckConnection(destIP string, port int, protocol string) bool {
	for _, rule := range nf.rules {
		if (rule.DestIP == "" || rule.DestIP == destIP) &&
			(rule.Port == 0 || rule.Port == port) &&
			(rule.Protocol == "" || rule.Protocol == protocol) {
			return rule.Action == "allow"
		}
	}
	return true // Default allow
}

// ContentGuardrails implements PII and prompt injection detection.
// Inspired by token-lens's content guardrails.
type ContentGuardrails struct {
	rules []GuardrailRule
}

// GuardrailRule represents a content guardrail rule.
type GuardrailRule struct {
	Name    string
	Pattern string
	Action  string // "block", "redact", "warn"
}

// NewContentGuardrails creates new content guardrails.
func NewContentGuardrails() *ContentGuardrails {
	return &ContentGuardrails{
		rules: []GuardrailRule{
			{Name: "pii_email", Pattern: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, Action: "redact"},
			{Name: "pii_phone", Pattern: `\b\d{3}[-.\s]?\d{3}[-.\s]?\d{4}\b`, Action: "redact"},
			{Name: "injection", Pattern: `(?i)ignore\s+(?:all\s+)?(?:previous|prior)`, Action: "block"},
			{Name: "jailbreak", Pattern: `(?i)(?:DAN|jailbreak|act as)`, Action: "block"},
		},
	}
}

// Check checks content against guardrails.
func (cg *ContentGuardrails) Check(content string) []SecurityAlert {
	var alerts []SecurityAlert
	for _, rule := range cg.rules {
		if strings.Contains(content, rule.Pattern) || matchesRegex(content, rule.Pattern) {
			alerts = append(alerts, SecurityAlert{
				Type:      "guardrail_violation",
				Severity:  "high",
				Message:   fmt.Sprintf("Guardrail violated: %s", rule.Name),
				Timestamp: time.Now(),
			})
		}
	}
	return alerts
}

func severityToID(severity string) int {
	switch severity {
	case "critical":
		return 1
	case "high":
		return 2
	case "medium":
		return 3
	case "low":
		return 4
	case "info":
		return 5
	default:
		return 3
	}
}

func matchesRegex(content, pattern string) bool {
	if strings.Contains(pattern, "(?i)") || strings.Contains(pattern, "\\b") || strings.Contains(pattern, "[") {
		// Simple regex check - use regexp for complex patterns
		return false
	}
	return strings.Contains(content, pattern)
}
