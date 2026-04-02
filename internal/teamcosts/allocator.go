// Package teamcosts provides team cost allocation capabilities for TokMan
package teamcosts

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Allocator manages team cost allocation
type Allocator struct {
	teams       map[string]*Team
	allocations map[string]*Allocation
	rules       []AllocationRule
	mu          sync.RWMutex
}

// Team represents a team with cost tracking
type Team struct {
	ID          string
	Name        string
	Description string
	Members     []string
	Budget      Budget
	Costs       CostSummary
	Metadata    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Budget defines team budget
type Budget struct {
	MonthlyLimit   float64
	YearlyLimit    float64
	AlertThreshold float64
	CurrentSpend   float64
}

// CostSummary holds aggregated cost information
type CostSummary struct {
	TotalSpend       float64
	ThisMonth        float64
	ThisYear         float64
	ProjectedMonthly float64
	ProjectedYearly  float64
	LastMonth        float64
	Trend            float64
	Breakdown        CostBreakdown
}

// CostBreakdown breaks down costs by category
type CostBreakdown struct {
	ByService map[string]float64
	ByModel   map[string]float64
	ByFeature map[string]float64
	ByDay     map[string]float64
}

// Allocation represents a cost allocation entry
type Allocation struct {
	ID          string
	TeamID      string
	Amount      float64
	Currency    string
	Period      AllocationPeriod
	Source      CostSource
	Description string
	Tags        map[string]string
	CreatedAt   time.Time
}

// AllocationPeriod defines the time period
type AllocationPeriod struct {
	Start time.Time
	End   time.Time
}

// CostSource identifies where costs originated
type CostSource struct {
	Type     string
	ID       string
	Name     string
	Category string
}

// AllocationRule defines automatic allocation rules
type AllocationRule struct {
	ID         string
	Name       string
	Priority   int
	Condition  RuleCondition
	Allocation RuleAllocation
	Enabled    bool
}

// RuleCondition defines when to apply a rule
type RuleCondition struct {
	SourceType string
	SourceID   string
	Tags       map[string]string
}

// RuleAllocation defines how to allocate costs
type RuleAllocation struct {
	TeamID      string
	Percentage  float64
	FixedAmount *float64
}

// Report generates cost reports
type Report struct {
	Period      AllocationPeriod
	GeneratedAt time.Time
	Teams       []TeamReport
	Summary     ReportSummary
	Trends      Trends
}

// TeamReport holds a team's cost report
type TeamReport struct {
	TeamID      string
	TeamName    string
	Budget      Budget
	ActualSpend float64
	Variance    float64
	VariancePct float64
	Breakdown   CostBreakdown
	TopCosts    []CostItem
}

// ReportSummary provides aggregate report data
type ReportSummary struct {
	TotalBudget      float64
	TotalSpend       float64
	TotalVariance    float64
	UtilizationRate  float64
	TeamCount        int
	OverBudgetTeams  int
	UnderBudgetTeams int
}

// Trends captures spending trends
type Trends struct {
	DailyGrowthRate   float64
	WeeklyGrowthRate  float64
	MonthlyGrowthRate float64
	SeasonalFactor    float64
}

// CostItem represents a single cost item
type CostItem struct {
	Source     CostSource
	Amount     float64
	Percentage float64
	Date       time.Time
}

// NewAllocator creates a team cost allocator
func NewAllocator() *Allocator {
	return &Allocator{
		teams:       make(map[string]*Team),
		allocations: make(map[string]*Allocation),
		rules:       make([]AllocationRule, 0),
	}
}

// CreateTeam creates a new team
func (a *Allocator) CreateTeam(name, description string, budget Budget) (*Team, error) {
	team := &Team{
		ID:          generateTeamID(),
		Name:        name,
		Description: description,
		Members:     make([]string, 0),
		Budget:      budget,
		Costs: CostSummary{
			Breakdown: CostBreakdown{
				ByService: make(map[string]float64),
				ByModel:   make(map[string]float64),
				ByFeature: make(map[string]float64),
				ByDay:     make(map[string]float64),
			},
		},
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	a.mu.Lock()
	a.teams[team.ID] = team
	a.mu.Unlock()

	return team, nil
}

// GetTeam returns a team by ID
func (a *Allocator) GetTeam(id string) (*Team, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	team, exists := a.teams[id]
	if !exists {
		return nil, fmt.Errorf("team %s not found", id)
	}

	return team, nil
}

// ListTeams returns all teams
func (a *Allocator) ListTeams() []*Team {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]*Team, 0, len(a.teams))
	for _, team := range a.teams {
		result = append(result, team)
	}

	return result
}

// UpdateTeam updates a team
func (a *Allocator) UpdateTeam(id string, updates TeamUpdates) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	team, exists := a.teams[id]
	if !exists {
		return fmt.Errorf("team %s not found", id)
	}

	if updates.Name != "" {
		team.Name = updates.Name
	}
	if updates.Description != "" {
		team.Description = updates.Description
	}
	if updates.Budget.MonthlyLimit > 0 {
		team.Budget.MonthlyLimit = updates.Budget.MonthlyLimit
	}

	team.UpdatedAt = time.Now()
	return nil
}

// DeleteTeam deletes a team
func (a *Allocator) DeleteTeam(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.teams[id]; !exists {
		return fmt.Errorf("team %s not found", id)
	}

	delete(a.teams, id)
	return nil
}

// AddMember adds a member to a team
func (a *Allocator) AddMember(teamID, userID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	team, exists := a.teams[teamID]
	if !exists {
		return fmt.Errorf("team %s not found", teamID)
	}

	// Check if already a member
	for _, member := range team.Members {
		if member == userID {
			return nil
		}
	}

	team.Members = append(team.Members, userID)
	team.UpdatedAt = time.Now()

	return nil
}

// RemoveMember removes a member from a team
func (a *Allocator) RemoveMember(teamID, userID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	team, exists := a.teams[teamID]
	if !exists {
		return fmt.Errorf("team %s not found", teamID)
	}

	members := make([]string, 0)
	for _, member := range team.Members {
		if member != userID {
			members = append(members, member)
		}
	}

	team.Members = members
	team.UpdatedAt = time.Now()

	return nil
}

// AllocateCost allocates a cost to a team
func (a *Allocator) AllocateCost(teamID string, amount float64, source CostSource, description string, tags map[string]string) (*Allocation, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	team, exists := a.teams[teamID]
	if !exists {
		return nil, fmt.Errorf("team %s not found", teamID)
	}

	allocation := &Allocation{
		ID:          generateAllocationID(),
		TeamID:      teamID,
		Amount:      amount,
		Currency:    "USD",
		Period:      AllocationPeriod{Start: time.Now(), End: time.Now()},
		Source:      source,
		Description: description,
		Tags:        tags,
		CreatedAt:   time.Now(),
	}

	a.allocations[allocation.ID] = allocation

	// Update team costs
	team.Costs.TotalSpend += amount
	team.Budget.CurrentSpend += amount

	// Update breakdowns
	team.Costs.Breakdown.ByService[source.Name] += amount
	if category, ok := tags["model"]; ok {
		team.Costs.Breakdown.ByModel[category] += amount
	}
	if feature, ok := tags["feature"]; ok {
		team.Costs.Breakdown.ByFeature[feature] += amount
	}
	day := time.Now().Format("2006-01-02")
	team.Costs.Breakdown.ByDay[day] += amount

	team.UpdatedAt = time.Now()

	return allocation, nil
}

// AddAllocationRule adds an automatic allocation rule
func (a *Allocator) AddAllocationRule(rule AllocationRule) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	rule.ID = generateRuleID()
	a.rules = append(a.rules, rule)

	// Sort by priority
	sort.Slice(a.rules, func(i, j int) bool {
		return a.rules[i].Priority < a.rules[j].Priority
	})

	return nil
}

// ApplyAllocationRules applies automatic allocation rules to a cost
func (a *Allocator) ApplyAllocationRules(source CostSource, amount float64, tags map[string]string) ([]*Allocation, error) {
	a.mu.RLock()
	rules := make([]AllocationRule, len(a.rules))
	copy(rules, a.rules)
	a.mu.RUnlock()

	allocations := make([]*Allocation, 0)

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check if rule matches
		if a.ruleMatches(rule, source, tags) {
			allocAmount := amount * (rule.Allocation.Percentage / 100.0)
			if rule.Allocation.FixedAmount != nil {
				allocAmount = *rule.Allocation.FixedAmount
			}

			alloc, err := a.AllocateCost(
				rule.Allocation.TeamID,
				allocAmount,
				source,
				fmt.Sprintf("Auto-allocated by rule: %s", rule.Name),
				tags,
			)
			if err != nil {
				continue
			}

			allocations = append(allocations, alloc)
		}
	}

	return allocations, nil
}

func (a *Allocator) ruleMatches(rule AllocationRule, source CostSource, tags map[string]string) bool {
	if rule.Condition.SourceType != "" && rule.Condition.SourceType != source.Type {
		return false
	}
	if rule.Condition.SourceID != "" && rule.Condition.SourceID != source.ID {
		return false
	}

	for key, value := range rule.Condition.Tags {
		if tags[key] != value {
			return false
		}
	}

	return true
}

// GenerateReport generates a cost report
func (a *Allocator) GenerateReport(period AllocationPeriod) *Report {
	a.mu.RLock()
	defer a.mu.RUnlock()

	report := &Report{
		Period:      period,
		GeneratedAt: time.Now(),
		Teams:       make([]TeamReport, 0, len(a.teams)),
	}

	var totalBudget, totalSpend float64

	for _, team := range a.teams {
		// Calculate costs for period
		var periodSpend float64
		for day, amount := range team.Costs.Breakdown.ByDay {
			dayTime, _ := time.Parse("2006-01-02", day)
			if (dayTime.Equal(period.Start) || dayTime.After(period.Start)) &&
				(dayTime.Equal(period.End) || dayTime.Before(period.End)) {
				periodSpend += amount
			}
		}

		variance := team.Budget.MonthlyLimit - periodSpend
		variancePct := 0.0
		if team.Budget.MonthlyLimit > 0 {
			variancePct = (variance / team.Budget.MonthlyLimit) * 100
		}

		// Get top costs
		topCosts := a.getTopCosts(team, 5)

		teamReport := TeamReport{
			TeamID:      team.ID,
			TeamName:    team.Name,
			Budget:      team.Budget,
			ActualSpend: periodSpend,
			Variance:    variance,
			VariancePct: variancePct,
			Breakdown:   team.Costs.Breakdown,
			TopCosts:    topCosts,
		}

		report.Teams = append(report.Teams, teamReport)

		totalBudget += team.Budget.MonthlyLimit
		totalSpend += periodSpend
	}

	// Generate summary
	overBudget := 0
	underBudget := 0
	for _, tr := range report.Teams {
		if tr.Variance < 0 {
			overBudget++
		} else {
			underBudget++
		}
	}

	report.Summary = ReportSummary{
		TotalBudget:      totalBudget,
		TotalSpend:       totalSpend,
		TotalVariance:    totalBudget - totalSpend,
		UtilizationRate:  (totalSpend / totalBudget) * 100,
		TeamCount:        len(a.teams),
		OverBudgetTeams:  overBudget,
		UnderBudgetTeams: underBudget,
	}

	// Calculate trends (simplified)
	report.Trends = Trends{
		DailyGrowthRate:   0.0,
		WeeklyGrowthRate:  0.0,
		MonthlyGrowthRate: 0.0,
	}

	return report
}

func (a *Allocator) getTopCosts(team *Team, n int) []CostItem {
	items := make([]CostItem, 0)

	for name, amount := range team.Costs.Breakdown.ByService {
		items = append(items, CostItem{
			Source: CostSource{Name: name},
			Amount: amount,
		})
	}

	// Sort by amount
	sort.Slice(items, func(i, j int) bool {
		return items[i].Amount > items[j].Amount
	})

	if len(items) > n {
		items = items[:n]
	}

	// Calculate percentages
	total := 0.0
	for _, item := range items {
		total += item.Amount
	}

	for i := range items {
		if total > 0 {
			items[i].Percentage = (items[i].Amount / total) * 100
		}
	}

	return items
}

// GetTeamCosts returns cost summary for a team
func (a *Allocator) GetTeamCosts(teamID string) (*CostSummary, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	team, exists := a.teams[teamID]
	if !exists {
		return nil, fmt.Errorf("team %s not found", teamID)
	}

	return &team.Costs, nil
}

// SetBudget sets a team's budget
func (a *Allocator) SetBudget(teamID string, budget Budget) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	team, exists := a.teams[teamID]
	if !exists {
		return fmt.Errorf("team %s not found", teamID)
	}

	team.Budget = budget
	team.UpdatedAt = time.Now()

	return nil
}

// TeamUpdates holds team update fields
type TeamUpdates struct {
	Name        string
	Description string
	Budget      Budget
}

func generateTeamID() string {
	return fmt.Sprintf("team-%d", time.Now().UnixNano())
}

func generateAllocationID() string {
	return fmt.Sprintf("alloc-%d", time.Now().UnixNano())
}

func generateRuleID() string {
	return fmt.Sprintf("rule-%d", time.Now().Unix())
}
