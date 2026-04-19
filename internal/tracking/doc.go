// Package tracking provides command execution tracking and token usage analytics for tok.
//
// The tracking package records command execution metrics to a SQLite database,
// enabling token usage analysis, savings reports, and checkpoint events.
//
// # Tracker
//
// The Tracker type is the main entry point for recording command metrics:
//
//	tracker, err := tracking.NewTracker(dbPath)
//	if err != nil { ... }
//	defer tracker.Close()
//
//	tracker.Record(&tracking.CommandRecord{
//	    Command: "git status",
//	    OriginalTokens: 500,
//	    FilteredTokens: 100,
//	})
//
// # Timed Execution
//
// Use Start() to track command execution time:
//
//	timer := tracking.Start()
//	// ... execute command ...
//	timer.Track("git status", "tok git status", originalTokens, filteredTokens)
//
// # Global Tracker
//
// A global tracker instance is available for convenience:
//
//	tracker := tracking.GetGlobalTracker()
//	defer tracking.CloseGlobalTracker()
//
// # Checkpoint Events
//
// The tracker automatically records checkpoint events for milestone commands
// and session context reads, enabling rich analytics and session reconstruction.
package tracking
