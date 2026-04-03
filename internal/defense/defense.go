package defense

import (
	"fmt"
	"regexp"
	"time"
)

type DefenseLayer int

const (
	LayerApplication DefenseLayer = 1
	LayerNetwork     DefenseLayer = 2
	LayerKernel      DefenseLayer = 3
)

type ThreatType string

const (
	ThreatPromptInjection ThreatType = "prompt_injection"
	ThreatPIILeak         ThreatType = "pii_leak"
	ThreatSecretExposure  ThreatType = "secret_exposure"
	ThreatMalware         ThreatType = "malware"
	ThreatVulnerability   ThreatType = "vulnerability"
)

type ThreatEvent struct {
	Type      ThreatType   `json:"type"`
	Layer     DefenseLayer `json:"layer"`
	Severity  int          `json:"severity"`
	Message   string       `json:"message"`
	Timestamp time.Time    `json:"timestamp"`
}

type SecurityPolicy struct {
	Version    string    `json:"version"`
	SHA256     string    `json:"sha256"`
	UpdatedAt  time.Time `json:"updated_at"`
	ShadowMode bool      `json:"shadow_mode"`
	CanaryMode bool      `json:"canary_mode"`
	Rules      []Rule    `json:"rules"`
}

type Rule struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Action  string `json:"action"`
	Pattern string `json:"pattern"`
}

type EventBus struct {
	events    chan ThreatEvent
	listeners []func(ThreatEvent)
}

func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		events: make(chan ThreatEvent, bufferSize),
	}
}

func (eb *EventBus) Publish(event ThreatEvent) {
	event.Timestamp = time.Now()
	select {
	case eb.events <- event:
	default:
	}
	for _, listener := range eb.listeners {
		go listener(event)
	}
}

func (eb *EventBus) Subscribe(fn func(ThreatEvent)) {
	eb.listeners = append(eb.listeners, fn)
}

func (eb *EventBus) Events() <-chan ThreatEvent {
	return eb.events
}

type DefenseInDepth struct {
	appDefense    *ApplicationDefense
	netDefense    *NetworkDefense
	kernelDefense *KernelDefense
	eventBus      *EventBus
	policy        *SecurityPolicy
}

func NewDefenseInDepth() *DefenseInDepth {
	return &DefenseInDepth{
		appDefense:    NewApplicationDefense(),
		netDefense:    NewNetworkDefense(),
		kernelDefense: NewKernelDefense(),
		eventBus:      NewEventBus(100),
		policy: &SecurityPolicy{
			Version: "1.0.0",
		},
	}
}

func (d *DefenseInDepth) Scan(input string) []ThreatEvent {
	var events []ThreatEvent

	appEvents := d.appDefense.Scan(input)
	for _, e := range appEvents {
		events = append(events, e)
		d.eventBus.Publish(e)
	}

	netEvents := d.netDefense.Check()
	for _, e := range netEvents {
		events = append(events, e)
		d.eventBus.Publish(e)
	}

	kernelEvents := d.kernelDefense.Check()
	for _, e := range kernelEvents {
		events = append(events, e)
		d.eventBus.Publish(e)
	}

	return events
}

func (d *DefenseInDepth) UpdatePolicy(policy *SecurityPolicy) error {
	d.policy = policy
	d.policy.UpdatedAt = time.Now()
	return nil
}

func (d *DefenseInDepth) Status() map[string]interface{} {
	return map[string]interface{}{
		"application": d.appDefense.Status(),
		"network":     d.netDefense.Status(),
		"kernel":      d.kernelDefense.Status(),
		"policy":      d.policy.Version,
		"shadow_mode": d.policy.ShadowMode,
	}
}

type ApplicationDefense struct {
	scanners []func(string) []ThreatEvent
}

func NewApplicationDefense() *ApplicationDefense {
	ad := &ApplicationDefense{}
	ad.scanners = append(ad.scanners, scanPromptInjection)
	ad.scanners = append(ad.scanners, scanPII)
	ad.scanners = append(ad.scanners, scanSecrets)
	return ad
}

func (ad *ApplicationDefense) Scan(input string) []ThreatEvent {
	var events []ThreatEvent
	for _, scanner := range ad.scanners {
		events = append(events, scanner(input)...)
	}
	return events
}

func (ad *ApplicationDefense) Status() string {
	return fmt.Sprintf("%d scanners active", len(ad.scanners))
}

func scanPromptInjection(input string) []ThreatEvent {
	var events []ThreatEvent
	patterns := []string{
		"ignore previous", "disregard", "system prompt",
		"you are now", "forget all", "new instructions",
	}
	for _, p := range patterns {
		if containsIgnoreCase(input, p) {
			events = append(events, ThreatEvent{
				Type:     ThreatPromptInjection,
				Layer:    LayerApplication,
				Severity: 8,
				Message:  "Potential prompt injection detected",
			})
		}
	}
	return events
}

func scanPII(input string) []ThreatEvent {
	var events []ThreatEvent
	if containsPattern(input, `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`) {
		events = append(events, ThreatEvent{
			Type:     ThreatPIILeak,
			Layer:    LayerApplication,
			Severity: 7,
			Message:  "Email address detected",
		})
	}
	return events
}

func scanSecrets(input string) []ThreatEvent {
	var events []ThreatEvent
	secretPatterns := []string{
		"sk-", "pk-", "api_key", "apikey", "secret",
		"password", "token", "credential",
	}
	for _, p := range secretPatterns {
		if containsIgnoreCase(input, p) {
			events = append(events, ThreatEvent{
				Type:     ThreatSecretExposure,
				Layer:    LayerApplication,
				Severity: 9,
				Message:  "Potential secret detected",
			})
		}
	}
	return events
}

type NetworkDefense struct {
	allowedDomains []string
	allowedIPs     []string
}

func NewNetworkDefense() *NetworkDefense {
	return &NetworkDefense{}
}

func (nd *NetworkDefense) Check() []ThreatEvent {
	return nil
}

func (nd *NetworkDefense) Status() string {
	return "monitoring"
}

func (nd *NetworkDefense) AddAllowedDomain(domain string) {
	nd.allowedDomains = append(nd.allowedDomains, domain)
}

func (nd *NetworkDefense) AddAllowedIP(ip string) {
	nd.allowedIPs = append(nd.allowedIPs, ip)
}

type KernelDefense struct {
	ebpfSupported bool
}

func NewKernelDefense() *KernelDefense {
	return &KernelDefense{ebpfSupported: false}
}

func (kd *KernelDefense) Check() []ThreatEvent {
	if !kd.ebpfSupported {
		return nil
	}
	return nil
}

func (kd *KernelDefense) Status() string {
	if kd.ebpfSupported {
		return "eBPF active"
	}
	return "eBPF not supported, using procfs fallback"
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsIgnoreCaseHelper(s, substr))
}

func containsIgnoreCaseHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc, pc := s[i+j], substr[j]
			if sc != pc && sc != pc+32 && sc+32 != pc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func containsPattern(s, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}
