package tracking

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/GrayCodeAI/tokman/internal/session"
)

func TestGetWorkspaceDashboardSnapshot(t *testing.T) {
	tr := newTestTracker(t)
	project := filepath.Join(t.TempDir(), "project")
	now := time.Now()

	insertCommandForDashboard(t, tr, CommandRecord{
		Command:        "git status",
		OriginalTokens: 1200,
		FilteredTokens: 500,
		SavedTokens:    700,
		ProjectPath:    project,
		SessionID:      "sess-1",
		ParseSuccess:   true,
		AgentName:      "Claude Code",
		ModelName:      "claude-4-sonnet",
		Provider:       "Anthropic",
	}, now)

	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	manager, err := session.NewSessionManager()
	if err != nil {
		t.Fatalf("NewSessionManager(): %v", err)
	}
	defer manager.Close()

	sess, err := manager.CreateSession("Claude Code", project)
	if err != nil {
		t.Fatalf("CreateSession(): %v", err)
	}
	if err := manager.AddContextBlock(session.BlockTypeUserQuery, "status", 20); err != nil {
		t.Fatalf("AddContextBlock(): %v", err)
	}
	if _, err := manager.CreateSnapshot(sess.ID); err != nil {
		t.Fatalf("CreateSnapshot(): %v", err)
	}

	snapshot, err := GetWorkspaceDashboardSnapshot(
		tr,
		manager,
		DashboardQueryOptions{Days: 30, ProjectPath: project},
		session.SessionListOptions{Limit: 10},
		10,
	)
	if err != nil {
		t.Fatalf("GetWorkspaceDashboardSnapshot(): %v", err)
	}

	if snapshot.Dashboard == nil || snapshot.Sessions == nil {
		t.Fatal("expected dashboard and session payloads")
	}
	if snapshot.Dashboard.Overview.TotalCommands != 1 {
		t.Fatalf("Overview.TotalCommands = %d, want 1", snapshot.Dashboard.Overview.TotalCommands)
	}
	if snapshot.DataQuality.TotalCommands != 1 {
		t.Fatalf("DataQuality.TotalCommands = %d, want 1", snapshot.DataQuality.TotalCommands)
	}
	if snapshot.Sessions.StoreSummary.TotalSessions != 1 {
		t.Fatalf("StoreSummary.TotalSessions = %d, want 1", snapshot.Sessions.StoreSummary.TotalSessions)
	}
	if len(snapshot.Sessions.SnapshotHistory) != 1 {
		t.Fatalf("len(SnapshotHistory) = %d, want 1", len(snapshot.Sessions.SnapshotHistory))
	}
}
