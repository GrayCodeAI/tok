package budgetalerts

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("expected manager to be created")
	}

	if manager.alerts == nil {
		t.Error("expected alerts map to be initialized")
	}

	if manager.rules == nil {
		t.Error("expected rules map to be initialized")
	}
}

func TestCreateRule(t *testing.T) {
	manager := NewManager()

	config := RuleConfig{
		Name:        "High Cost Alert",
		Description: "Alert when daily cost exceeds $100",
		Enabled:     true,
		Condition: Condition{
			Metric:      "daily_cost",
			Operator:    OpGreaterThan,
			Value:       100,
			Duration:    time.Hour,
			Aggregation: AggSum,
		},
		Thresholds: Thresholds{
			Warning:   100,
			Critical:  200,
			Emergency: 500,
		},
		Cooldown:    time.Hour,
		AutoResolve: true,
	}

	rule, err := manager.CreateRule(config)
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	if rule == nil {
		t.Fatal("expected rule to be created")
	}

	if rule.Name != "High Cost Alert" {
		t.Errorf("expected name 'High Cost Alert', got %s", rule.Name)
	}

	if rule.ID == "" {
		t.Error("expected rule ID to be generated")
	}
}

func TestGetRule(t *testing.T) {
	manager := NewManager()

	config := RuleConfig{Name: "Test"}
	rule, _ := manager.CreateRule(config)

	found, err := manager.GetRule(rule.ID)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}

	if found.ID != rule.ID {
		t.Errorf("expected rule ID %s, got %s", rule.ID, found.ID)
	}
}

func TestGetRuleNotFound(t *testing.T) {
	manager := NewManager()

	_, err := manager.GetRule("non-existent")
	if err == nil {
		t.Error("expected error for non-existent rule")
	}
}

func TestUpdateRule(t *testing.T) {
	manager := NewManager()

	config := RuleConfig{Name: "Original"}
	rule, _ := manager.CreateRule(config)

	update := RuleConfig{Name: "Updated"}
	err := manager.UpdateRule(rule.ID, update)
	if err != nil {
		t.Fatalf("failed to update rule: %v", err)
	}

	updated, _ := manager.GetRule(rule.ID)
	if updated.Name != "Updated" {
		t.Errorf("expected name 'Updated', got %s", updated.Name)
	}
}

func TestDeleteRule(t *testing.T) {
	manager := NewManager()

	config := RuleConfig{Name: "ToDelete"}
	rule, _ := manager.CreateRule(config)

	err := manager.DeleteRule(rule.ID)
	if err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}

	_, err = manager.GetRule(rule.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestListRules(t *testing.T) {
	manager := NewManager()

	manager.CreateRule(RuleConfig{Name: "Rule1"})
	manager.CreateRule(RuleConfig{Name: "Rule2"})

	rules := manager.ListRules()

	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestRegisterChannel(t *testing.T) {
	manager := NewManager()

	channel := NotificationChannel{
		ID:   "email-1",
		Name: "Email Notifications",
		Type: ChannelEmail,
	}

	manager.RegisterChannel(channel)

	if _, ok := manager.channels["email-1"]; !ok {
		t.Error("expected channel to be registered")
	}
}

func TestCompareOperators(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		operator  Operator
		value     float64
		threshold float64
		expected  bool
	}{
		{OpGreaterThan, 150, 100, true},
		{OpGreaterThan, 50, 100, false},
		{OpGreaterThanEqual, 100, 100, true},
		{OpLessThan, 50, 100, true},
		{OpLessThan, 150, 100, false},
		{OpLessThanEqual, 100, 100, true},
		{OpEqual, 100, 100, true},
		{OpEqual, 50, 100, false},
		{OpNotEqual, 50, 100, true},
		{OpNotEqual, 100, 100, false},
	}

	for _, tt := range tests {
		result := manager.compare(tt.value, tt.operator, tt.threshold)
		if result != tt.expected {
			t.Errorf("%s(%.0f, %.0f): expected %v, got %v",
				tt.operator, tt.value, tt.threshold, tt.expected, result)
		}
	}
}

func TestDetermineSeverity(t *testing.T) {
	manager := NewManager()

	thresholds := Thresholds{
		Warning:   100,
		Critical:  200,
		Emergency: 500,
	}

	tests := []struct {
		value    float64
		expected Severity
	}{
		{50, SeverityInfo},
		{100, SeverityWarning},
		{150, SeverityWarning},
		{200, SeverityCritical},
		{300, SeverityCritical},
		{500, SeverityEmergency},
		{600, SeverityEmergency},
	}

	for _, tt := range tests {
		result := manager.determineSeverity(tt.value, thresholds)
		if result != tt.expected {
			t.Errorf("value %.0f: expected %s, got %s",
				tt.value, tt.expected, result)
		}
	}
}

func TestGetActiveAlerts(t *testing.T) {
	manager := NewManager()

	// Create an alert manually
	manager.mu.Lock()
	manager.alerts["alert-1"] = &Alert{
		ID:     "alert-1",
		Status: AlertStatusActive,
	}
	manager.alerts["alert-2"] = &Alert{
		ID:     "alert-2",
		Status: AlertStatusResolved,
	}
	manager.mu.Unlock()

	active := manager.GetActiveAlerts()

	if len(active) != 1 {
		t.Errorf("expected 1 active alert, got %d", len(active))
	}
}

func TestAcknowledgeAlert(t *testing.T) {
	manager := NewManager()

	manager.mu.Lock()
	manager.alerts["alert-1"] = &Alert{
		ID:     "alert-1",
		Status: AlertStatusActive,
	}
	manager.mu.Unlock()

	err := manager.AcknowledgeAlert("alert-1")
	if err != nil {
		t.Fatalf("failed to acknowledge: %v", err)
	}

	alert := manager.alerts["alert-1"]
	if alert.Status != AlertStatusAcknowledged {
		t.Errorf("expected status 'acknowledged', got %s", alert.Status)
	}
}

func TestResolveAlert(t *testing.T) {
	manager := NewManager()

	manager.mu.Lock()
	manager.alerts["alert-1"] = &Alert{
		ID:     "alert-1",
		Status: AlertStatusActive,
	}
	manager.mu.Unlock()

	err := manager.ResolveAlert("alert-1", "Resolved manually")
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	alert := manager.alerts["alert-1"]
	if alert.Status != AlertStatusResolved {
		t.Errorf("expected status 'resolved', got %s", alert.Status)
	}

	if alert.ResolvedAt == nil {
		t.Error("expected resolved time to be set")
	}
}

func TestGetHistory(t *testing.T) {
	manager := NewManager()

	// Add some history
	manager.mu.Lock()
	manager.history = []AlertEvent{
		{Timestamp: time.Now(), AlertID: "alert-1", Type: EventTriggered},
		{Timestamp: time.Now(), AlertID: "alert-1", Type: EventResolved},
		{Timestamp: time.Now(), AlertID: "alert-2", Type: EventTriggered},
	}
	manager.mu.Unlock()

	history := manager.GetHistory(2)

	if len(history) != 2 {
		t.Errorf("expected 2 events, got %d", len(history))
	}
}

func BenchmarkCreateRule(b *testing.B) {
	manager := NewManager()
	config := RuleConfig{Name: "Bench"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CreateRule(config)
	}
}
