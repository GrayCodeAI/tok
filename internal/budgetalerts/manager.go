// Package budgetalerts provides budget alerting capabilities for TokMan
package budgetalerts

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager manages budget alerts
type Manager struct {
	alerts    map[string]*Alert
	rules     map[string]*Rule
	channels  map[string]NotificationChannel
	history   []AlertEvent
	mu        sync.RWMutex
	evaluator *RuleEvaluator
}

// Alert represents a budget alert instance
type Alert struct {
	ID          string
	RuleID      string
	Name        string
	Description string
	Status      AlertStatus
	Severity    Severity
	TriggeredAt time.Time
	ResolvedAt  *time.Time
	Value       float64
	Threshold   float64
	Message     string
	Metadata    map[string]string
}

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusMuted        AlertStatus = "muted"
)

// Severity represents alert severity
type Severity string

const (
	SeverityInfo      Severity = "info"
	SeverityWarning   Severity = "warning"
	SeverityCritical  Severity = "critical"
	SeverityEmergency Severity = "emergency"
)

// Rule defines an alert rule
type Rule struct {
	ID            string
	Name          string
	Description   string
	Enabled       bool
	Condition     Condition
	Thresholds    Thresholds
	Notifications NotificationConfig
	Cooldown      time.Duration
	AutoResolve   bool
	lastTriggered time.Time
	triggerCount  int
}

// Condition defines when to trigger an alert
type Condition struct {
	Metric      string
	Operator    Operator
	Value       float64
	Duration    time.Duration
	Aggregation AggregationType
}

// Operator defines comparison operators
type Operator string

const (
	OpGreaterThan      Operator = ">"
	OpGreaterThanEqual Operator = ">="
	OpLessThan         Operator = "<"
	OpLessThanEqual    Operator = "<="
	OpEqual            Operator = "=="
	OpNotEqual         Operator = "!="
)

// AggregationType defines how to aggregate metrics
type AggregationType string

const (
	AggAvg   AggregationType = "avg"
	AggSum   AggregationType = "sum"
	AggMin   AggregationType = "min"
	AggMax   AggregationType = "max"
	AggCount AggregationType = "count"
	AggLast  AggregationType = "last"
)

// Thresholds defines multi-level thresholds
type Thresholds struct {
	Warning   float64
	Critical  float64
	Emergency float64
}

// NotificationConfig defines how to send notifications
type NotificationConfig struct {
	Channels []string
	Throttle time.Duration
	Template string
}

// NotificationChannel defines a notification channel
type NotificationChannel struct {
	ID     string
	Name   string
	Type   ChannelType
	Config map[string]string
}

// ChannelType defines notification channel types
type ChannelType string

const (
	ChannelEmail     ChannelType = "email"
	ChannelSlack     ChannelType = "slack"
	ChannelWebhook   ChannelType = "webhook"
	ChannelPagerDuty ChannelType = "pagerduty"
	ChannelConsole   ChannelType = "console"
)

// AlertEvent represents an alert event in history
type AlertEvent struct {
	Timestamp time.Time
	AlertID   string
	RuleID    string
	Type      EventType
	Message   string
}

// EventType represents the type of event
type EventType string

const (
	EventTriggered    EventType = "triggered"
	EventResolved     EventType = "resolved"
	EventAcknowledged EventType = "acknowledged"
	EventEscalated    EventType = "escalated"
)

// RuleEvaluator evaluates alert rules
type RuleEvaluator struct {
	metricProvider MetricProvider
}

// MetricProvider provides metric values
type MetricProvider interface {
	GetMetric(name string, duration time.Duration, agg AggregationType) (float64, error)
	GetCurrentValue(name string) (float64, error)
}

// NewManager creates a budget alert manager
func NewManager() *Manager {
	return &Manager{
		alerts:    make(map[string]*Alert),
		rules:     make(map[string]*Rule),
		channels:  make(map[string]NotificationChannel),
		history:   make([]AlertEvent, 0),
		evaluator: &RuleEvaluator{},
	}
}

// SetMetricProvider sets the metric provider for evaluation
func (m *Manager) SetMetricProvider(provider MetricProvider) {
	m.evaluator.metricProvider = provider
}

// CreateRule creates a new alert rule
func (m *Manager) CreateRule(config RuleConfig) (*Rule, error) {
	rule := &Rule{
		ID:            generateID(),
		Name:          config.Name,
		Description:   config.Description,
		Enabled:       config.Enabled,
		Condition:     config.Condition,
		Thresholds:    config.Thresholds,
		Notifications: config.Notifications,
		Cooldown:      config.Cooldown,
		AutoResolve:   config.AutoResolve,
	}

	m.mu.Lock()
	m.rules[rule.ID] = rule
	m.mu.Unlock()

	return rule, nil
}

// GetRule returns a rule by ID
func (m *Manager) GetRule(id string) (*Rule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rule, exists := m.rules[id]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", id)
	}

	return rule, nil
}

// UpdateRule updates an existing rule
func (m *Manager) UpdateRule(id string, config RuleConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rule, exists := m.rules[id]
	if !exists {
		return fmt.Errorf("rule %s not found", id)
	}

	rule.Name = config.Name
	rule.Description = config.Description
	rule.Enabled = config.Enabled
	rule.Condition = config.Condition
	rule.Thresholds = config.Thresholds
	rule.Notifications = config.Notifications
	rule.Cooldown = config.Cooldown
	rule.AutoResolve = config.AutoResolve

	return nil
}

// DeleteRule deletes a rule
func (m *Manager) DeleteRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rules[id]; !exists {
		return fmt.Errorf("rule %s not found", id)
	}

	delete(m.rules, id)
	return nil
}

// ListRules returns all rules
func (m *Manager) ListRules() []*Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Rule, 0, len(m.rules))
	for _, rule := range m.rules {
		result = append(result, rule)
	}

	return result
}

// RegisterChannel registers a notification channel
func (m *Manager) RegisterChannel(channel NotificationChannel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[channel.ID] = channel
}

// EvaluateRules evaluates all enabled rules
func (m *Manager) EvaluateRules(ctx context.Context) error {
	m.mu.RLock()
	rules := make([]*Rule, 0, len(m.rules))
	for _, rule := range m.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	m.mu.RUnlock()

	for _, rule := range rules {
		if err := m.evaluateRule(ctx, rule); err != nil {
			// Log error but continue evaluating other rules
			continue
		}
	}

	return nil
}

func (m *Manager) evaluateRule(ctx context.Context, rule *Rule) error {
	// Check cooldown
	if time.Since(rule.lastTriggered) < rule.Cooldown {
		return nil
	}

	// Get metric value
	value, err := m.evaluator.metricProvider.GetMetric(
		rule.Condition.Metric,
		rule.Condition.Duration,
		rule.Condition.Aggregation,
	)
	if err != nil {
		return err
	}

	// Evaluate condition
	triggered := m.compare(value, rule.Condition.Operator, rule.Condition.Value)

	if triggered {
		// Determine severity
		severity := m.determineSeverity(value, rule.Thresholds)

		// Create or update alert
		alert := m.createOrUpdateAlert(rule, value, severity)

		// Send notifications
		m.sendNotifications(ctx, alert, rule)

		rule.lastTriggered = time.Now()
		rule.triggerCount++
	} else if rule.AutoResolve {
		// Check for auto-resolve
		m.checkAutoResolve(rule)
	}

	return nil
}

func (m *Manager) compare(value float64, op Operator, threshold float64) bool {
	switch op {
	case OpGreaterThan:
		return value > threshold
	case OpGreaterThanEqual:
		return value >= threshold
	case OpLessThan:
		return value < threshold
	case OpLessThanEqual:
		return value <= threshold
	case OpEqual:
		return value == threshold
	case OpNotEqual:
		return value != threshold
	default:
		return false
	}
}

func (m *Manager) determineSeverity(value float64, thresholds Thresholds) Severity {
	if value >= thresholds.Emergency {
		return SeverityEmergency
	}
	if value >= thresholds.Critical {
		return SeverityCritical
	}
	if value >= thresholds.Warning {
		return SeverityWarning
	}
	return SeverityInfo
}

func (m *Manager) createOrUpdateAlert(rule *Rule, value float64, severity Severity) *Alert {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if alert already exists
	for _, alert := range m.alerts {
		if alert.RuleID == rule.ID && alert.Status == AlertStatusActive {
			// Update existing alert
			alert.Value = value
			alert.Severity = severity
			return alert
		}
	}

	// Create new alert
	alert := &Alert{
		ID:          generateID(),
		RuleID:      rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Status:      AlertStatusActive,
		Severity:    severity,
		TriggeredAt: time.Now(),
		Value:       value,
		Threshold:   rule.Condition.Value,
		Message:     fmt.Sprintf("%s: %.2f exceeds threshold %.2f", rule.Name, value, rule.Condition.Value),
	}

	m.alerts[alert.ID] = alert

	// Add to history
	m.history = append(m.history, AlertEvent{
		Timestamp: time.Now(),
		AlertID:   alert.ID,
		RuleID:    rule.ID,
		Type:      EventTriggered,
		Message:   alert.Message,
	})

	return alert
}

func (m *Manager) sendNotifications(ctx context.Context, alert *Alert, rule *Rule) {
	for _, channelID := range rule.Notifications.Channels {
		channel, exists := m.channels[channelID]
		if !exists {
			continue
		}

		// Send notification based on channel type
		switch channel.Type {
		case ChannelConsole:
			fmt.Printf("[ALERT %s] %s: %s\n", alert.Severity, alert.Name, alert.Message)
		case ChannelWebhook:
			// Send webhook notification
		case ChannelSlack:
			// Send Slack notification
		case ChannelEmail:
			// Send email notification
		}
	}
}

func (m *Manager) checkAutoResolve(rule *Rule) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, alert := range m.alerts {
		if alert.RuleID == rule.ID && alert.Status == AlertStatusActive {
			alert.Status = AlertStatusResolved
			now := time.Now()
			alert.ResolvedAt = &now

			m.history = append(m.history, AlertEvent{
				Timestamp: time.Now(),
				AlertID:   alert.ID,
				RuleID:    rule.ID,
				Type:      EventResolved,
				Message:   fmt.Sprintf("Alert auto-resolved: %s", alert.Name),
			})
		}
	}
}

// GetActiveAlerts returns all active alerts
func (m *Manager) GetActiveAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Alert, 0)
	for _, alert := range m.alerts {
		if alert.Status == AlertStatusActive {
			result = append(result, alert)
		}
	}

	return result
}

// AcknowledgeAlert acknowledges an alert
func (m *Manager) AcknowledgeAlert(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, exists := m.alerts[id]
	if !exists {
		return fmt.Errorf("alert %s not found", id)
	}

	alert.Status = AlertStatusAcknowledged

	m.history = append(m.history, AlertEvent{
		Timestamp: time.Now(),
		AlertID:   alert.ID,
		RuleID:    alert.RuleID,
		Type:      EventAcknowledged,
		Message:   fmt.Sprintf("Alert acknowledged: %s", alert.Name),
	})

	return nil
}

// ResolveAlert manually resolves an alert
func (m *Manager) ResolveAlert(id string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	alert, exists := m.alerts[id]
	if !exists {
		return fmt.Errorf("alert %s not found", id)
	}

	alert.Status = AlertStatusResolved
	now := time.Now()
	alert.ResolvedAt = &now

	m.history = append(m.history, AlertEvent{
		Timestamp: time.Now(),
		AlertID:   alert.ID,
		RuleID:    alert.RuleID,
		Type:      EventResolved,
		Message:   reason,
	})

	return nil
}

// GetHistory returns alert history
func (m *Manager) GetHistory(limit int) []AlertEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.history) {
		limit = len(m.history)
	}

	start := len(m.history) - limit
	if start < 0 {
		start = 0
	}

	return m.history[start:]
}

// RuleConfig holds rule configuration
type RuleConfig struct {
	Name          string
	Description   string
	Enabled       bool
	Condition     Condition
	Thresholds    Thresholds
	Notifications NotificationConfig
	Cooldown      time.Duration
	AutoResolve   bool
}

func generateID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}
