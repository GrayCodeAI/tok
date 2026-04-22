package shared

import (
	"context"
	"testing"
)

func TestWithAppState(t *testing.T) {
	ctx := context.Background()
	state := &AppState{
		CfgFile:     "/tmp/config",
		Verbose:     2,
		DryRun:      true,
		QueryIntent: "test intent",
	}

	ctx = WithAppState(ctx, state)

	retrieved := AppStateFrom(ctx)
	if retrieved == nil {
		t.Fatal("AppStateFrom returned nil")
	}
	if retrieved.CfgFile != state.CfgFile {
		t.Errorf("CfgFile = %q, want %q", retrieved.CfgFile, state.CfgFile)
	}
	if retrieved.Verbose != state.Verbose {
		t.Errorf("Verbose = %d, want %d", retrieved.Verbose, state.Verbose)
	}
	if !retrieved.DryRun {
		t.Error("DryRun should be true")
	}
	if retrieved.QueryIntent != state.QueryIntent {
		t.Errorf("QueryIntent = %q, want %q", retrieved.QueryIntent, state.QueryIntent)
	}
}

func TestAppStateFrom_FallbackToGlobal(t *testing.T) {
	// When no state is in context, should return global state
	ctx := context.Background()
	retrieved := AppStateFrom(ctx)
	if retrieved != globalState {
		t.Error("AppStateFrom should fallback to globalState when none in context")
	}
}

func TestAppStateFrom_NilContext(t *testing.T) {
	// Even with empty context, should not panic and return global state
	ctx := context.Background()
	retrieved := AppStateFrom(ctx)
	if retrieved == nil {
		t.Fatal("AppStateFrom returned nil for empty context")
	}
}

func TestAppStateFrom_DifferentKeysDoNotCollide(t *testing.T) {
	// Use a different key type to verify isolation
	type otherKey struct{}
	ctx := context.WithValue(context.Background(), otherKey{}, &AppState{Verbose: 99})

	// Should NOT pick up the wrong key type
	retrieved := AppStateFrom(ctx)
	if retrieved.Verbose == 99 {
		t.Error("AppStateFrom picked up state from wrong key type")
	}
}

func TestAppStateFrom_NestedContexts(t *testing.T) {
	outerState := &AppState{Verbose: 1, QueryIntent: "outer"}
	innerState := &AppState{Verbose: 2, QueryIntent: "inner"}

	outerCtx := WithAppState(context.Background(), outerState)
	innerCtx := WithAppState(outerCtx, innerState)

	// Inner context should return inner state
	if got := AppStateFrom(innerCtx); got != innerState {
		t.Error("inner context returned wrong state")
	}

	// Outer context should still return outer state
	if got := AppStateFrom(outerCtx); got != outerState {
		t.Error("outer context returned wrong state")
	}
}

func TestAppState_ConcurrentAccess(t *testing.T) {
	state := &AppState{}

	// Concurrent reads and writes should be safe
	done := make(chan bool, 4)

	for i := 0; i < 2; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				state.mu.Lock()
				state.Verbose = j
				state.mu.Unlock()
			}
			done <- true
		}()
	}

	for i := 0; i < 2; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				state.mu.RLock()
				_ = state.Verbose
				state.mu.RUnlock()
			}
			done <- true
		}()
	}

	for i := 0; i < 4; i++ {
		<-done
	}
}
