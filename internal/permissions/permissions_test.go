package permissions

import "testing"

func TestNewPermissionEngine(t *testing.T) {
	e := NewPermissionEngine()
	if e == nil {
		t.Fatal("NewPermissionEngine() returned nil")
	}
}

func TestGrantAndCheck(t *testing.T) {
	e := NewPermissionEngine()
	e.Grant("admin", "files", "read")
	if !e.Check("admin", "files", "read") {
		t.Error("should have read permission for admin")
	}
	if e.Check("guest", "files", "read") {
		t.Error("guest should not have read permission")
	}
}

func TestListRules(t *testing.T) {
	e := NewPermissionEngine()
	e.Grant("admin", "files", "read")
	e.Grant("admin", "files", "write")
	rules := e.ListRules("admin")
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestCheckDenyByDefault(t *testing.T) {
	e := NewPermissionEngine()
	if e.Check("anyone", "secret", "read") {
		t.Error("should deny by default")
	}
}

func TestMultipleRoles(t *testing.T) {
	e := NewPermissionEngine()
	e.Grant("user", "data", "read")
	e.Grant("editor", "data", "write")

	if !e.Check("user", "data", "read") {
		t.Error("user should have read")
	}
	if e.Check("user", "data", "write") {
		t.Error("user should not have write")
	}
}
