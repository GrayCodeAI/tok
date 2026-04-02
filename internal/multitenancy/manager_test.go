package multitenancy

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestTenantManager(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skip("SQLite not available")
	}
	defer db.Close()

	tm := NewTenantManager(db)
	if err := tm.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	tm.Create(&Tenant{
		ID:    "t1",
		Name:  "Test Tenant",
		Plan:  "pro",
		Quota: 10000,
	})

	if !tm.IsWithinQuota("t1") {
		t.Error("Expected tenant to be within quota")
	}

	tm.RecordUsage("t1", 500)
	tenants := tm.List()
	if len(tenants) != 1 {
		t.Errorf("Expected 1 tenant, got %d", len(tenants))
	}
}

func TestRBACManager(t *testing.T) {
	m := NewRBACManager()

	m.AssignRole("user1", RoleAdmin)
	if !m.CheckPermission("user1", "anything", "any") {
		t.Error("Admin should have all permissions")
	}

	m.AssignRole("user2", RoleUser)
	if !m.CheckPermission("user2", "compression", "execute") {
		t.Error("User should have execute permission")
	}
	if m.CheckPermission("user2", "admin", "delete") {
		t.Error("User should not have admin delete permission")
	}
}

func TestFeatureFlagManager(t *testing.T) {
	m := NewFeatureFlagManager()

	m.Set("new_feature", true, "tenant1")
	if !m.IsEnabled("new_feature", "tenant1") {
		t.Error("Feature should be enabled for tenant1")
	}
	if m.IsEnabled("new_feature", "tenant2") {
		t.Error("Feature should not be enabled for tenant2")
	}
}
