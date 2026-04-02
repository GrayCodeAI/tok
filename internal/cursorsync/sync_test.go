package cursorsync

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestCursorSync(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Skip("SQLite not available")
	}
	defer db.Close()

	cs := NewCursorSync(db)
	if err := cs.Init(); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	cs.AddAccount(&CursorAccount{ID: "a1", Name: "Account 1", Email: "a1@example.com", Active: false})
	cs.AddAccount(&CursorAccount{ID: "a2", Name: "Account 2", Email: "a2@example.com", Active: true})

	active := cs.GetActiveAccount()
	if active == nil {
		t.Fatal("Expected active account")
	}
	if active.ID != "a2" {
		t.Errorf("Expected a2, got %s", active.ID)
	}

	status := cs.Status()
	if status == nil {
		t.Error("Expected status")
	}
}

func TestCursorSyncSwitch(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()

	cs := NewCursorSync(db)
	cs.Init()

	cs.AddAccount(&CursorAccount{ID: "a1", Name: "Account 1", Email: "a1@example.com"})
	cs.AddAccount(&CursorAccount{ID: "a2", Name: "Account 2", Email: "a2@example.com"})

	cs.SwitchAccount("a1")
	active := cs.GetActiveAccount()
	if active.ID != "a1" {
		t.Errorf("Expected a1, got %s", active.ID)
	}
}
