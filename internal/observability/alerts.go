package observability

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// AlertLevel represents the severity of an alert.
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert represents an alert event.
type Alert struct {
	ID          string
	Level       AlertLevel
	Title       string
	Description string
	Metric      string
	Value       interface{}
	Threshold   interface{}
	Timestamp   time.Time
	Context     map[string]interface{}
}

// AlertRule defines when an alert should trigger.
type AlertRule struct {
	ID          string
	Name        string
	Description string
	Metric      string
	Condition   AlertCondition
	Duration    time.Duration
	Severity    AlertLevel
	Enabled     bool
	Actions     []AlertAction
}

// AlertCondition represents a condition that triggers an alert.
type AlertCondition interface {
	Check(value interface{}) bool
	GetThreshold() interface{}
}

// GreaterThan checks if value > threshold.
type GreaterThan struct {
	Threshold interface{}
}

func (gt *GreaterThan) Check(value interface{}) bool {
	switch v := value.(type) {
	case float64:
		if t, ok := gt.Threshold.(float64); ok {
			return v > t
		}
	case int:
		if t, ok := gt.Threshold.(int); ok {
			return v > t
		}
	}
	return false
}

func (gt *GreaterThan) GetThreshold() interface{} {
	return gt.Threshold
}

// LessThan checks if value < threshold.
type LessThan struct {
	Threshold interface{}
}

func (lt *LessThan) Check(value interface{}) bool {
	switch v := value.(type) {
	case float64:
		if t, ok := lt.Threshold.(float64); ok {
			return v < t
		}
	case int:
		if t, ok := lt.Threshold.(int); ok {
			return v < t
		}
	}
	return false
}

func (lt *LessThan) GetThreshold() interface{} {
	return lt.Threshold
}

// AlertAction represents an action to take when alert triggers.
type AlertAction interface {
	Execute(ctx context.Context, alert Alert) error
}

// EmailAction sends an email alert.
type EmailAction struct {
	To      []string
	Subject string
}

func (ea *EmailAction) Execute(ctx context.Context, alert Alert) error {
	// TODO: Implement email sending
	return nil
}

// SlackAction sends a Slack message.
type SlackAction struct {
	WebhookURL string
	Channel    string
}

func (sa *SlackAction) Execute(ctx context.Context, alert Alert) error {
	// TODO: Implement Slack webhook
	return nil
}

// AlertManager manages alert rules and events.
type AlertManager struct {
	mu             sync.RWMutex
	rules          map[string]*AlertRule
	activeAlerts   map[string]*Alert
	alertHistory   []*Alert
	maxHistorySize int
	logger         *slog.Logger
	alertChannels  map[AlertLevel]chan Alert
}

// NewAlertManager creates a new alert manager.
func NewAlertManager(logger *slog.Logger) *AlertManager {
	if logger == nil {
		logger = slog.Default()
	}

	return &AlertManager{
		rules:          make(map[string]*AlertRule),
		activeAlerts:   make(map[string]*Alert),
		alertHistory:   make([]*Alert, 0),
		maxHistorySize: 1000,
		logger:         logger,
		alertChannels: map[AlertLevel]chan Alert{
			AlertLevelInfo:     make(chan Alert, 10),
			AlertLevelWarning:  make(chan Alert, 10),
			AlertLevelCritical: make(chan Alert, 10),
		},
	}
}

// RegisterRule registers a new alert rule.
func (am *AlertManager) RegisterRule(rule *AlertRule) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.rules[rule.ID]; exists {
		return fmt.Errorf("rule already exists: %s", rule.ID)
	}

	am.rules[rule.ID] = rule
	am.logger.Info("alert rule registered",
		slog.String("rule_id", rule.ID),
		slog.String("name", rule.Name),
		slog.String("metric", rule.Metric),
	)

	return nil
}

// RemoveRule removes an alert rule.
func (am *AlertManager) RemoveRule(ruleID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	delete(am.rules, ruleID)
	return nil
}

// CheckMetric evaluates a metric against all rules.
func (am *AlertManager) CheckMetric(metricName string, value interface{}, ctx map[string]interface{}) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, rule := range am.rules {
		if !rule.Enabled || rule.Metric != metricName {
			continue
		}

		if rule.Condition.Check(value) {
			alert := Alert{
				ID:          generateAlertID(),
				Level:       rule.Severity,
				Title:       rule.Name,
				Description: rule.Description,
				Metric:      metricName,
				Value:       value,
				Threshold:   rule.Condition.GetThreshold(),
				Timestamp:   time.Now(),
				Context:     ctx,
			}

			am.recordAlert(alert)

			// Execute actions
			for _, action := range rule.Actions {
				if err := action.Execute(context.Background(), alert); err != nil {
					am.logger.Error("failed to execute alert action",
						slog.String("alert_id", alert.ID),
						slog.String("error", err.Error()),
					)
				}
			}

			// Send to channel
			select {
			case am.alertChannels[rule.Severity] <- alert:
			default:
				am.logger.Warn("alert channel full", slog.String("alert_id", alert.ID))
			}
		}
	}

	return nil
}

// GetActiveAlerts returns currently active alerts.
func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetAlertHistory returns alert history.
func (am *AlertManager) GetAlertHistory(limit int) []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if limit > len(am.alertHistory) {
		limit = len(am.alertHistory)
	}

	return am.alertHistory[:limit]
}

// recordAlert records an alert.
func (am *AlertManager) recordAlert(alert Alert) {
	am.activeAlerts[alert.ID] = &alert
	am.alertHistory = append(am.alertHistory, &alert)

	// Trim history
	if len(am.alertHistory) > am.maxHistorySize {
		am.alertHistory = am.alertHistory[1:]
	}

	am.logger.Warn("alert triggered",
		slog.String("alert_id", alert.ID),
		slog.String("level", string(alert.Level)),
		slog.String("metric", alert.Metric),
		slog.Any("value", alert.Value),
	)
}

// ResolveAlert marks an alert as resolved.
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.activeAlerts[alertID]; !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	delete(am.activeAlerts, alertID)
	am.logger.Info("alert resolved", slog.String("alert_id", alertID))

	return nil
}

// GetAlertChannel returns a channel for alerts of a specific level.
func (am *AlertManager) GetAlertChannel(level AlertLevel) <-chan Alert {
	return am.alertChannels[level]
}

func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}
