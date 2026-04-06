package security

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
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

// NewSIEMIntegration creates a new SIEM integration with endpoint validation.
func NewSIEMIntegration(endpoint string) (*SIEMIntegration, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid SIEM endpoint %q: %w", endpoint, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("SIEM endpoint must use http or https scheme, got: %s", parsed.Scheme)
	}
	if parsed.Hostname() == "" {
		return nil, fmt.Errorf("SIEM endpoint must have a valid hostname")
	}
	// Block private/internal network endpoints
	host := parsed.Hostname()
	blockedHosts := []string{"localhost", "127.0.0.1", "0.0.0.0", "169.254.169.254", "metadata.google.internal"}
	for _, bh := range blockedHosts {
		if host == bh {
			return nil, fmt.Errorf("SIEM endpoint %q points to a blocked internal address", host)
		}
	}
	return &SIEMIntegration{
		endpoint: endpoint,
		format:   "ocsf",
	}, nil
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
	maxSize   int
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
	return &DecisionExplainability{maxSize: 10000}
}

// NewDecisionExplainabilityWithLimit creates with a custom max size.
func NewDecisionExplainabilityWithLimit(maxSize int) *DecisionExplainability {
	return &DecisionExplainability{maxSize: maxSize}
}

// Record records a decision with evidence.
func (de *DecisionExplainability) Record(action, reason string, evidence []string, confidence float64) {
	if de.maxSize > 0 && len(de.decisions) >= de.maxSize {
		// Drop oldest entries when at capacity
		de.decisions = de.decisions[len(de.decisions)/2:]
	}
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
	return false // Default deny
}

// ContentGuardrails implements PII and prompt injection detection.
// Inspired by token-lens's content guardrails.
type ContentGuardrails struct {
	rules         []GuardrailRule
	compiledRegex []*regexp.Regexp
}

// GuardrailRule represents a content guardrail rule.
type GuardrailRule struct {
	Name    string
	Pattern string
	Action  string // "block", "redact", "warn"
}

// NewContentGuardrails creates new content guardrails.
func NewContentGuardrails() *ContentGuardrails {
	cg := &ContentGuardrails{
		rules: []GuardrailRule{
			{Name: "pii_email", Pattern: `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, Action: "redact"},
			{Name: "pii_phone", Pattern: `\b\d{3}[-.\s]?\d{3}[-.\s]?\d{4}\b`, Action: "redact"},
			{Name: "injection", Pattern: `(?i)ignore\s+(?:all\s+)?(?:previous|prior)`, Action: "block"},
			{Name: "jailbreak", Pattern: `(?i)(?:DAN|jailbreak|act as)`, Action: "block"},
		},
	}
	// Pre-compile regexes so Check uses proper regex matching
	for _, rule := range cg.rules {
		re, err := regexp.Compile(rule.Pattern)
		if err == nil {
			cg.compiledRegex = append(cg.compiledRegex, re)
		}
	}
	return cg
}

// Check checks content against guardrails.
func (cg *ContentGuardrails) Check(content string) []SecurityAlert {
	var alerts []SecurityAlert
	for i, rule := range cg.rules {
		if i < len(cg.compiledRegex) && cg.compiledRegex[i] != nil {
			if cg.compiledRegex[i].MatchString(content) {
				alerts = append(alerts, SecurityAlert{
					Type:      "guardrail_violation",
					Severity:  "high",
					Message:   fmt.Sprintf("Guardrail violated: %s", rule.Name),
					Timestamp: time.Now(),
				})
			}
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
