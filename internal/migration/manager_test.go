package migration

import (
	"testing"
)

func TestNewMigrationManager(t *testing.T) {
	mgr := NewMigrationManager()
	if mgr == nil {
		t.Error("Expected non-nil manager")
	}
}

func TestMigrationManagerRegister(t *testing.T) {
	mgr := NewMigrationManager()
	mgr.Register(CreateV1Migration())

	if len(mgr.migrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(mgr.migrations))
	}
}

func TestMigrationManagerRunAll(t *testing.T) {
	mgr := NewMigrationManager()
	mgr.Register(CreateV1Migration())
	mgr.Register(CreateV2Migration())

	err := mgr.RunAll(nil)
	if err != nil {
		t.Errorf("RunAll failed: %v", err)
	}
}

func TestMigrationManagerRunUpTo(t *testing.T) {
	mgr := NewMigrationManager()
	mgr.Register(CreateV1Migration())
	mgr.Register(CreateV2Migration())

	err := mgr.RunUpTo(nil, 1)
	if err != nil {
		t.Errorf("RunUpTo failed: %v", err)
	}
}
