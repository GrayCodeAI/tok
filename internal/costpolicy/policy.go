// Package costpolicy provides cost policy enforcement
package costpolicy

import (
	"fmt"
	"sync"
	"time"
)

// Policy represents a cost policy
type Policy struct {
	ID          string
	Name        string
	Description string
	Enabled     bool
	Rules       []PolicyRule
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PolicyRule represents a policy rule
type PolicyRule struct {
	ID          string
	Type        RuleType
	Condition   RuleCondition
	Action      RuleAction
	Priority    int
	Description string
}

// RuleType represents the type of rule
type RuleType string

const (
	RuleTypeBudgetLimit   RuleType = "budget_limit"
	RuleTypeModelRestrict RuleType = "model_restrict"
	RuleTypeRateLimit     RuleType = "rate_limit"
	RuleTypeApproval      RuleType = "approval_required"
	RuleTypeNotification  RuleType = "notify"
	RuleTypeAutoShutdown  RuleType = "auto_shutdown"
)

// RuleCondition represents a rule condition
type RuleCondition struct {
	Metric    string
	Operator  string
	Value     float64
	TimeRange string
	Team      string
	Model     string
}

// RuleAction represents a rule action
type RuleAction struct {
	Type       ActionType
	Message    string
	Recipients []string
	Limit      float64
	Shutdown   bool
}

// ActionType represents the type of action
type ActionType string

const (
	ActionTypeAllow    ActionType = "allow"
	ActionTypeDeny     ActionType = "deny"
	ActionTypeWarn     ActionType = "warn"
	ActionTypeNotify   ActionType = "notify"
	ActionTypeShutdown ActionType = "shutdown"
	ActionTypeLimit    ActionType = "limit"
)

// EnforcementEngine enforces cost policies
type EnforcementEngine struct {
	policies   map[string]*Policy
	mu         sync.RWMutex
	violations []PolicyViolation
}

// PolicyViolation represents a policy violation
type PolicyViolation struct {
	ID        string
	PolicyID  string
	RuleID    string
	User      string
	Team      string
	Action    string
	Cost      float64
	Timestamp time.Time
	Status    string
}

// NewEnforcementEngine creates a new enforcement engine
func NewEnforcementEngine() *EnforcementEngine {
	return &EnforcementEngine{
		policies:   make(map[string]*Policy),
		violations: make([]PolicyViolation, 0),
	}
}

// AddPolicy adds a policy
func (e *EnforcementEngine) AddPolicy(policy *Policy) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if policy.ID == "" {
		policy.ID = generateID()
	}

	e.policies[policy.ID] = policy
	return nil
}

// RemovePolicy removes a policy
func (e *EnforcementEngine) RemovePolicy(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.policies, id)
}

// CheckPolicy checks if an action violates any policy
func (e *EnforcementEngine) CheckPolicy(user, team, model string, cost float64) (*PolicyDecision, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	decision := &PolicyDecision{
		Allowed: true,
	}

	for _, policy := range e.policies {
		if !policy.Enabled {
			continue
		}

		for _, rule := range policy.Rules {
			if e.evaluateRule(rule, user, team, model, cost) {
				switch rule.Action.Type {
				case ActionTypeDeny:
					decision.Allowed = false
					decision.Reason = rule.Action.Message
					decision.PolicyID = policy.ID
					decision.RuleID = rule.ID

					e.recordViolation(policy.ID, rule.ID, user, team, model, cost)
					return decision, nil

				case ActionTypeWarn:
					decision.Warnings = append(decision.Warnings, rule.Action.Message)

				case ActionTypeNotify:
					decision.Notifications = append(decision.Notifications, Notification{
						Recipients: rule.Action.Recipients,
						Message:    rule.Action.Message,
					})

				case ActionTypeLimit:
					if cost > rule.Action.Limit {
						decision.Allowed = false
						decision.Reason = fmt.Sprintf("Cost %.2f exceeds limit %.2f", cost, rule.Action.Limit)
						decision.PolicyID = policy.ID
						decision.RuleID = rule.ID

						e.recordViolation(policy.ID, rule.ID, user, team, model, cost)
						return decision, nil
					}

				case ActionTypeShutdown:
					decision.ShouldShutdown = true
					decision.Reason = rule.Action.Message
				}
			}
		}
	}

	return decision, nil
}

func (e *EnforcementEngine) evaluateRule(rule PolicyRule, user, team, model string, cost float64) bool {
	cond := rule.Condition

	// Check team filter
	if cond.Team != "" && cond.Team != team {
		return false
	}

	// Check model filter
	if cond.Model != "" && cond.Model != model {
		return false
	}

	// Evaluate condition
	switch cond.Operator {
	case ">":
		return cost > cond.Value
	case ">=":
		return cost >= cond.Value
	case "<":
		return cost < cond.Value
	case "<=":
		return cost <= cond.Value
	case "==":
		return cost == cond.Value
	default:
		return false
	}
}

func (e *EnforcementEngine) recordViolation(policyID, ruleID, user, team, model string, cost float64) {
	violation := PolicyViolation{
		ID:        generateID(),
		PolicyID:  policyID,
		RuleID:    ruleID,
		User:      user,
		Team:      team,
		Action:    model,
		Cost:      cost,
		Timestamp: time.Now(),
		Status:    "open",
	}

	e.violations = append(e.violations, violation)
}

// GetViolations returns policy violations
func (e *EnforcementEngine) GetViolations(filter ViolationFilter) []PolicyViolation {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]PolicyViolation, 0)

	for _, v := range e.violations {
		if filter.Matches(v) {
			result = append(result, v)
		}
	}

	return result
}

// ViolationFilter filters violations
type ViolationFilter struct {
	PolicyID string
	User     string
	Team     string
	Status   string
	FromDate time.Time
	ToDate   time.Time
}

// Matches checks if a violation matches the filter
func (f ViolationFilter) Matches(v PolicyViolation) bool {
	if f.PolicyID != "" && v.PolicyID != f.PolicyID {
		return false
	}
	if f.User != "" && v.User != f.User {
		return false
	}
	if f.Team != "" && v.Team != f.Team {
		return false
	}
	if f.Status != "" && v.Status != f.Status {
		return false
	}
	if !f.FromDate.IsZero() && v.Timestamp.Before(f.FromDate) {
		return false
	}
	if !f.ToDate.IsZero() && v.Timestamp.After(f.ToDate) {
		return false
	}
	return true
}

// PolicyDecision represents a policy decision
type PolicyDecision struct {
	Allowed        bool
	Reason         string
	PolicyID       string
	RuleID         string
	ShouldShutdown bool
	Warnings       []string
	Notifications  []Notification
}

// Notification represents a notification
type Notification struct {
	Recipients []string
	Message    string
}

func generateID() string {
	return fmt.Sprintf("policy-%d", time.Now().UnixNano())
}

// StandardPolicies returns standard cost policies
func StandardPolicies() []*Policy {
	return []*Policy{
		{
			Name:        "Budget Limit",
			Description: "Enforce budget limits per team",
			Enabled:     true,
			Rules: []PolicyRule{
				{
					Type: RuleTypeBudgetLimit,
					Condition: RuleCondition{
						Metric:   "monthly_spend",
						Operator: ">=",
						Value:    10000,
					},
					Action: RuleAction{
						Type:    ActionTypeDeny,
						Message: "Monthly budget limit exceeded",
					},
					Priority: 1,
				},
			},
		},
		{
			Name:        "Model Restrictions",
			Description: "Restrict expensive models",
			Enabled:     true,
			Rules: []PolicyRule{
				{
					Type: RuleTypeModelRestrict,
					Condition: RuleCondition{
						Model: "gpt-4-turbo",
					},
					Action: RuleAction{
						Type:       ActionTypeNotify,
						Message:    "Using expensive model gpt-4-turbo",
						Recipients: []string{"admin@company.com"},
					},
					Priority: 2,
				},
			},
		},
		{
			Name:        "Cost Approval",
			Description: "Require approval for high-cost requests",
			Enabled:     false,
			Rules: []PolicyRule{
				{
					Type: RuleTypeApproval,
					Condition: RuleCondition{
						Metric:   "request_cost",
						Operator: ">",
						Value:    100,
					},
					Action: RuleAction{
						Type:    ActionTypeDeny,
						Message: "Approval required for costs > $100",
					},
					Priority: 3,
				},
			},
		},
	}
}
