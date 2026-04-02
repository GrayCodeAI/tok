package teamcosts

import (
	"fmt"
	"testing"
	"time"
)

func TestNewAllocator(t *testing.T) {
	allocator := NewAllocator()
	if allocator == nil {
		t.Fatal("expected allocator to be created")
	}

	if allocator.teams == nil {
		t.Error("expected teams map to be initialized")
	}
}

func TestCreateTeam(t *testing.T) {
	allocator := NewAllocator()

	budget := Budget{
		MonthlyLimit: 1000,
	}

	team, err := allocator.CreateTeam("Engineering", "Engineering team", budget)
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	if team == nil {
		t.Fatal("expected team to be created")
	}

	if team.Name != "Engineering" {
		t.Errorf("expected name 'Engineering', got %s", team.Name)
	}

	if team.Budget.MonthlyLimit != 1000 {
		t.Errorf("expected budget 1000, got %.2f", team.Budget.MonthlyLimit)
	}
}

func TestGetTeam(t *testing.T) {
	allocator := NewAllocator()

	budget := Budget{MonthlyLimit: 1000}
	team, _ := allocator.CreateTeam("Test", "Test team", budget)

	found, err := allocator.GetTeam(team.ID)
	if err != nil {
		t.Fatalf("failed to get team: %v", err)
	}

	if found.ID != team.ID {
		t.Errorf("expected team ID %s, got %s", team.ID, found.ID)
	}
}

func TestGetTeamNotFound(t *testing.T) {
	allocator := NewAllocator()

	_, err := allocator.GetTeam("non-existent")
	if err == nil {
		t.Error("expected error for non-existent team")
	}
}

func TestListTeams(t *testing.T) {
	allocator := NewAllocator()

	allocator.CreateTeam("Team1", "Team 1", Budget{})
	allocator.CreateTeam("Team2", "Team 2", Budget{})

	teams := allocator.ListTeams()

	if len(teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(teams))
	}
}

func TestUpdateTeam(t *testing.T) {
	allocator := NewAllocator()

	budget := Budget{MonthlyLimit: 1000}
	team, _ := allocator.CreateTeam("Original", "Original team", budget)

	updates := TeamUpdates{
		Name:        "Updated",
		Description: "Updated description",
		Budget:      Budget{MonthlyLimit: 2000},
	}

	err := allocator.UpdateTeam(team.ID, updates)
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	updated, _ := allocator.GetTeam(team.ID)
	if updated.Name != "Updated" {
		t.Errorf("expected name 'Updated', got %s", updated.Name)
	}

	if updated.Budget.MonthlyLimit != 2000 {
		t.Errorf("expected budget 2000, got %.2f", updated.Budget.MonthlyLimit)
	}
}

func TestDeleteTeam(t *testing.T) {
	allocator := NewAllocator()

	budget := Budget{MonthlyLimit: 1000}
	team, _ := allocator.CreateTeam("ToDelete", "To delete", budget)

	err := allocator.DeleteTeam(team.ID)
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	_, err = allocator.GetTeam(team.ID)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestAddMember(t *testing.T) {
	allocator := NewAllocator()

	budget := Budget{MonthlyLimit: 1000}
	team, _ := allocator.CreateTeam("Test", "Test team", budget)

	err := allocator.AddMember(team.ID, "user-123")
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	team, _ = allocator.GetTeam(team.ID)
	if len(team.Members) != 1 {
		t.Errorf("expected 1 member, got %d", len(team.Members))
	}

	// Adding same member again should not duplicate
	err = allocator.AddMember(team.ID, "user-123")
	if err != nil {
		t.Fatalf("failed to add member: %v", err)
	}

	team, _ = allocator.GetTeam(team.ID)
	if len(team.Members) != 1 {
		t.Errorf("expected still 1 member, got %d", len(team.Members))
	}
}

func TestRemoveMember(t *testing.T) {
	allocator := NewAllocator()

	budget := Budget{MonthlyLimit: 1000}
	team, _ := allocator.CreateTeam("Test", "Test team", budget)

	allocator.AddMember(team.ID, "user-123")
	allocator.AddMember(team.ID, "user-456")

	err := allocator.RemoveMember(team.ID, "user-123")
	if err != nil {
		t.Fatalf("failed to remove member: %v", err)
	}

	team, _ = allocator.GetTeam(team.ID)
	if len(team.Members) != 1 {
		t.Errorf("expected 1 member, got %d", len(team.Members))
	}
}

func TestAllocateCost(t *testing.T) {
	allocator := NewAllocator()

	budget := Budget{MonthlyLimit: 1000}
	team, _ := allocator.CreateTeam("Test", "Test team", budget)

	source := CostSource{
		Type:     "compute",
		ID:       "instance-1",
		Name:     "EC2 Instance",
		Category: "infrastructure",
	}

	allocation, err := allocator.AllocateCost(team.ID, 100.50, source, "Monthly cost", map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("failed to allocate: %v", err)
	}

	if allocation == nil {
		t.Fatal("expected allocation to be created")
	}

	if allocation.Amount != 100.50 {
		t.Errorf("expected amount 100.50, got %.2f", allocation.Amount)
	}

	// Check team costs updated
	team, _ = allocator.GetTeam(team.ID)
	if team.Costs.TotalSpend != 100.50 {
		t.Errorf("expected total spend 100.50, got %.2f", team.Costs.TotalSpend)
	}

	if team.Budget.CurrentSpend != 100.50 {
		t.Errorf("expected current spend 100.50, got %.2f", team.Budget.CurrentSpend)
	}
}

func TestGenerateReport(t *testing.T) {
	allocator := NewAllocator()

	// Create teams
	budget1 := Budget{MonthlyLimit: 1000}
	team1, _ := allocator.CreateTeam("Team1", "Team 1", budget1)

	budget2 := Budget{MonthlyLimit: 2000}
	team2, _ := allocator.CreateTeam("Team2", "Team 2", budget2)

	// Add costs
	source := CostSource{Type: "compute", Name: "EC2"}
	allocator.AllocateCost(team1.ID, 500, source, "Cost", nil)
	allocator.AllocateCost(team2.ID, 1500, source, "Cost", nil)

	// Verify costs were recorded
	t1, _ := allocator.GetTeam(team1.ID)
	if t1.Costs.TotalSpend != 500 {
		t.Fatalf("expected team1 total spend 500, got %.2f", t1.Costs.TotalSpend)
	}

	// Generate report with a wide period
	period := AllocationPeriod{
		Start: time.Now().Add(-365 * 24 * time.Hour),
		End:   time.Now().Add(24 * time.Hour),
	}

	report := allocator.GenerateReport(period)

	if report == nil {
		t.Fatal("expected report to be generated")
	}

	if len(report.Teams) != 2 {
		t.Errorf("expected 2 teams in report, got %d", len(report.Teams))
	}

	if report.Summary.TotalBudget != 3000 {
		t.Errorf("expected total budget 3000, got %.2f", report.Summary.TotalBudget)
	}

	// Check that we have some spend
	if report.Summary.TotalSpend == 0 {
		t.Errorf("expected non-zero total spend, got %.2f", report.Summary.TotalSpend)
	}
}

func BenchmarkCreateTeam(b *testing.B) {
	allocator := NewAllocator()
	budget := Budget{MonthlyLimit: 1000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allocator.CreateTeam(fmt.Sprintf("team-%d", i), "Test", budget)
	}
}
